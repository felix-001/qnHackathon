package handler

import (
	"net/http"

	"github.com/felix-001/qnHackathon/internal/model"
	"github.com/felix-001/qnHackathon/internal/service"
	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	configService *service.ConfigService
	gitlabMgr     *service.GitLabMgr
}

func NewConfigHandler(configService *service.ConfigService) *ConfigHandler {
	return &ConfigHandler{
		configService: configService,
	}
}

func (h *ConfigHandler) SetGitLabMgr(gitlabMgr *service.GitLabMgr) {
	h.gitlabMgr = gitlabMgr
}

func (h *ConfigHandler) List(c *gin.Context) {
	projectID := c.Query("projectId")
	environment := c.Query("environment")

	configs, err := h.configService.List(projectID, environment)
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

func (h *ConfigHandler) Get(c *gin.Context) {
	id := c.Param("id")

	config, err := h.configService.Get(id)
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

type CreateConfigRequest struct {
	Config   *model.Config `json:"config"`
	Operator string        `json:"operator"`
	Reason   string        `json:"reason"`
	SubmitToGitLab bool    `json:"submitToGitLab"`
}

func (h *ConfigHandler) Create(c *gin.Context) {
	var req CreateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	err := h.configService.Create(req.Config, req.Operator, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	var mrURL string
	if req.SubmitToGitLab && h.gitlabMgr != nil {
		mrURL, err = h.configService.SubmitToGitLab(req.Config, h.gitlabMgr, req.Operator, req.Reason)
		if err != nil {
			c.JSON(http.StatusInternalServerError, model.Response{
				Code:    500,
				Message: "Config created but GitLab submission failed: " + err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "success",
		Data: map[string]interface{}{
			"config": req.Config,
			"mrURL":  mrURL,
		},
	})
}

type UpdateConfigRequest struct {
	Config   *model.Config `json:"config"`
	Operator string        `json:"operator"`
	Reason   string        `json:"reason"`
	SubmitToGitLab bool    `json:"submitToGitLab"`
}

func (h *ConfigHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	err := h.configService.Update(id, req.Config, req.Operator, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	var mrURL string
	if req.SubmitToGitLab && h.gitlabMgr != nil {
		mrURL, err = h.configService.SubmitToGitLab(req.Config, h.gitlabMgr, req.Operator, req.Reason)
		if err != nil {
			c.JSON(http.StatusInternalServerError, model.Response{
				Code:    500,
				Message: "Config updated but GitLab submission failed: " + err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "success",
		Data: map[string]interface{}{
			"config": req.Config,
			"mrURL":  mrURL,
		},
	})
}

type DeleteConfigRequest struct {
	Operator string `json:"operator"`
	Reason   string `json:"reason"`
}

func (h *ConfigHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	var req DeleteConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	err := h.configService.Delete(id, req.Operator, req.Reason)
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

func (h *ConfigHandler) GetHistory(c *gin.Context) {
	configID := c.Param("id")

	history, err := h.configService.GetHistory(configID)
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
		Data:    history,
	})
}

func (h *ConfigHandler) Compare(c *gin.Context) {
	id1 := c.Query("id1")
	id2 := c.Query("id2")

	if id1 == "" || id2 == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "id1 and id2 are required",
		})
		return
	}

	result, err := h.configService.CompareHistory(id1, id2)
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
		Data:    result,
	})
}
