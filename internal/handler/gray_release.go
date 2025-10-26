package handler

import (
	"net/http"

	"github.com/felix-001/qnHackathon/internal/model"
	"github.com/felix-001/qnHackathon/internal/service"
	"github.com/gin-gonic/gin"
)

type GrayReleaseHandler struct {
	grayReleaseService *service.GrayReleaseService
}

func NewGrayReleaseHandler(grayReleaseService *service.GrayReleaseService) *GrayReleaseHandler {
	return &GrayReleaseHandler{
		grayReleaseService: grayReleaseService,
	}
}

func (h *GrayReleaseHandler) Create(c *gin.Context) {
	var config model.GrayReleaseConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	err := h.grayReleaseService.CreateGrayRelease(&config)
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

func (h *GrayReleaseHandler) List(c *gin.Context) {
	projectID := c.Query("projectId")
	environment := c.Query("environment")

	configs, err := h.grayReleaseService.ListGrayReleases(projectID, environment)
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
		Data:    configs,
	})
}

func (h *GrayReleaseHandler) Get(c *gin.Context) {
	id := c.Param("id")

	config, err := h.grayReleaseService.GetGrayRelease(id)
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

func (h *GrayReleaseHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var config model.GrayReleaseConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	err := h.grayReleaseService.UpdateGrayRelease(id, &config)
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

func (h *GrayReleaseHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.grayReleaseService.DeleteGrayRelease(id)
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
	})
}

func (h *GrayReleaseHandler) GetDeviceStats(c *gin.Context) {
	projectID := c.Query("projectId")
	environment := c.Query("environment")

	stats, err := h.grayReleaseService.GetDeviceStats(projectID, environment)
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
		Data:    stats,
	})
}

type FullReleaseRequest struct {
	ProjectID   string `json:"projectId"`
	Environment string `json:"environment"`
	Version     string `json:"version"`
	Operator    string `json:"operator"`
}

func (h *GrayReleaseHandler) FullRelease(c *gin.Context) {
	var req FullReleaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	err := h.grayReleaseService.FullRelease(req.ProjectID, req.Environment, req.Version, req.Operator)
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
	})
}

func (h *GrayReleaseHandler) UpdateDeviceStatus(c *gin.Context) {
	var device model.DeviceGrayStatus
	if err := c.ShouldBindJSON(&device); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	err := h.grayReleaseService.UpdateDeviceStatus(&device)
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
	})
}

func (h *GrayReleaseHandler) CheckDeviceGrayRule(c *gin.Context) {
	var device model.DeviceGrayStatus
	if err := c.ShouldBindJSON(&device); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	matched, version, err := h.grayReleaseService.CheckDeviceGrayRule(&device)
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
		Data: map[string]interface{}{
			"matched": matched,
			"version": version,
		},
	})
}
