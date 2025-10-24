package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	cfg "github.com/felix-001/qnHackathon/internal/config"
	"github.com/google/go-github/v48/github"
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

func (s *GitLabMgr) UpdateCsvFile(branch, filename, newTarName, message string) bool {
	master := "master"

	file, resp, err := s.Client.RepositoryFiles.GetFile(
		s.Conf.ProjectID,
		filename,
		&gitlab.GetFileOptions{
			Ref: &master,
		})
	if err != nil {
		log.Logger.Err(err).Msgf("GetFile failed, branch: %s, %s, %s, %s", branch, filename, newTarName, message)
		return false
	}

	if resp.StatusCode != http.StatusOK {
		log.Logger.Error().Msgf("GetFile resp.Status :%d", resp.StatusCode)
		return false
	}

	// Base64 解码文件内容
	decodedContent, err := base64.StdEncoding.DecodeString(file.Content)
	if err != nil {
		log.Logger.Error().Msgf("Failed to decode base64 content: %v", err)
		return false
	}
	// 定义正则表达式匹配 MIKUD_LIVE 开头和 .tar.gz 结尾的文件名
	re := regexp.MustCompile(`MIKUD_LIVE.*\.tar\.gz`)
	OldTar := re.FindString(string(decodedContent))
	if OldTar != "" {
		s.OldTarName = OldTar
	}

	// 修改后的文件内容
	updatedContent := re.ReplaceAllString(string(decodedContent), newTarName)

	// 提交更改
	_, _, err = s.Client.RepositoryFiles.UpdateFile(s.Conf.ProjectID,
		filename,
		&gitlab.UpdateFileOptions{
			Branch:        &branch,
			CommitMessage: &message,
			Content:       &updatedContent,
		})
	if err != nil {
		log.Logger.Error().Msgf("Failed to update file: %v", err)
	}

	return true
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

func saveHTMLToFile(html string, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(html)
	if err != nil {
		return err
	}

	return nil
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
			Str("title", mr.Title).
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

func generateMergeRequestHTMLTable(mergeRequestMap map[string][]*gitlab.BasicMergeRequest, gitbubMap map[string]*github.PullRequest) string {
	html := `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>MIKU GitLab Merge Requests</title>
	<style>
		body {
			font-family: Arial, sans-serif;
		}
		table {
			width: 100%;
			border-collapse: collapse;
			margin-top: 20px;
		}
		th, td {
			border: 1px solid #ddd;
			padding: 12px;
			text-align: left;
		}
		th {
			background-color: #4CAF50;
			color: white;
			cursor: pointer;
		}
		tr:nth-child(even) {
			background-color: #f2f2f2;
		}
		tr:hover {
			background-color: #ddd;
		}
		.fixed-header {
			position: sticky;
			top: 0;
			z-index: 10;
		}
	</style>
</head>
<body>
	<h1>GitLab Merge Requests</h1>
`

	html += `
		<table>
			<thead class="fixed-header">
				<tr>
					<th onclick="sortTable(this, 0)">Title</th>
					<th onclick="sortTable(this, 1)">Author</th>
					<th onclick="sortTable(this, 2)">State</th>
					<th onclick="sortTable(this, 3)">Created At</th>
					<th onclick="sortTable(this, 4)">变更描述</th>
				</tr>
			</thead>
			<tbody>
		`
	// 添加每个状态的表格
	for key, mergeRequests := range mergeRequestMap {
		for _, mr := range mergeRequests {
			githubstr := "null"
			githubMR, ok := gitbubMap[key]
			if ok {
				githubstr = *githubMR.DiffURL
				// 去掉 .diff
				if strings.Contains(githubstr, ".diff") {
					githubstr = strings.Split(githubstr, ".diff")[0]
				}
			}

			// 检查 Jira 是否为链接
			jiraHTML := ""
			if len(mr.Labels) > 0 {
				if strings.HasPrefix(mr.Labels[0], "http://") || strings.HasPrefix(mr.Labels[0], "https://") {
					jiraHTML = fmt.Sprintf("<a href=\"%s\" target=\"_blank\">%s</a>", mr.Labels[0], mr.Labels[0])
				}
			}

			// 合并 Jira 和 GitLab 的信息
			floyHTML := fmt.Sprintf(
				"description: %s<br>jira: %s<br>floy:  %s<br>github:  %s<br>",
				mr.Description,
				jiraHTML, // Jira 链接
				fmt.Sprintf("<a href=\"%s\" target=\"_blank\">%s</a>", mr.WebURL, mr.WebURL), // GitLab 链接
				fmt.Sprintf("<a href=\"%s\" target=\"_blank\">%s</a>", githubstr, githubstr), // GitHub 链接，如果有的话
			)

			html += fmt.Sprintf(
				"<tr>"+
					"<td>%s</td>"+
					"<td>%s</td>"+
					"<td>%s</td>"+
					"<td>%s</td>"+
					"<td>%s</td>"+
					"</tr>",
				mr.Title,
				mr.Author.Name,
				mr.State,
				mr.CreatedAt.Format(time.RFC3339),
				floyHTML, // 合并后的 floy 列的内容
			)
		}

	}
	html += `
			</tbody>
		</table>
		`

	html += `
	<script>
		function sortTable(header, column) {
			const table = header.closest("table");
			const rows = Array.from(table.querySelectorAll("tbody tr"));
			const isAscending = header.dataset.sortOrder !== "asc";
			header.dataset.sortOrder = isAscending ? "asc" : "desc";

			rows.sort((a, b) => {
				const cellA = a.cells[column].textContent.trim();
				const cellB = b.cells[column].textContent.trim();

				if (isAscending) {
					return cellA.localeCompare(cellB);
				} else {
					return cellB.localeCompare(cellA);
				}
			});

			const tbody = table.querySelector("tbody");
			rows.forEach(row => tbody.appendChild(row));
		}
	</script>
</body>
</html>
`
	return html
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
		// “issue: https://jira.qiniu.io/browse/MIKU-1624 author: @liyuanquan”, 提取issue的链接
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

func GetModuleCvsFile(module string) string {
	return fmt.Sprintf("floy/miku-%s/miku-%s.csv", module, module)
}
