package service

import (
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
