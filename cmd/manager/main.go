package main

import (
	"flag"
	"log"

	cfg "github.com/felix-001/qnHackathon/internal/config"

	"github.com/felix-001/qnHackathon/internal/handler"
	"github.com/felix-001/qnHackathon/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	qconfig "github.com/qiniu/x/config"
)

var (
	// configFile 配置文件
	configFile = flag.String("f", "./config/manager.json", "the config file")
)

func main() {
	flag.Parse()
	cfg := &cfg.Config{}
	err := qconfig.LoadFile(cfg, *configFile)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}
	mgr := service.NewManager(cfg)
	mgr.Run()
	r := gin.Default()

	r.Use(cors.Default())

	r.LoadHTMLGlob("web/templates/*.html")

	projectService := service.NewProjectService()
	releaseService := service.NewReleaseService()
	monitoringService := service.NewMonitoringService()

	projectHandler := handler.NewProjectHandler(projectService)
	releaseHandler := handler.NewReleaseHandler(releaseService)
	monitoringHandler := handler.NewMonitoringHandler(monitoringService)
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
		api.GET("/releases/:id", releaseHandler.Get)
		api.POST("/releases/:id/rollback", releaseHandler.Rollback)

		api.GET("/monitoring/realtime", monitoringHandler.GetRealtime)
	}

	r.Run(":8081")
}
