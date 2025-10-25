package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/bndr/gojenkins"
	cfg "github.com/felix-001/qnHackathon/internal/config"
)

type BuildResult struct {
	BuildID     int64            `json:"buildID"`
	BuildURL    string           `json:"buildURL"`
	IsFinished  bool             `json:"isFinished"`
	IsSuccess   bool             `json:"isSuccess"`
	Result      string           `json:"result"`
	Build       *gojenkins.Build `json:"build"`       // 构建信息
	BuildError  error            `json:"buildError"`  // 构建错误信息
	README      string           `json:"readme"`      // README 文件内容
	TarFileName string           `json:"tarFilePath"` // 构建产物文件名
}

type JenkinsMgr struct {
	Client          *gojenkins.Jenkins
	Build           *BuildResult
	BuildidToModule string
	Conf            cfg.JenkinsConf
}

func NewJenkinsMgr(conf cfg.JenkinsConf) *JenkinsMgr {
	client := NewJenkinsClient(conf)
	if client == nil {
		log.Logger.Error().Msg("新建JenkinsClient失败")
		return nil
	}

	return &JenkinsMgr{
		Client: client,
		Conf:   conf,
	}
}

func NewJenkinsClient(conf cfg.JenkinsConf) *gojenkins.Jenkins {
	jenkins := gojenkins.CreateJenkins(nil, conf.JenkinsURL, conf.Username, conf.ApiToken)
	ctx := context.Background()
	client, err := jenkins.Init(ctx)
	if err != nil {
		log.Logger.Error().Msgf("连接Jenkins失败: %v, conf: %+v", err, conf)
		return nil
	}
	return client
}

// https://jenkins.qiniu.io/job/mikud-live-module-pipeline/
func (s *JenkinsMgr) GetJenKinsBuilds() []*gojenkins.Build {
	ctx := context.Background()
	// 获取指定的 Job
	job, err := s.Client.GetJob(ctx, s.Conf.ProjectID)
	if err != nil {
		log.Logger.Error().Msgf("获取Job失败: %v", err)
		return nil
	}
	log.Printf("Job 名称: %s", job.GetName())

	// 获取所有构建
	buildidlst, err := job.GetAllBuildIds(ctx)
	if err != nil {
		log.Logger.Error().Msgf("获取所有构建失败: %v", err)
		return nil
	}

	// 筛选出成功的构建
	cnt := 0
	successfulBuilds := []*gojenkins.Build{}
	for _, buildid := range buildidlst {
		//log.Logger.Error().Msgf("构建编号: %+v", buildid)
		build, err := job.GetBuild(ctx, buildid.Number)
		if err != nil {
			log.Logger.Error().Msgf("获取构建信息失败: %v", err)
			continue
		}
		ar := build.GetArtifacts()
		for _, artifact := range ar {
			if artifact.FileName == "" {
				log.Logger.Error().Msgf("构建编号: %d, 没有构建产物", build.Raw.QueueID)
				continue
			}
			log.Logger.Error().Msgf("构建编号: %d, %d, 文件名: %s", build.Raw.QueueID, buildid.Number, artifact.FileName)
			if artifact.FileName == "README" {
				data, err := artifact.GetData(ctx)
				if err != nil {
					log.Logger.Error().Msgf("获取 README.md 内容失败: %v", err)
					continue
				}
				log.Logger.Error().Msgf("readme data: %s", string(data))
				continue
			}
		}

		log.Logger.Info().Msgf("Artifacts: %+v, id: %d", build.GetArtifacts(), build.Raw.QueueID)
		//log.Logger.Error().Msgf("构建编号: %+v", build)
		successfulBuilds = append(successfulBuilds, build)
		cnt++
		if cnt > 10 {
			break
		}
	}

	return successfulBuilds
}

func (s *JenkinsMgr) GetBuildParams(buildbin []string, goversion string) map[string][]string {
	params := make(map[string][]string)
	params["DESCRIPTION"] = []string{"自动发布线上服务"}
	params["BRANCH"] = []string{"main"}
	if goversion == "" {
		goversion = "miku_go1.22.9"
	}
	params["GO_VERSION"] = []string{goversion}
	params["BIN"] = buildbin
	now := time.Now()
	params["PACKAGE_NAME"] = []string{fmt.Sprintf("MIKUD_LIVE.%s.tar.gz", now.Format("2006-01-02-15-04-05"))}
	params["REPORTED"] = []string{"false"}
	params["TAG"] = []string{"origin/main"}

	return params
}

func (s *JenkinsMgr) GetBuildQueryString() map[string]string {
	querys := make(map[string]string)

	querys["quickFilter"] = ""
	querys["statusCode"] = "303"
	querys["redirectTo"] = "."
	querys["Submit"] = "Build"

	return querys
}

