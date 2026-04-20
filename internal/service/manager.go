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
		log.Error().Msg("Jenkins 构建失败或超时")
		return nil
	}

	streamdPath, err := m.jenkinsMgr.DownloadBin(buildResult, "streamd")
	if err != nil {
		log.Error().Err(err).Msg("下载 streamd 失败")
		return nil
	}
	log.Info().Str("path", streamdPath).Msg("下载 streamd 成功")

	parts := strings.Split(streamdPath, "/")
	if len(parts) != 3 {
		log.Error().Str("path", streamdPath).Msg("解析 streamdPath 失败")
		return nil
	}

	m.gitlabMgr.UpdateVersion("streamd.json", parts[2])
	mrUrl := m.gitlabMgr.GetMrUrl(parts[2])
	if mrUrl == "" {
		log.Error().Msg("无法获取 MergeRequest URL")
		return nil
	}
	log.Info().Str("url", mrUrl).Msg("获取 MR URL 成功")

	return &BuildInfo{
		GitlabPRURL: mrUrl,
		TarFileName: buildResult.TarFileName,
	}
}

func (m *Manager) ListBins() []string {
	return []string{"stremad"}
}
