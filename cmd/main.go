package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"discord-bot-service/bot"
	"discord-bot-service/dify"
	"discord-bot-service/internal/api"
	"discord-bot-service/internal/config"
	"discord-bot-service/internal/repository/mongodb"
	"discord-bot-service/internal/service"
	"discord-bot-service/pkg/database"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
)

var (
	conversationIds = make(map[string]string)
	nodeService     *service.NodeDifyService
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize MongoDB client
	client, err := database.NewMongoDBClient(cfg.MongoDBURI)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	executor := bot.NewFlowExecutor()

	executor.RegisterNodeExecutor("start", startNodeExecutor)
	executor.RegisterNodeExecutor("server", serverNodeExecutor)
	executor.RegisterNodeExecutor("channel", channelNodeExecutor)
	executor.RegisterNodeExecutor("dify", difyNodeExecutor)
	executor.RegisterNodeExecutor("discordReply", discordReplyNodeExecutor)

	// Initialize repository
	db := client.Database(cfg.MongoDBName)
	repo := mongodb.NewRepository(db)

	// Initialize service
	flowService := service.NewFlowDataService(repo)
	botService := service.NewBotService(repo)
	nodeService = service.NewNodeDifyService(repo)

	// Setup Gin router
	router := gin.Default()

	// Setup routes
	api.SetupFlowDataRoutes(router, flowService)
	api.SetupBotRoutes(router, botService, flowService, executor, cfg.MessageServiceURL)
	api.SetupNodeRoutes(router, nodeService)
	// Start server
	log.Printf("Starting server on %s", cfg.ServerAddress)
	if err := router.Run(cfg.ServerAddress); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func addDomain(difyURL string, text string) string {
	// ](/files/tools/ で始まるパスにマッチ
	re := regexp.MustCompile(`\]\(/files/tools/[^\s]*\)`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		return fmt.Sprintf("](%s%s", difyURL, match[2:]) // ドメインを追加して置換
	})
}

func startNodeExecutor(props bot.NodeProps) (bot.NodeResult, error) {
	return bot.NodeResult{
		Type:     "start",
		Continue: true,
	}, nil
}
func serverNodeExecutor(props bot.NodeProps) (bot.NodeResult, error) {
	return bot.NodeResult{
		Type:     "server",
		Continue: props.Session.State.User.ID+"-"+props.Message.GuildID == props.Node.ID,
	}, nil
}
func channelNodeExecutor(props bot.NodeProps) (bot.NodeResult, error) {
	return bot.NodeResult{
		Type:     "channel",
		Continue: props.Session.State.User.ID+"-"+props.Message.ChannelID == props.Node.ID,
	}, nil
}
func difyNodeExecutor(props bot.NodeProps) (bot.NodeResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	botConfig, err := nodeService.GetNodeDifyByName(ctx, props.Node.Data.Label)
	if err != nil {
		return bot.NodeResult{
			Type:     "dify",
			Continue: true,
		}, nil

	}

	cleanContent := strings.ReplaceAll(props.Message.Content, "<@"+props.Session.State.User.ID+">", "")
	cleanContent = strings.TrimSpace(cleanContent)
	conversationId := conversationIds[props.Node.Data.Label+props.Message.ChannelID]
	props.Session.ChannelTyping(props.Message.ChannelID)
	response, err := dify.GenerateMessage(botConfig.Url, botConfig.Token, conversationId, props.Message.ChannelID+"zzxxxMxxzz"+cleanContent)
	if err != nil {
		SendMessage(props.Session, props.Message.ChannelID, err.Error())
		return bot.NodeResult{
			Type:     "dify",
			Continue: false,
		}, nil
	}
	conversationIds[props.Node.Data.Label+props.Message.ChannelID] = response.ConversationID

	SendMessage(props.Session, props.Message.ChannelID, addDomain(botConfig.Url, response.Answer))

	return bot.NodeResult{
		Type:     "dify",
		Continue: true,
	}, nil
}
func discordReplyNodeExecutor(props bot.NodeProps) (bot.NodeResult, error) {
	// エンドノードのロジックを実装
	return bot.NodeResult{
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
