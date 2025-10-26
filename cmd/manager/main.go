package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	cfg "github.com/felix-001/qnHackathon/internal/config"
	"github.com/felix-001/qnHackathon/internal/db"
	"github.com/felix-001/qnHackathon/internal/handler"
	"github.com/felix-001/qnHackathon/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	configFile := flag.String("f", "./config/manager.json", "the config file")
	flag.Parse()
	cfg := &cfg.Config{}
	bytes, err := os.ReadFile(*configFile)
	if err != nil {
		log.Println("read fail", *configFile, err)
		return
	}
	err = json.Unmarshal(bytes, cfg)
	if err != nil {
		log.Println("unmarshal fail", *configFile, err)
		return
	}

	mongodb, err := db.NewMongoDB(cfg.MongoConf)
	if err != nil {
		log.Println("Failed to connect to MongoDB:", err)
		return
	}
	defer mongodb.Close()

	mgr := service.NewManager(cfg)
	r := gin.Default()

	r.Use(cors.Default())

	r.LoadHTMLGlob("web/templates/*.html")

	projectService := service.NewProjectService(mongodb)
	releaseService := service.NewReleaseService(mongodb)
	monitoringService := service.NewMonitoringService()
	binService := service.NewBinService()
	configService := service.NewConfigService(mongodb)

	projectHandler := handler.NewProjectHandler(projectService)
	releaseHandler := handler.NewReleaseHandler(releaseService, mgr, projectService)
	monitoringHandler := handler.NewMonitoringHandler(monitoringService)
	binHandler := handler.NewBinHandler(binService)
	gitlabMgr := service.NewGitLabMgr(cfg.GitlabConf)
	binHandler.SetGitLabMgr(gitlabMgr)
	binHandler.SetReleaseService(releaseService)
	configHandler := handler.NewConfigHandler(configService)
	configHandler.SetGitLabMgr(gitlabMgr)
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
		api.GET("/configs/:id/history", configHandler.GetHistory)
		api.GET("/configs/history", configHandler.GetHistoryByProject)
		api.GET("/configs/compare", configHandler.Compare)
		api.POST("/configs/rollouts", configHandler.CreateRollout)
		api.GET("/configs/rollouts", configHandler.ListRollouts)
		api.POST("/configs/publish", configHandler.PublishAll)
		api.GET("/configs/device-stats", configHandler.GetDeviceStats)

		api.GET("/keepalive", binHandler.GetKeepalive)
		api.POST("/keepalive", binHandler.PostKeepalive)
		api.GET("/bins/:bin_name", binHandler.GetBin)
		api.POST("/bins/:bin_name", binHandler.PostBin)
		api.POST("/bins/:bin_name/progress", binHandler.PostProgress)
		api.GET("/download/:bin_file_name", binHandler.Download)
	}

	r.GET("/health", binHandler.Health)

	r.Run(":38012")
}
