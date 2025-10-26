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
	machineID := c.Query("machineId")
	metrics, err := h.service.GetRealtime(releaseID, machineID)
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
	machineID := c.Query("machineId")
	timeSeries, err := h.service.GetTimeSeries(releaseID, machineID)
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
