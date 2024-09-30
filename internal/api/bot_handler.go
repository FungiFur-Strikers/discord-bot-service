package api

import (
	"context"
	"discord-bot-service/bot"
	"discord-bot-service/internal/service"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type BotHandler struct {
	service        *service.BotService
	sessionService *bot.BotManager
}

func NewBotHandler(service *service.BotService, sessionService *bot.BotManager) *BotHandler {
	return &BotHandler{service: service, sessionService: sessionService}
}

func (h *BotHandler) GetAllBots(c *gin.Context) {
	botList, err := h.service.GetAllBots(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, botList)
}

func (h *BotHandler) AddBot(c *gin.Context) {
	var input struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bot, err := h.service.AddBot(c.Request.Context(), input.Token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// BotManagerにボットを追加
	if err := h.sessionService.AddBot(input.Token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start bot session"})
		return
	}

	c.JSON(http.StatusCreated, bot)
}

func (h *BotHandler) DeleteBot(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteBot(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// BotManagerからボットを削除
	h.sessionService.RemoveBot(id)

	c.Status(http.StatusNoContent)
}

func SetupBotRoutes(r *gin.Engine, botService *service.BotService, flowService *service.FlowDataService, nodeExe *bot.FlowExecutor, apiURL string) {
	sessionService := bot.NewBotManager(flowService, nodeExe, apiURL)
	handler := NewBotHandler(botService, sessionService)

	r.GET("/bot", handler.GetAllBots)
	r.POST("/bot", handler.AddBot)
	r.DELETE("/bots/:id", handler.DeleteBot)
	go initializeBots(botService, sessionService)

}

func initializeBots(botService *service.BotService, sessionManager *bot.BotManager) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	bots, err := botService.GetAllBots(ctx)
	if err != nil {
		log.Printf("Failed to get all bots: %v", err)
		return
	}

	for _, bot := range bots {

		err := sessionManager.AddBot(bot.Token)
		if err != nil {
			log.Printf("Failed to add bot %s: %v", bot.ID, err)
		} else {
			log.Printf("Successfully added bot %s", bot.ID)
		}
	}
}
