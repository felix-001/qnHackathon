package service

import (
	"strings"

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

type BuildInfo struct {
	GitlabPRURL string
	TarFileName string
}

func (m *Manager) Build() *BuildInfo {
	m.jenkinsMgr.StartJob()

	buildResult := m.jenkinsMgr.WaitForJobCompletion()
	if buildResult == nil || !buildResult.IsSuccess {
		log.Logger.Error().Msg("Jenkins build failed or timed out")
		return nil
	}

	streamdPath, err := m.jenkinsMgr.DownloadBin(buildResult, "streamd")
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to download streamd")
		return nil
	}
	log.Logger.Info().Msgf("Successfully downloaded streamd to: %s", streamdPath)

	parts := strings.Split(streamdPath, "/")
	if len(parts) != 3 {
		log.Logger.Info().Msgf("parse streamdPath err")
		return nil
	}

	m.gitlabMgr.UpdateVersion("streamd.json", parts[2])
	mrUrl := m.gitlabMgr.GetMrUrl(parts[2])
	if mrUrl == "" {
		log.Logger.Error().Msg("GetMrUrl: 无法获取 MergeRequest URL")
		return nil
	}
	log.Logger.Info().Msgf("GetMrUrl: %s", mrUrl)

	return &BuildInfo{
		GitlabPRURL: mrUrl,
		TarFileName: buildResult.TarFileName,
	}
}
