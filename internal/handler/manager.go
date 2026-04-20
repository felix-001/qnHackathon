package handler

import (
	"net/http"

	"github.com/felix-001/qnHackathon/internal/service"
	"github.com/gin-gonic/gin"
)

type ManagerHandler struct {
	manager *service.Manager
}

func NewManagerHandler(manager *service.Manager) *ManagerHandler {
	return &ManagerHandler{
		manager: manager,
	}
}

func (h *ManagerHandler) ListBins(c *gin.Context) {
	bins := h.manager.ListBins()
	c.JSON(http.StatusOK, gin.H{
		"bins": bins,
	})
}
