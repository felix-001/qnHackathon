package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type WebHandler struct{}

func NewWebHandler() *WebHandler {
	return &WebHandler{}
}

func (h *WebHandler) Index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func (h *WebHandler) Projects(c *gin.Context) {
	c.HTML(http.StatusOK, "projects.html", nil)
}

func (h *WebHandler) Releases(c *gin.Context) {
	c.HTML(http.StatusOK, "releases.html", nil)
}

func (h *WebHandler) Monitoring(c *gin.Context) {
	c.HTML(http.StatusOK, "monitoring.html", nil)
}

func (h *WebHandler) Config(c *gin.Context) {
	c.HTML(http.StatusOK, "config.html", nil)
}
