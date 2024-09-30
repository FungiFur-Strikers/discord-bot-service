package api

import (
	"discord-bot-service/internal/models"
	"discord-bot-service/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FlowDataHandler struct {
	service *service.FlowDataService
}

func NewFlowDataHandler(service *service.FlowDataService) *FlowDataHandler {
	return &FlowDataHandler{service: service}
}

func (h *FlowDataHandler) SaveFlowData(c *gin.Context) {
	var flowData models.FlowData
	if err := c.ShouldBindJSON(&flowData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.SaveFlowData(c.Request.Context(), &flowData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, flowData)
}

func (h *FlowDataHandler) GetFlowData(c *gin.Context) {
	id := c.Param("id")
	flowData, err := h.service.GetFlowData(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, flowData)
}

func (h *FlowDataHandler) GetAllFlowData(c *gin.Context) {
	flowDataList, err := h.service.GetAllFlowData(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, flowDataList)
}

func (h *FlowDataHandler) DeleteFlowData(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteFlowData(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func SetupFlowDataRoutes(r *gin.Engine, service *service.FlowDataService) {
	handler := NewFlowDataHandler(service)

	r.POST("/flow-data", handler.SaveFlowData)
	r.GET("/flow-data/:id", handler.GetFlowData)
	r.GET("/flow-data", handler.GetAllFlowData)
	r.DELETE("/flow-data/:id", handler.DeleteFlowData)
}
