package flow

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	store Store
}

func NewHandler(store Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) SaveFlow(c *gin.Context) {
	var requestBody struct {
		Edges []Edge `json:"edges"`
		Nodes []Node `json:"nodes"`
		ID    string `json:"id"`
	}

	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	flowData := FlowData{
		Edges: requestBody.Edges,
		Nodes: requestBody.Nodes,
		ID:    requestBody.ID,
	}

	h.store.Save(requestBody.ID, flowData)
	c.JSON(http.StatusOK, gin.H{"message": "Flow data saved successfully"})
}

func (h *Handler) GetFlow(c *gin.Context) {
	id := c.Param("id")
	data, ok := h.store.Get(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Flow data not found"})
		return
	}
	c.JSON(http.StatusOK, data)
}
