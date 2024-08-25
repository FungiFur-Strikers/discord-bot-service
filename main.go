package main

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"backend/bot"
	"backend/dify"
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
	difyClients     = make(map[string]*Dify)
	mu              sync.RWMutex
	conversationIds = make(map[string]string)
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

func startNodeExecutor(props flow.NodeProps) (flow.NodeResult, error) {
	return flow.NodeResult{
		Type:     "start",
		Continue: true,
	}, nil
}
func serverNodeExecutor(props flow.NodeProps) (flow.NodeResult, error) {
	return flow.NodeResult{
		Type:     "server",
		Continue: props.Message.GuildID == props.Node.ID,
	}, nil
}
func channelNodeExecutor(props flow.NodeProps) (flow.NodeResult, error) {

	return flow.NodeResult{
		Type:     "channel",
		Continue: props.Message.ChannelID == props.Node.ID,
	}, nil
}
func difyNodeExecutor(props flow.NodeProps) (flow.NodeResult, error) {
	log.Printf("??: %v", props.Node.Data.Label)
	botConfig := difyClients[props.Node.Data.Label]
	log.Printf("dify: %v", botConfig)
	if botConfig == nil {
		log.Printf("??: %v", props.Node.Data.Label+"が見つからない")
		return flow.NodeResult{
			Type:     "dify",
			Continue: true,
		}, nil

	}

	cleanContent := strings.ReplaceAll(props.Message.Content, "<@"+props.Session.State.User.ID+">", "")
	cleanContent = strings.TrimSpace(cleanContent)
	conversationId := conversationIds[props.Message.ChannelID]
	response, err := dify.GenerateMessage(botConfig.Url, botConfig.Token, conversationId, props.Message.ChannelID+"zzxxxMxxzz"+cleanContent)
	if err != nil {
		SendMessage(props.Session, props.Message.ChannelID, err.Error())
		return flow.NodeResult{
			Type:     "dify",
			Continue: false,
		}, nil
	}
	SendMessage(props.Session, props.Message.ChannelID, response.Answer)

	return flow.NodeResult{
		Type:     "dify",
		Continue: true,
	}, nil
}
func discordReplyNodeExecutor(props flow.NodeProps) (flow.NodeResult, error) {
	// エンドノードのロジックを実装
	return flow.NodeResult{
		Type:     "Rep",
		Continue: true,
	}, nil
}

const maxMessageLength = 1000

// Function to split a message into chunks
func SplitMessage(message string) []string {
	if len(message) <= maxMessageLength {
		return []string{message}
	}

	var chunks []string
	var buffer strings.Builder
	words := strings.Split(message, " ")

	for _, word := range words {
		if buffer.Len()+len(word)+1 > maxMessageLength {
			chunks = append(chunks, buffer.String())
			buffer.Reset()
		}
		if buffer.Len() > 0 {
			buffer.WriteString(" ")
		}
		buffer.WriteString(word)
	}

	if buffer.Len() > 0 {
		chunks = append(chunks, buffer.String())
	}

	return chunks
}

// Function to send a message and split it if necessary
func SendMessage(session *discordgo.Session, channelID, message string) {
	chunks := SplitMessage(message)
	for _, chunk := range chunks {
		_, err := session.ChannelMessageSend(channelID, chunk)
		if err != nil {
			fmt.Println("Error sending message:", err)
		}
	}
}
