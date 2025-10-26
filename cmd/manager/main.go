package main

import (
	"encoding/json"
	"flag"
	"os"

	cfg "github.com/felix-001/qnHackathon/internal/config"
	"github.com/felix-001/qnHackathon/internal/db"
	"github.com/felix-001/qnHackathon/internal/handler"
	"github.com/felix-001/qnHackathon/internal/service"
	"github.com/felix-001/qnHackathon/internal/util"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func main() {
	util.InitLogger()
	
	configFile := flag.String("f", "./config/manager.json", "the config file")
	flag.Parse()
	cfg := &cfg.Config{}
	bytes, err := os.ReadFile(*configFile)
	if err != nil {
		log.Error().Err(err).Str("file", *configFile).Msg("读取配置文件失败")
		return
	}
	err = json.Unmarshal(bytes, cfg)
	if err != nil {
		log.Error().Err(err).Str("file", *configFile).Msg("解析配置文件失败")
		return
	}

	mongodb, err := db.NewMongoDB(cfg.MongoConf)
	if err != nil {
		log.Error().Err(err).Msg("连接 MongoDB 失败")
		return
	}
	defer mongodb.Close()

	mgr := service.NewManager(cfg)
	r := gin.Default()

	r.Use(cors.Default())

	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/templates/*.html")

	projectService := service.NewProjectService(mongodb)
	releaseService := service.NewReleaseService(mongodb)
	monitoringService := service.NewMonitoringService()
	binService := service.NewBinService()
	configService := service.NewConfigService(mongodb)
	grayReleaseService := service.NewGrayReleaseService(mongodb)

	projectHandler := handler.NewProjectHandler(projectService)
	releaseHandler := handler.NewReleaseHandler(releaseService, mgr, projectService)
	monitoringHandler := handler.NewMonitoringHandler(monitoringService)
	binHandler := handler.NewBinHandler(binService)
	gitlabMgr := service.NewGitLabMgr(cfg.GitlabConf)
	binHandler.SetGitLabMgr(gitlabMgr)
	binHandler.SetReleaseService(releaseService)
	configHandler := handler.NewConfigHandler(configService)
	configHandler.SetGitLabMgr(gitlabMgr)
	grayReleaseHandler := handler.NewGrayReleaseHandler(grayReleaseService)
	webHandler := handler.NewWebHandler()

	r.GET("/", webHandler.Index)
	r.GET("/projects", webHandler.Projects)
	r.GET("/releases", webHandler.Releases)
	r.GET("/monitoring", webHandler.Monitoring)
	r.GET("/config", webHandler.Config)

	api := r.Group("/api/v1")
	{
		api.GET("/projects", projectHandler.List)
		api.POST("/projects", projectHandler.Create)
		api.PUT("/projects/:id", projectHandler.Update)
		api.DELETE("/projects/:id", projectHandler.Delete)

		api.GET("/releases", releaseHandler.List)
		api.POST("/releases", releaseHandler.Create)
		api.POST("/releases/batch-delete", releaseHandler.BatchDelete)
		api.GET("/releases/:id", releaseHandler.Get)
		api.POST("/releases/:id/rollback", releaseHandler.Rollback)
		api.POST("/releases/:id/approve", releaseHandler.Approve)
		api.POST("/releases/:id/deploy", releaseHandler.Deploy)

		api.GET("/monitoring/realtime", monitoringHandler.GetRealtime)
		api.GET("/monitoring/timeseries", monitoringHandler.GetTimeSeries)

		api.GET("/configs", configHandler.List)
		api.POST("/configs", configHandler.Create)
		api.GET("/configs/:id", configHandler.Get)
		api.PUT("/configs/:id", configHandler.Update)
		api.DELETE("/configs/:id", configHandler.Delete)
		api.POST("/configs/:id/rollback", configHandler.Rollback)
		api.GET("/configs/:id/history", configHandler.GetHistory)
		api.GET("/configs/history", configHandler.GetHistoryByProject)
		api.GET("/configs/compare", configHandler.Compare)
		api.GET("/configs/versions", configHandler.GetVersions)

		api.GET("/gray-releases", grayReleaseHandler.List)
		api.POST("/gray-releases", grayReleaseHandler.Create)
		api.GET("/gray-releases/:id", grayReleaseHandler.Get)
		api.PUT("/gray-releases/:id", grayReleaseHandler.Update)
		api.DELETE("/gray-releases/:id", grayReleaseHandler.Delete)
		api.GET("/gray-releases/device-stats", grayReleaseHandler.GetDeviceStats)
		api.POST("/gray-releases/full-release", grayReleaseHandler.FullRelease)
		api.POST("/gray-releases/device-status", grayReleaseHandler.UpdateDeviceStatus)
		api.POST("/gray-releases/check-rule", grayReleaseHandler.CheckDeviceGrayRule)

		api.GET("/keepalive", binHandler.GetKeepalive)
		api.POST("/keepalive", binHandler.PostKeepalive)
		api.GET("/bins/:bin_name", binHandler.GetBin)
		api.POST("/bins/:bin_name", binHandler.PostBin)
		api.POST("/bins/:bin_name/progress", binHandler.PostProgress)
		api.GET("/download/:bin_file_name", binHandler.Download)
	}

	r.GET("/health", binHandler.Health)

	log.Info().Str("port", "38012").Msg("服务启动成功")
	r.Run(":38012")
}
