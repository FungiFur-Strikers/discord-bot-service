package main

import (
	"log"
	"sync"

	"backend/bot"
	"backend/flow"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
)

type Dify struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Token string `json:"token"`
	Url   string `json:"url"`
}

type DifyInput struct {
	Name  string `json:"name" binding:"required"`
	Token string `json:"token" binding:"required"`
	Url   string `json:"url" binding:"required"`
}

var (
	difyClients = make(map[string]*Dify)
	mu          sync.RWMutex
)

func main() {
	r := gin.Default()
	store := flow.NewInMemoryStore()
	flowHandler := flow.NewHandler(store)
	executor := flow.NewFlowExecutor()

	executor.RegisterNodeExecutor("start", startNodeExecutor)
	executor.RegisterNodeExecutor("server", serverNodeExecutor)
	executor.RegisterNodeExecutor("channel", channelNodeExecutor)
	executor.RegisterNodeExecutor("dify", difyNodeExecutor)
	executor.RegisterNodeExecutor("discordReply", discordReplyNodeExecutor)

	botManager := bot.NewBotManager(store, executor, "http://dify.kajidog.com:12203")

	r.POST("/bot", botManager.AddBot)

	r.POST("/bot/restart", func(c *gin.Context) {
		err := botManager.RestartAllBots()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "All bots restarted successfully"})
	})

	r.GET("/bot/list", botManager.GetBotList)
	r.GET("/bot/flow/:id", flowHandler.GetFlow)
	r.POST("/bot/flow", flowHandler.SaveFlow)

	r.GET("/dify/list", getDifyList)
	r.POST("/dify", addDify)

	r.Run(":8080")
}

func getDifyList(c *gin.Context) {
	mu.RLock()
	defer mu.RUnlock()

	clientList := make([]Dify, 0, len(difyClients))
	for _, client := range difyClients {
		clientList = append(clientList, *client)
	}

	c.JSON(200, clientList)
}

func addDify(c *gin.Context) {
	var input DifyInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Here you might want to validate the token and URL
	// For example, make a test request to the Dify API

	dify := &Dify{
		Name:  input.Name,
		Token: input.Token,
		Url:   input.Url,
	}

	mu.Lock()
	difyClients[dify.Name] = dify
	mu.Unlock()

	c.JSON(200, dify)
}

func startNodeExecutor(node flow.Node, m *discordgo.MessageCreate) (flow.NodeResult, error) {
	return flow.NodeResult{
		Type:     "start",
		Continue: true,
	}, nil
}
func serverNodeExecutor(node flow.Node, m *discordgo.MessageCreate) (flow.NodeResult, error) {
	return flow.NodeResult{
		Type:     "server",
		Continue: m.GuildID == node.ID,
	}, nil
}
func channelNodeExecutor(node flow.Node, m *discordgo.MessageCreate) (flow.NodeResult, error) {

	return flow.NodeResult{
		Type:     "channel",
		Continue: m.ChannelID == node.ID,
	}, nil
}
func difyNodeExecutor(node flow.Node, m *discordgo.MessageCreate) (flow.NodeResult, error) {
	// エンドノードのロジックを実装
	println("dify")

	return flow.NodeResult{
		Type:     "dify",
		Continue: true,
	}, nil
}
func discordReplyNodeExecutor(node flow.Node, m *discordgo.MessageCreate) (flow.NodeResult, error) {
	// エンドノードのロジックを実装
	log.Printf("Rep: %v", node.Data)
	return flow.NodeResult{
		Type:     "Rep",
		Continue: true,
	}, nil
}
