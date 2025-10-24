package handler

import (
	"net/http"

	"github.com/felix-001/qnHackathon/internal/model"
	"github.com/felix-001/qnHackathon/internal/service"
	"github.com/gin-gonic/gin"
)

type ReleaseHandler struct {
	service *service.ReleaseService
}

func NewReleaseHandler(service *service.ReleaseService) *ReleaseHandler {
	return &ReleaseHandler{service: service}
}

func (h *ReleaseHandler) List(c *gin.Context) {
	releases := h.service.List()
	c.JSON(http.StatusOK, model.Response{
		Code:    0,
		Message: "success",
		Data:    releases,
	})
}

func (h *ReleaseHandler) Create(c *gin.Context) {
	var release model.Release
	if err := c.ShouldBindJSON(&release); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    1,
			Message: err.Error(),
		})
		return
	}

	if err := h.service.Create(&release); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    1,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    0,
		Message: "success",
		Data:    release,
	})
}

func (h *ReleaseHandler) Get(c *gin.Context) {
	id := c.Param("id")
	release, err := h.service.Get(id)
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
		Data:    release,
	})
}

func (h *ReleaseHandler) Rollback(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		TargetVersion string `json:"targetVersion"`
		Reason        string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    1,
			Message: err.Error(),
		})
		return
	}

	if err := h.service.Rollback(id, req.TargetVersion, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    1,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    0,
		Message: "success",
	})
}
