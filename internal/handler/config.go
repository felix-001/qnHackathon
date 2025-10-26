package handler

import (
	"net/http"

	"github.com/felix-001/qnHackathon/internal/model"
	"github.com/felix-001/qnHackathon/internal/service"
	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	configService *service.ConfigService
}

func NewConfigHandler(configService *service.ConfigService) *ConfigHandler {
	return &ConfigHandler{
		configService: configService,
	}
}

func (h *ConfigHandler) Get(c *gin.Context) {
	projectName := c.Param("name")
	if projectName == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "Project name is required",
		})
		return
	}

	config, err := h.configService.GetConfig(projectName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "success",
		Data:    config,
	})
}

func (h *ConfigHandler) Update(c *gin.Context) {
	projectName := c.Param("name")
	if projectName == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "Project name is required",
		})
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "Invalid request body",
		})
		return
	}

	if err := h.configService.SaveConfig(projectName, req.Content); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "Config saved successfully",
	})
}
