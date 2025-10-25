package handler

import (
	"net/http"
	"time"

	"github.com/felix-001/qnHackathon/internal/model"
	"github.com/felix-001/qnHackathon/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type ReleaseHandler struct {
	service *service.ReleaseService
	manager *service.Manager
}

func NewReleaseHandler(service *service.ReleaseService, manager *service.Manager) *ReleaseHandler {
	return &ReleaseHandler{
		service: service,
		manager: manager,
	}
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

	go func() {
		log.Logger.Info().Msgf("Starting build for release %s", release.ID)
		gitlabPRURL := h.manager.Build()
		if gitlabPRURL != "" {
			log.Logger.Info().Msgf("Build completed, GitLab PR: %s", gitlabPRURL)
			if err := h.service.UpdateGitlabPR(release.ID, gitlabPRURL); err != nil {
				log.Logger.Error().Err(err).Msg("Failed to update GitLab PR URL")
			}
		} else {
			log.Logger.Error().Msg("Build failed or returned empty GitLab PR URL")
		}
	}()

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

func (h *ReleaseHandler) ApproveReview(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.ApproveReview(id); err != nil {
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

func (h *ReleaseHandler) Deploy(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.StartDeploy(id); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    1,
			Message: err.Error(),
		})
		return
	}

	go func() {
		time.Sleep(2 * time.Second)
		if err := h.service.CompleteDeploy(id); err != nil {
			log.Logger.Error().Err(err).Msg("Failed to complete deployment")
		}
	}()

	c.JSON(http.StatusOK, model.Response{
		Code:    0,
		Message: "success",
	})
}
