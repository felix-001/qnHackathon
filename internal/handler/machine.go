package handler

import (
	"net/http"

	"github.com/felix-001/qnHackathon/internal/model"
	"github.com/felix-001/qnHackathon/internal/service"
	"github.com/gin-gonic/gin"
)

type MachineHandler struct {
	service *service.MachineService
}

func NewMachineHandler(service *service.MachineService) *MachineHandler {
	return &MachineHandler{service: service}
}

func (h *MachineHandler) ListByProject(c *gin.Context) {
	projectID := c.Query("projectId")
	machines, err := h.service.ListByProject(projectID)
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
		Data:    machines,
	})
}
