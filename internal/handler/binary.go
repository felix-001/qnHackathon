package handler

import (
	"net/http"

	"github.com/felix-001/qnHackathon/internal/model"
	"github.com/felix-001/qnHackathon/internal/service"
	"github.com/gin-gonic/gin"
)

type BinaryHandler struct {
	service *service.BinaryService
}

func NewBinaryHandler(service *service.BinaryService) *BinaryHandler {
	return &BinaryHandler{
		service: service,
	}
}

func (h *BinaryHandler) GetBinaryHash(c *gin.Context) {
	binName := c.Param("name")

	hash, err := h.service.GetBinaryHash(binName)
	if err != nil {
		c.JSON(http.StatusNotFound, model.Response{
			Code:    1,
			Message: "Binary not found: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    0,
		Message: "success",
		Data: map[string]interface{}{
			"name":    binName,
			"version": "latest",
			"hash":    hash,
		},
	})
}

func (h *BinaryHandler) UpdateBinaryHash(c *gin.Context) {
	binName := c.Param("name")

	var req struct {
		Hash string `json:"hash" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    1,
			Message: err.Error(),
		})
		return
	}

	if err := h.service.UpdateBinaryHash(binName, req.Hash); err != nil {
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

func (h *BinaryHandler) Keepalive(c *gin.Context) {
	nodeName := c.Query("node")
	if nodeName == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    1,
			Message: "Node name is required",
		})
		return
	}

	node, exists := h.service.GetNodeInfo(nodeName)

	if !exists {
		if c.Request.Method == http.MethodPost {
			var nodeInfo model.NodeInfo
			if err := c.ShouldBindJSON(&nodeInfo); err != nil {
				c.JSON(http.StatusBadRequest, model.Response{
					Code:    1,
					Message: err.Error(),
				})
				return
			}

			if err := h.service.UpdateNodeInfo(&nodeInfo); err != nil {
				c.JSON(http.StatusInternalServerError, model.Response{
					Code:    1,
					Message: err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, model.Response{
				Code:    0,
				Message: "success",
				Data:    nodeInfo,
			})
			return
		}

		c.JSON(http.StatusNotFound, model.Response{
			Code:    1,
			Message: "Node not found. Please POST to register.",
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    0,
		Message: "success",
		Data:    node,
	})
}

func (h *BinaryHandler) DownloadBinary(c *gin.Context) {
	binName := c.Param("name")

	binPath := h.service.GetBinaryPath(binName)

	c.File(binPath)
}
