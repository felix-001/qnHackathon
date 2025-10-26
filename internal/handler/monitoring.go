package handler

import (
	"net/http"

	"github.com/felix-001/qnHackathon/internal/model"
	"github.com/felix-001/qnHackathon/internal/service"
	"github.com/gin-gonic/gin"
)

type MonitoringHandler struct {
	service *service.MonitoringService
}

func NewMonitoringHandler(service *service.MonitoringService) *MonitoringHandler {
	return &MonitoringHandler{service: service}
}

func (h *MonitoringHandler) GetRealtime(c *gin.Context) {
	releaseID := c.Query("releaseId")
	metrics, err := h.service.GetRealtime(releaseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    1,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    0,
		Message: "success",
		Data: map[string]interface{}{
			"metrics": metrics,
		},
	})
}

func (h *MonitoringHandler) GetTimeSeries(c *gin.Context) {
	releaseID := c.Query("releaseId")
	timeSeries, err := h.service.GetTimeSeries(releaseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    1,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    0,
		Message: "success",
		Data: map[string]interface{}{
			"timeSeries": timeSeries,
		},
	})
}

func (h *MonitoringHandler) GetNodesByProject(c *gin.Context) {
	projectID := c.Query("projectId")
	nodes, err := h.service.GetNodesByProject(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    1,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    0,
		Message: "success",
		Data: map[string]interface{}{
			"nodes": nodes,
		},
	})
}

func (h *MonitoringHandler) GetNodeMetrics(c *gin.Context) {
	nodeID := c.Query("nodeId")
	metrics, err := h.service.GetNodeMetrics(nodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    1,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    0,
		Message: "success",
		Data: map[string]interface{}{
			"metrics": metrics,
		},
	})
}
