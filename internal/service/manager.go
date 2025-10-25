package service

import (
	"fmt"

	cfg "github.com/felix-001/qnHackathon/internal/config"
	"github.com/rs/zerolog/log"
)

type Manager struct {
	githubMgr  *GitHubMgr
	jenkinsMgr *JenkinsMgr
	gitlabMgr  *GitLabMgr
}

func NewManager(conf *cfg.Config) *Manager {
	return &Manager{
		githubMgr:  NewGitHubMgr(conf.GitHubConf),
		jenkinsMgr: NewJenkinsMgr(conf.JenkinsConf),
		gitlabMgr:  NewGitLabMgr(conf.GitlabConf),
	}
}

func (m *Manager) Run() {
	m.jenkinsMgr.StartJob()

	buildResult := m.jenkinsMgr.WaitForJobCompletion()
	if buildResult == nil || !buildResult.IsSuccess {
		log.Logger.Error().Msg("Jenkins build failed or timed out")
		return
	}

	streamdPath, err := m.jenkinsMgr.DownloadStreamd(buildResult)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to download streamd")
		return
	}
	log.Logger.Info().Msgf("Successfully downloaded streamd to: %s", streamdPath)

	m.gitlabMgr.UpdateVersion("streamd.json", "streamd-20251025-14-38-30.tar.gz")
	mrUrl := m.gitlabMgr.GetMrUrl("streamd-20251025-14-38-30.tar.gz")
	if mrUrl == "" {
		log.Logger.Error().Msg("GetMrUrl: 无法获取 MergeRequest URL")
		return
	}
	log.Logger.Info().Msgf("GetMrUrl: %s", mrUrl)

	//m.gitlabMgr.GetMergeRequest()
	//m.gitlabMgr.CreateBranch("streamd")
	//m.gitlabMgr.GetFile("streamd", "streamd.json")
}

func (m *Manager) Build(version string) (string, error) {
	branch := m.gitlabMgr.CreateBranch("streamd")
	if branch == "" {
		log.Logger.Error().Msg("Failed to create branch")
		return "", fmt.Errorf("failed to create branch")
	}

	m.gitlabMgr.UpdateVersion("streamd.json", version)
	
	success := m.gitlabMgr.CommitPush(branch, "Release "+version, "Automated release for version "+version)
	if !success {
		log.Logger.Error().Msg("Failed to create merge request")
		return "", fmt.Errorf("failed to create merge request")
	}

	log.Logger.Info().Msgf("Created GitLab MR: %s", m.gitlabMgr.MrUrl)
	return m.gitlabMgr.MrUrl, nil
}
