package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	cfg "github.com/felix-001/qnHackathon/internal/config"
	"github.com/rs/zerolog/log"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type GitLabMgr struct {
	Client     *gitlab.Client
	Conf       cfg.GitlabConf
	Branch     *gitlab.Branch
	OldTarName string
	MrUrl      string
}

func NewGitLabClient(Conf cfg.GitlabConf) *gitlab.Client {
	client, err := gitlab.NewClient(Conf.PrivateToken, gitlab.WithBaseURL(Conf.GitLabURL))
	if err != nil {
		log.Logger.Error().Msgf("NewGitLabClient Failed to create GitLab client: %v", err)
		return nil
	}
	return client
}

func NewGitLabMgr(Conf cfg.GitlabConf) *GitLabMgr {
	client := NewGitLabClient(Conf)
	if client == nil {
		log.Logger.Error().Msg("新建GitLabMgr失败")
		return nil
	}
	return &GitLabMgr{
		Client: client,
		Conf:   Conf,
	}
}

func (s *GitLabMgr) CreateBranch(module string) string {
	// module + 日期
	branch := module + "_" + time.Now().Local().Format("06_01_02_15_04_05")
	// 创建一个新的分支
	Main := "master"
	gitbranch, resp, err := s.Client.Branches.CreateBranch(s.Conf.ProjectID, &gitlab.CreateBranchOptions{
		Branch: &branch,
		Ref:    &Main, // 从 main 分支创建新分支
	})
	if err != nil {
		log.Logger.Error().Msgf("CreateBranch Failed to create branch: %v, branch: %s", err, branch)
		return ""
	}
	if resp.StatusCode != http.StatusCreated {
		log.Logger.Error().Msgf("CreateBranch Unexpected response status code: %d, branch: %s", resp.StatusCode, branch)
		return ""
	}
	s.Branch = gitbranch
	log.Logger.Info().Msgf("Branch Details: Name: %s, Commit ID: %s, %s", gitbranch.Name, gitbranch.Commit.ID, gitbranch.WebURL)

	log.Logger.Info().Msgf("Branch '%s' created successfully.", branch)
	return branch
}

func (s *GitLabMgr) CommitPush(branch string, title, message string) bool {
	master := "master"
	options := &gitlab.CreateMergeRequestOptions{
		SourceBranch: &branch,
		TargetBranch: &master,
		Title:        &title,
		Description:  &message,
	}
	mr, _, err := s.Client.MergeRequests.CreateMergeRequest(s.Conf.ProjectID, options)
	if err != nil {
		log.Logger.Error().Msgf("Failed to create merge request: %v", err)
		return false
	}

	s.MrUrl = mr.WebURL
	return true
}

func GetNodeFromUser(client *gitlab.Client, projectID string, mrid int) string {
	// 获取 Merge Request 的评论
	notes, resp, err := client.Notes.ListMergeRequestNotes(projectID, mrid, &gitlab.ListMergeRequestNotesOptions{}, gitlab.WithContext(context.Background()))
	if err != nil {
		log.Logger.Warn().Err(err).
			Str("projectID", projectID).
			Int("mrIID", mrid).
			Msgf("Failed to get Merge Request notes: %v", err)
		return ""
	}
	if resp.StatusCode != 200 {
		log.Logger.Warn().Err(err).
			Str("projectID", projectID).
			Int("mrIID", mrid).
			Msgf("Unexpected response status code: %d", resp.StatusCode)
		return ""
	}

	// 写qiniu-bot的内容到mr
	for _, note := range notes {
		if note.Author.Username != "qiniu-bot" {
			continue
		}
		if strings.Contains(note.Body, "issue:") {
			// 提取 issue 链接
			parts := strings.Split(note.Body, " ")
			if len(parts) == 4 {
				note := parts[1]
				if strings.HasPrefix(note, "https://") {
					return note
				}
			}
		}
	}
	return ""
}

func (s *GitLabMgr) GetMergeRequest() map[string][]*gitlab.BasicMergeRequest {
	// 获取项目信息
	project, _, err := s.Client.Projects.GetProject(s.Conf.ProjectID, nil)
	if err != nil {
		log.Logger.Error().Msgf("GetMergeRequest Failed to get project: %v\n", err)
		return nil
	}
	log.Logger.Info().Msgf("GetMergeRequest Name: %s, URL: %s, ID: %d\n", project.Name, project.WebURL, project.ID)

	// 获取项目的所有 Merge Requests
	page := 1
	perPage := 100 // 每页获取100条记录，最大值为100
	mergeRequests, _, err := s.Client.MergeRequests.ListProjectMergeRequests(s.Conf.ProjectID, &gitlab.ListProjectMergeRequestsOptions{
		ListOptions: gitlab.ListOptions{
			Page:    page,
			PerPage: perPage,
		},
	})
	if err != nil {
		log.Printf("Failed to get merge requests: %v\n", err)
		return nil
	}
	log.Logger.Info().Msgf("There are %d merge requests in total.\n", len(mergeRequests))

	// 创建一个 map 来存储分类后的 Merge Requests
	mergeRequestMap := make(map[string][]*gitlab.BasicMergeRequest)
	// 遍历 Merge Requests 并分类
	for _, mr := range mergeRequests {
		// 使用状态作为 key
		girid, _, err := ParseTitle(mr.Title)
		if err != nil {
			//log.Logger.Info().Msgf("Failed to get key from merge request title: %v\n", err)
			continue
		}
		key := fmt.Sprintf("%s-%s", mr.State, girid)

		mr.Labels = append(mr.Labels, GetNodeFromUser(s.Client, s.Conf.ProjectID, mr.IID))

		// diff
		_, res, err := s.Client.MergeRequests.GetMergeRequestChanges(s.Conf.ProjectID, mr.IID, nil)
		if err != nil {
			// 处理错误，例如记录日志
			log.Logger.Info().Msgf("Failed to get merge request diffs: %v\n", err)
			// 根据需要决定是否继续执行或跳过当前合并请求
			continue
		}

		// 检查 HTTP 响应状态码
		if res.StatusCode != http.StatusOK {
			log.Logger.Info().Msgf("Failed to get merge request diffs: %v\n", res)
			continue
		}

		log.Logger.Info().
			Str("key", girid).
			Str("mrIID", fmt.Sprintf("%d", mr.IID)).
			Str("id", fmt.Sprintf("%d", mr.ID)).
			Str("title", mr.Title).
			Str("webURL", mr.WebURL).
			Str("author", mr.Author.Name).
			Str("state", mr.State).
			Str("createdAt", mr.CreatedAt.String()).
			Str("description", mr.Description).
			Str("labels", strings.Join(mr.Labels, ",")).
			Msgf("DetailedMergeStatus requests: %+v\n", mr.DetailedMergeStatus)

		mergeRequestMap[key] = append(mergeRequestMap[key], mr)
	}

	return mergeRequestMap
}

