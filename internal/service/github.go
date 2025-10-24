package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	cfg "github.com/felix-001/qnHackathon/internal/config"
	"github.com/google/go-github/v48/github"
	"golang.org/x/oauth2"
)

type GitMergeRequest struct {
	GiraId        string
	GiraUrl       string //https://jira.qiniu.io/browse/-277
	Pr            *github.PullRequest
	GitId         string
	Title         string
	Author        string
	CreateAt      *time.Time
	MergeMessage  string // 本次合并的提交信息, 变更内容
	ChangeModules string // 变更模块
}

type GitHubMgr struct {
	Client *github.Client
	Conf   cfg.GitHubConf
}

func NewGitHubClient(ctx context.Context, token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	if client == nil {
		log.Logger.Info().Msgf("Failed to create GitHub client")
		return nil
	}
	return client
}

func NewGitHubMgr(conf cfg.GitHubConf) *GitHubMgr {
	client := NewGitHubClient(context.Background(), conf.GitHubToken)
	if client == nil {
		log.Logger.Info().Msgf("Failed to create GitHub client")
		return nil
	}
	return &GitHubMgr{
		Client: client,
		Conf:   conf}
}

func (s *GitHubMgr) GetGitHubPullRequest() map[string]*GitMergeRequest {
	// 获取仓库信息
	repository, _, err := s.Client.Repositories.Get(context.Background(), s.Conf.Owner, s.Conf.Repo)
	if err != nil {
		log.Logger.Info().Msgf("Failed to get repository: %v", err)
	}
	log.Printf("Repository URL: %s\n", *repository.URL)

	// 获取最近的 PR 记录
	opt := &github.PullRequestListOptions{
		State: "all", // 可选值：open, closed, all
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 30, // 根据需要调整每页显示的数量
		},
	}

	prs, resp, err := s.Client.PullRequests.List(context.Background(), s.Conf.Owner, s.Conf.Repo, opt)
	if err != nil {
		log.Logger.Info().Msgf("Failed to get PRs: %v", err)
	}
	defer resp.Body.Close()

	// 打印 PR 信息
	mrMap := make(map[string]*GitMergeRequest)
	for _, pr := range prs {
		IsMerged := pr.MergedAt != nil && pr.GetState() == "closed"
		if !IsMerged {
			continue
		}
		// 是否以开头
		if strings.Index(*pr.Title, "-") != 0 {
			continue
		}

		giraId, changeModules, err := ParseTitle(*pr.Title)
		if err != nil {
			log.Logger.Info().Msgf("Failed to parse title: %v", err)
		}

		giraurl := ""
		if giraId != "" && giraId != "0" {
			giraurl = fmt.Sprintf("https://jira.qiniu.io/browse/-%s", giraId)
		}

		mergeMessage := ""
		if pr.Body != nil {
			mergeMessage = *pr.Body
		}

		newPr := &GitMergeRequest{
			GiraId:        giraId,
			GiraUrl:       giraurl,
			Pr:            pr,
			GitId:         fmt.Sprint(*pr.Number),
			Title:         *pr.Title,
			Author:        *pr.GetUser().Login,
			CreateAt:      pr.CreatedAt,
			MergeMessage:  mergeMessage,
			ChangeModules: changeModules,
		}

		/*
			log.Logger.Info().
				Str("GiraId", newPr.GiraId).
				Str("GiraUrl", newPr.GiraUrl).
				Str("Title", newPr.Title).
				Str("Author", newPr.Author).
				Str("Merge", TimeToBeijing(*newPr.Pr.MergedAt)).
				Str("CreateAt", TimeToBeijing(*newPr.CreateAt)).
				Str("MergeMessage", newPr.MergeMessage).
				Str("ChangeModules", newPr.ChangeModules).
				Msgf("GitHub Pull Request\n")
		*/
		mrMap[fmt.Sprint(*pr.Number)] = newPr
	}
	return mrMap
}
