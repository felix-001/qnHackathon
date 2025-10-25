package handler

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/felix-001/qnHackathon/internal/service"
	"github.com/gin-gonic/gin"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type BinHandler struct {
	binService     *service.BinService
	gitlabMgr      *service.GitLabMgr
	releaseService *service.ReleaseService
}

func NewBinHandler(binService *service.BinService) *BinHandler {
	return &BinHandler{
		binService: binService,
	}
}

func (h *BinHandler) SetGitLabMgr(gitlabMgr *service.GitLabMgr) {
	h.gitlabMgr = gitlabMgr
}

func (h *BinHandler) SetReleaseService(releaseService *service.ReleaseService) {
	h.releaseService = releaseService
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
	if h.gitlabMgr == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GitLab manager not initialized"})
		return
	}

	masterBranch := "master"
	file, resp, err := h.gitlabMgr.Client.RepositoryFiles.GetFile(
		h.gitlabMgr.Conf.ProjectID,
		"streamd.json",
		&gitlab.GetFileOptions{
			Ref: &masterBranch,
		})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to fetch streamd.json: %v", err)})
		return
	}
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch streamd.json from GitLab"})
		return
	}

	var streamdData struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal([]byte(file.Content), &streamdData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to parse streamd.json: %v", err)})
		return
	}

	filePath := filepath.Join("downloads", streamdData.Version)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "version file not found in downloads"})
		return
	}

	fileHandle, err := os.Open(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to open file: %v", err)})
		return
	}
	defer fileHandle.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, fileHandle); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to calculate MD5: %v", err)})
		return
	}

	md5Sum := hex.EncodeToString(hash.Sum(nil))

	c.JSON(http.StatusOK, gin.H{
		"version": streamdData.Version,
		"md5":     md5Sum,
	})
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
		ReleaseID      string `json:"releaseId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nodeName, targetHash, and status are required"})
		return
	}

	if h.releaseService != nil && req.ReleaseID != "" {
		release, err := h.releaseService.Get(req.ReleaseID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get release"})
			return
		}
		if release.Status != "approved" {
			c.JSON(http.StatusForbidden, gin.H{"error": "release not approved"})
			return
		}

		if req.Status == "success" {
			if err := h.releaseService.Complete(req.ReleaseID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update release status"})
				return
			}
		}
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

	releaseID := c.Query("releaseId")
	if h.releaseService != nil && releaseID != "" {
		release, err := h.releaseService.Get(releaseID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get release"})
			return
		}
		if release.Status != "approved" {
			c.JSON(http.StatusForbidden, gin.H{"error": "release not approved"})
			return
		}
	}

	filePath := filepath.Join("downloads", binFileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.File(filePath)
}

func (h *BinHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":      "healthy",
		"nodes_count": h.binService.GetNodesCount(),
		"bins_count":  h.binService.GetBinsCount(),
	})
}