func (s *GitLabMgr) GetFile(branch, filename string) {
	master := "master"

	file, resp, err := s.Client.RepositoryFiles.GetFile(
		s.Conf.ProjectID,
		filename,
		&gitlab.GetFileOptions{
			Ref: &master,
		})
	if err != nil {
		log.Logger.Err(err).Msgf("GetFile failed, branch: %s, %s, %s, %s", branch, filename)
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Logger.Error().Msgf("GetFile resp.Status :%d", resp.StatusCode)
		return
	}

	// Base64 解码文件内容
	decodedContent, err := base64.StdEncoding.DecodeString(file.Content)
	if err != nil {
		log.Logger.Error().Msgf("Failed to decode base64 content: %v", err)
		return
	}
	log.Info().Msgf("GetFile content: %s", string(decodedContent))
}

func (s *GitLabMgr) UpdateVersion(filename, newVersion string) {
	branch := s.CreateBranch("streamd")
	updatedContent := fmt.Sprintf(`{"version": "%s"}`, newVersion)
	// 提交更改
	_, _, err := s.Client.RepositoryFiles.UpdateFile(s.Conf.ProjectID,
		filename,
		&gitlab.UpdateFileOptions{
			Branch:        &branch,
			CommitMessage: &newVersion,
			Content:       &updatedContent,
		})
	if err != nil {
		log.Logger.Error().Msgf("Failed to update file: %v", err)
	}

	s.CommitPush(branch, newVersion, newVersion)
}

func (s *GitLabMgr) GetMrUrl(expectedTitle string) string {
	// 获取所有 MR 分类 map
	mrMap := s.GetMergeRequest()
	if mrMap == nil {
		log.Logger.Warn().Msg("GetMrUrl: 无法获取 MergeRequest 列表")
		return ""
	}

	// 遍历 map 中的所有 MR 列表
	for _, mrList := range mrMap {
		for _, mr := range mrList {
			if mr.Title == expectedTitle {
				return mr.WebURL
			}
		}
	}

	return ""
}

func (s *GitLabMgr) CreateOrUpdateFile(filePath, content, branch, commitMessage string) error {
	master := "master"
	
	if branch == "" {
		branch = s.CreateBranch("config")
		if branch == "" {
			return fmt.Errorf("failed to create branch")
		}
	}

	_, resp, err := s.Client.RepositoryFiles.GetFile(
		s.Conf.ProjectID,
		filePath,
		&gitlab.GetFileOptions{
			Ref: &master,
		})

	if err != nil && resp != nil && resp.StatusCode == http.StatusNotFound {
		_, _, err = s.Client.RepositoryFiles.CreateFile(s.Conf.ProjectID,
			filePath,
			&gitlab.CreateFileOptions{
				Branch:        &branch,
				CommitMessage: &commitMessage,
				Content:       &content,
			})
		if err != nil {
			log.Logger.Error().Msgf("Failed to create file: %v", err)
			return err
		}
	} else {
		_, _, err = s.Client.RepositoryFiles.UpdateFile(s.Conf.ProjectID,
			filePath,
			&gitlab.UpdateFileOptions{
				Branch:        &branch,
				CommitMessage: &commitMessage,
				Content:       &content,
			})
		if err != nil {
			log.Logger.Error().Msgf("Failed to update file: %v", err)
			return err
		}
	}

	return nil
}

func (s *GitLabMgr) CreateMergeRequest(sourceBranch, targetBranch, title, description string) (string, error) {
	if targetBranch == "" {
		targetBranch = "master"
	}

	options := &gitlab.CreateMergeRequestOptions{
		SourceBranch: &sourceBranch,
		TargetBranch: &targetBranch,
		Title:        &title,
		Description:  &description,
	}

	mr, _, err := s.Client.MergeRequests.CreateMergeRequest(s.Conf.ProjectID, options)
	if err != nil {
		log.Logger.Error().Msgf("Failed to create merge request: %v", err)
		return "", err
	}

	s.MrUrl = mr.WebURL
	return mr.WebURL, nil
}
