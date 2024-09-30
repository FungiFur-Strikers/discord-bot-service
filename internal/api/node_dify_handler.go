package api

import (
	"discord-bot-service/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type NodeDifyHandler struct {
	service *service.NodeDifyService
}

func NewNodeDifyHandler(service *service.NodeDifyService) *NodeDifyHandler {
	return &NodeDifyHandler{service: service}
}

func (h *NodeDifyHandler) GetAllNodeDifys(c *gin.Context) {
	difyNodes, err := h.service.GetAllNodeDifys(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, difyNodes)
}

func (h *NodeDifyHandler) AddNodeDify(c *gin.Context) {
	var input struct {
		Name  string `json:"name" binding:"required"`
		Token string `json:"token" binding:"required"`
		Url   string `json:"url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dify, err := h.service.AddNodeDify(c.Request.Context(), input.Name, input.Token, input.Url)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dify)
}

func SetupNodeRoutes(r *gin.Engine, service *service.NodeDifyService) {
	handler := NewNodeDifyHandler(service)

	r.GET("/dify/list", handler.GetAllNodeDifys)
	r.POST("/dify", handler.AddNodeDify)
}