func getGoVersion(buildbin []string) (version string) {
	if len(buildbin) == 0 || len(buildbin) > 1 {
		return ""
	}
	switch buildbin[0] {
	case "streamd":
		return "miku_go1.20.11"
	case "collector":
		return "miku_go1.23.4"
	case "etlv2":
		return "miku-ubuntu22.04_mvn"
	case "sched", "netprobe-srv", "netprobe-cli", "agent", "lived":
		return "miku_go1.22.9"
	default:
		return ""
	}
}

func BuildJob(ctx context.Context, j *gojenkins.Job, params map[string][]string, querys map[string]string) (int64, error) {
	endpoint := "/build?delay=0sec"
	parameters, err := j.GetParameters(ctx)
	if err != nil {
		return 0, err
	}
	if len(parameters) > 0 {
		endpoint = "/buildWithParameters"
	}
	data := url.Values{}
	for k, v := range params {
		for _, val := range v {
			data.Add(k, val)
		}
	}
	resp, err := j.Jenkins.Requester.Post(ctx, j.Base+endpoint, bytes.NewBufferString(data.Encode()), nil, querys)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return 0, fmt.Errorf("Could not invoke job %q: %s", j.GetName(), resp.Status)
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return 0, errors.New("Don't have key \"Location\" in response of header")
	}

	u, err := url.Parse(location)
	if err != nil {
		return 0, err
	}

	number, err := strconv.ParseInt(path.Base(u.Path), 10, 64)
	if err != nil {
		return 0, err
	}

	return number, nil
}

func (s *JenkinsMgr) BuildJenKinsBuildOne(buildbins []string) int64 {
	goversion := getGoVersion(buildbins)
	if goversion == "" {
		return 0
	}

	// 需要单独构建的模块
	params := s.GetBuildParams(buildbins, goversion)
	// 执行构建
	ctx := context.Background()
	job, err := s.Client.GetJob(ctx, s.Conf.ProjectID)
	if err != nil {
		log.Logger.Error().Msgf("获取Job失败: %v", err)
	}
	buildid, err := BuildJob(ctx, job, params, s.GetBuildQueryString())
	if buildid != 0 {
		log.Logger.Info().Msgf("JenKins任务提交成功: %v, buildid: %d", params, buildid)
	}

	if buildid == 0 {
		log.Logger.Error().Msgf("JenKins任务提交失败: %v, params: %+v, ", err, params)
	}
	return buildid
}

func (s *JenkinsMgr) BuildJenKinsBuild(module string) int64 {
	buildid := s.BuildJenKinsBuildOne([]string{module})
	if buildid != 0 {
		log.Logger.Info().Msgf("Jenkins 任务提交成功: %s, buildid: %d", module, buildid)
	} else {
		log.Logger.Error().Msgf("Jenkins 任务提交失败: %s", module)
	}
	s.BuildidToModule = module
	return buildid
}

func (s *JenkinsMgr) GetBuildStatus(buildid int64) (isbuilding bool, buildmoreInfo *BuildResult) {
	ctx := context.Background()
	job, err := s.Client.GetJob(ctx, s.Conf.ProjectID)
	if err != nil {
		log.Logger.Error().Msgf("获取Job失败: %v", err)
		return false, nil
	}

	buildidlst, err := job.GetAllBuildIds(ctx)
	if err != nil {
		log.Logger.Error().Msgf("获取所有构建失败: %v", err)
		return false, nil
	}

	for _, build := range buildidlst {
		buildtask, err := job.GetBuild(ctx, build.Number)
		if err != nil {
			log.Logger.Error().Msgf("获取构建信息失败: %v, build: %d", err, build.Number)
			continue
		}

		if buildtask.Raw.QueueID != buildid {
			continue
		}

		if buildtask.Raw.Building {
			log.Logger.Info().Msgf("构建编号: %d, 正在构建中", buildtask.Raw.QueueID)
			return true, nil
		}

		buildResult := &BuildResult{
			BuildID:    buildtask.Raw.QueueID,
			IsFinished: true,
			BuildURL:   buildtask.GetUrl(),
			Result:     buildtask.GetResult(),
			Build:      buildtask,
		}

		result := buildtask.GetResult()
		if result != "SUCCESS" {
			log.Logger.Error().Msgf("构建编号: %d, 构建失败, 结果: %s", buildtask.Raw.QueueID, result)
			buildResult.BuildError = fmt.Errorf("构建编号: %d, 构建失败, 结果: %s", buildtask.Raw.QueueID, result)
			return false, buildResult
		}

		artifacts := buildtask.GetArtifacts()
		if len(artifacts) == 0 {
			log.Logger.Error().Msgf("构建编号: %d, 没有构建产物", buildtask.Raw.QueueID)
			return false, buildResult
		}

		buildResult.TarFileName = ""
		buildResult.README = ""
		for _, artifact := range artifacts {
			log.Logger.Info().Msgf("构建编号: %d, 构建产物: %s", buildtask.Raw.QueueID, artifact.FileName)
			if strings.HasSuffix(artifact.FileName, ".tar.gz") {
				buildResult.TarFileName = artifact.FileName
			}
			if strings.HasSuffix(artifact.FileName, "README") {
				readme, err := artifact.GetData(ctx)
				if err != nil {
					log.Logger.Error().Msgf("获取 README.md 内容失败: %v", err)
				} else {
					buildResult.README = string(readme)
				}
			}
		}
		if buildResult.TarFileName != "" && buildResult.README != "" {
			buildResult.IsSuccess = true
		}

		log.Logger.Info().Msgf("Jenkins 构建 ID: %d, URL: %s, 状态: %t, 结果: %s, Building: %t, tar: %s, readme: %s",
			buildtask.GetBuildNumber(), buildtask.GetUrl(), buildtask.IsGood(ctx), result,
			buildtask.Raw.Building, buildResult.TarFileName, buildResult.README)
		return buildtask.Raw.Building, buildResult
	}

	return true, nil
}

