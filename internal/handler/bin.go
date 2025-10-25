package handler

import (
	"net/http"

	"github.com/felix-001/qnHackathon/internal/service"
	"github.com/gin-gonic/gin"
)

type BinHandler struct {
	binService *service.BinService
}

func NewBinHandler(binService *service.BinService) *BinHandler {
	return &BinHandler{
		binService: binService,
	}
}

func (h *BinHandler) GetKeepalive(c *gin.Context) {
	nodeID := c.Query("node_id")
	if nodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "node_id parameter required"})
		return
	}

	node, ok := h.binService.GetNode(nodeID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}

	c.JSON(http.StatusOK, node)
}

func (h *BinHandler) PostKeepalive(c *gin.Context) {
	var req struct {
		NodeID          string `json:"node_id" binding:"required"`
		CPUArch         string `json:"cpu_arch"`
		OSRelease       string `json:"os_release"`
		NodeName        string `json:"node_name"`
		BinProxyVersion string `json:"bin_proxy_version"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "node_id required in request body"})
		return
	}

	node := h.binService.RegisterNode(
		req.NodeID,
		req.CPUArch,
		req.OSRelease,
		req.NodeName,
		req.BinProxyVersion,
	)

	c.JSON(http.StatusCreated, gin.H{
		"message": "node registered",
		"node":    node,
	})
}

func (h *BinHandler) GetBin(c *gin.Context) {
	binName := c.Param("bin_name")

	bin, ok := h.binService.GetBin(binName)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "binary not found"})
		return
	}

	c.JSON(http.StatusOK, bin)
}

func (h *BinHandler) PostBin(c *gin.Context) {
	binName := c.Param("bin_name")

	var req struct {
		NodeID    string `json:"node_id" binding:"required"`
		SHA256Sum string `json:"sha256sum" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sha256sum and node_id required"})
		return
	}

	h.binService.UpdateNodeBin(req.NodeID, binName, req.SHA256Sum)

	c.JSON(http.StatusOK, gin.H{
		"message":    "binary version updated for node",
		"node_id":    req.NodeID,
		"bin_name":   binName,
		"sha256sum":  req.SHA256Sum,
	})
}

func (h *BinHandler) PostProgress(c *gin.Context) {
	binName := c.Param("bin_name")

	var req struct {
		NodeName       string `json:"nodeName" binding:"required"`
		TargetHash     string `json:"targetHash" binding:"required"`
		ProcessingTime *int   `json:"processingTime"`
		Status         string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nodeName, targetHash, and status are required"})
		return
	}

	h.binService.RecordProgress(
		req.NodeName,
		binName,
		req.TargetHash,
		req.Status,
		req.ProcessingTime,
	)

	c.JSON(http.StatusOK, gin.H{
		"message":  "progress recorded",
		"nodeName": req.NodeName,
		"binName":  binName,
		"status":   req.Status,
	})
}

func (h *BinHandler) Download(c *gin.Context) {
	binFileName := c.Param("bin_file_name")

	c.JSON(http.StatusOK, gin.H{
		"message":       "download endpoint - implement actual file serving as needed",
		"bin_file_name": binFileName,
		"note":          "This is a mock endpoint. In production, serve actual binary files here.",
	})
}

func (h *BinHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":      "healthy",
		"nodes_count": h.binService.GetNodesCount(),
		"bins_count":  h.binService.GetBinsCount(),
	})
}
