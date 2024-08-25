package bot

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"backend/flow"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
)

type BotManager struct {
	bots         map[string]*discordgo.Session
	botConfigs   map[string]Bot
	flowStore    flow.Store
	flowExecutor *flow.FlowExecutor
	mu           sync.RWMutex
	ApiURL       string
}

func NewBotManager(flowStore flow.Store, flowExecutor *flow.FlowExecutor, apiURL string) *BotManager {
	return &BotManager{
		bots:         make(map[string]*discordgo.Session),
		botConfigs:   make(map[string]Bot),
		flowStore:    flowStore,
		flowExecutor: flowExecutor,
		ApiURL:       apiURL,
	}
}

func (bm *BotManager) GetBotList(c *gin.Context) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	botList := make([]Bot, 0, len(bm.botConfigs))
	for _, bot := range bm.botConfigs {
		botList = append(botList, bot)
	}

	c.JSON(200, botList)
}

func (bm *BotManager) AddBot(c *gin.Context) {
	var input BotInput

	// 受信したデータをログに出力
	body, _ := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	log.Printf("Received data: %s", string(body))

	// JSONデコードを手動で行い、エラーを詳細に確認
	err := json.Unmarshal(body, &input)
	if err != nil {
		log.Printf("JSON Unmarshal error: %v", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("Gin binding error: %v", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	dg, err := discordgo.New("Bot " + input.Token)
	if err != nil {
		println(0)
		println(err)
		return
	}

	// ボット情報を取得
	user, err := dg.User("@me")
	if err != nil {
		println(1)
		println(err)
		return
	}

	bot := Bot{
		ID:     user.ID,
		Name:   user.Username,
		Avatar: user.AvatarURL(""),
		Guilds: make(map[string]Guild),
	}

	guilds, err := dg.UserGuilds(100, "", "", true)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get guilds"})
		return
	}
	for _, g := range guilds {
		iconURL := discordgo.EndpointGuildIcon(g.ID, g.Icon)
		guild := Guild{
			Name:     g.Name,
			ID:       g.ID,
			Icon:     iconURL,
			Channels: make(map[string]Channel),
		}

		channels, err := dg.GuildChannels(g.ID)
		if err != nil {
			log.Printf("Failed to get channels for guild %s: %v", g.ID, err)
			continue
		}

		for _, ch := range channels {
			guild.Channels[ch.ID] = Channel{
				ID:   ch.ID,
				Name: ch.Name,
			}
		}

		bot.Guilds[g.ID] = guild
	}

	// メッセージハンドラを設定
	dg.AddHandler(bm.handleMessage)
	println(3)

	err = dg.Open()
	if err != nil {
		println(2)

		println(err)
		return
	}

	bm.mu.Lock()
	bm.bots[bot.ID] = dg
	bm.botConfigs[bot.ID] = bot
	bm.mu.Unlock()

	c.JSON(200, bot)

}

func (bm *BotManager) RemoveBot(botID string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if dg, exists := bm.bots[botID]; exists {
		err := dg.Close()
		if err != nil {
			return err
		}
		delete(bm.bots, botID)
		delete(bm.botConfigs, botID)
	}

	return nil
}

func (bm *BotManager) RestartAllBots() error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	for botID, dg := range bm.bots {
		err := dg.Close()
		if err != nil {
			return err
		}

		newDg, err := discordgo.New("Bot " + dg.Token)
		if err != nil {
			return err
		}

		newDg.AddHandler(bm.handleMessage)

		err = newDg.Open()
		if err != nil {
			return err
		}

		bm.bots[botID] = newDg
	}

	return nil
}

func (bm *BotManager) handleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	bm.messageCreate(s, m)

	if m.Author.ID == s.State.User.ID {
		return
	}

	bm.mu.RLock()
	botConfig, exists := bm.botConfigs[s.State.User.ID]
	bm.mu.RUnlock()

	if !exists {
		return
	}

	// ボットIDを使用してフローを取得
	flowData, ok := bm.flowStore.Get("t" + botConfig.Name)
	if !ok {
		// フローが見つからない場合のエラーハンドリング
		return
	}

	// フローを実行
	_, err := bm.flowExecutor.ExecuteFlow(flowData, m)
	if err != nil {
		// エラーハンドリング（ログ出力など）
		return
	}
	// 結果を使用してDiscordに返信するなどの処理
	// 例: s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("フロー実行結果: %v", results))
}

func (bm *BotManager) messageCreate(_ *discordgo.Session, m *discordgo.MessageCreate) {
	// Create a struct to match the API's expected input
	message := struct {
		Data struct {
			SentAt    time.Time `json:"sent_at"`
			Sender    string    `json:"sender"`
			ChannelID string    `json:"channel_id"`
			Content   string    `json:"content"`
		} `json:"data"`
	}{
		Data: struct {
			SentAt    time.Time `json:"sent_at"`
			Sender    string    `json:"sender"`
			ChannelID string    `json:"channel_id"`
			Content   string    `json:"content"`
		}{
			SentAt:    m.Timestamp,
			Sender:    m.Author.Username,
			ChannelID: m.ChannelID,
			Content:   m.Content,
		},
	}

	// Convert the message to JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshalling message: %v", err)
		return
	}

	// Send the message to the API
	resp, err := http.Post(bm.ApiURL+"/api/messages", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error sending message to API: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Printf("API returned non-201 status code: %d", resp.StatusCode)
	}
}