func (s *JenkinsMgr) JenkBuild(module string) bool {
	buildid := s.BuildJenKinsBuild(module)
	start_time := time.Now()
	jenTicker := time.NewTicker(1 * time.Minute) // 每分钟检查一次
	defer func() {
		log.Logger.Info().Msgf("停止Jenkins构建状态检查, time cost: %d ms", time.Since(start_time).Milliseconds())

	}()
	defer jenTicker.Stop()

	for range jenTicker.C {
		log.Logger.Info().Msgf("Checking buildids: %v, module: %s", buildid, module)
		isrunning, buildret := s.GetBuildStatus(buildid)
		if !isrunning && buildret != nil && buildret.IsSuccess {
			s.Build = buildret
			return true
		}
		if !isrunning {
			log.Logger.Error().Msgf("build end, %d, module: %s", buildid, module)
			if buildret != nil {
				s.Build = buildret
			}
			return false
		}

		log.Logger.Info().Msgf("Build %d for %s is still running.", buildid, module)
		if time.Since(start_time) > 3*time.Hour {
			// 发送告警
			log.Logger.Error().Msgf("Builds for GitHub are still running after 3 hours, stopping checks.")
			s.Build = &BuildResult{
				BuildID:    buildid,
				IsFinished: false,
				Result:     "time too long",
			}
			break
		}
	}
	return false
}

func (s *JenkinsMgr) StartJob() {
	_, err := s.Client.BuildJob(context.Background(), "streamd", nil)
	if err != nil {
		log.Logger.Error().Err(err).Msg("BuildJob")
		return
	}
}

func (s *JenkinsMgr) WaitForJobCompletion() *BuildResult {
	ctx := context.Background()
	job, err := s.Client.GetJob(ctx, s.Conf.ProjectID)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to get job")
		return nil
	}

	startTime := time.Now()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			lastBuild, err := job.GetLastBuild(ctx)
			if err != nil {
				log.Logger.Error().Err(err).Msg("Failed to get last build")
				continue
			}

			if lastBuild.IsRunning(ctx) {
				log.Logger.Info().Msgf("Build %d is still running...", lastBuild.GetBuildNumber())
			} else {
				isBuilding, buildResult := s.GetBuildStatus(lastBuild.Raw.QueueID)
				if !isBuilding && buildResult != nil && buildResult.IsFinished {
					if buildResult.IsSuccess {
						log.Logger.Info().Msgf("Build completed successfully: %d", buildResult.BuildID)
						return buildResult
					}
					log.Logger.Error().Msgf("Build failed: %v", buildResult.BuildError)
					return buildResult
				}
			}

			if time.Since(startTime) > 3*time.Hour {
				log.Logger.Error().Msg("Build timeout after 3 hours")
				return &BuildResult{
					IsFinished: true,
					IsSuccess:  false,
					Result:     "timeout",
				}
			}
		}
	}
}

func (s *JenkinsMgr) DownloadStreamd(buildResult *BuildResult) (string, error) {
	if buildResult == nil || buildResult.Build == nil {
		return "", fmt.Errorf("invalid build result")
	}

	ctx := context.Background()
	artifacts := buildResult.Build.GetArtifacts()
	
	for _, artifact := range artifacts {
		if artifact.FileName == "streamd" {
			downloadDir := "./downloads"
			success, err := artifact.SaveToDir(ctx, downloadDir)
			if err != nil {
				return "", fmt.Errorf("failed to save artifact: %w", err)
			}
			if !success {
				return "", fmt.Errorf("failed to save artifact: unknown error")
			}
			
			downloadPath := fmt.Sprintf("%s/%s", downloadDir, artifact.FileName)
			log.Logger.Info().Msgf("Successfully downloaded streamd to: %s", downloadPath)
			return downloadPath, nil
		}
	}
	
	return "", fmt.Errorf("streamd artifact not found in build")
}
