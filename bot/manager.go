package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"discord-bot-service/internal/service"

	"github.com/bwmarrin/discordgo"
)

type BotManager struct {
	bots         map[string]*discordgo.Session
	flowExecutor *FlowExecutor
	mu           sync.RWMutex
	ApiURL       string
	lastId       string
	flowService  *service.FlowDataService
	timeout      time.Duration
}

func NewBotManager(flowService *service.FlowDataService, flowExecutor *FlowExecutor, apiURL string) *BotManager {
	return &BotManager{
		bots:         make(map[string]*discordgo.Session),
		flowExecutor: flowExecutor,
		ApiURL:       apiURL,
		flowService:  flowService,
		timeout:      30 * time.Second,
	}
}

func (bm *BotManager) AddBot(token string) error {
	fmt.Println(token)
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		println(err)
		return err
	}

	// ボット情報を取得
	user, err := dg.User("@me")
	if err != nil {
		println(err)
		return err
	}

	// メッセージハンドラを設定
	dg.AddHandler(bm.handleMessage)

	err = dg.Open()
	if err != nil {
		println(err)
		return err
	}

	bm.mu.Lock()
	bm.bots[user.ID] = dg
	bm.mu.Unlock()

	return nil
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
	bm.saveToMessageService(s, m)
	ctx, cancel := context.WithTimeout(context.Background(), bm.timeout)
	defer cancel()
	if m.Author.ID == s.State.User.ID {
		return
	}

	// メンションされてない場合
	if !strings.Contains(m.Content, "<@"+s.State.User.ID+">") {
		return
	}

	// ボットIDを使用してフローを取得
	flowData, err := bm.flowService.GetFlowData(ctx, s.State.User.Username)
	if err != nil {
		// フローが見つからない場合のエラーハンドリング
		println("flow not found")
		return
	}
	// フローを実行
	_, err = bm.flowExecutor.ExecuteFlow(*flowData, m, s)
	if err != nil {
		return
	}

}

// メッセージサーバーにメッセージを保存
func (bm *BotManager) saveToMessageService(s *discordgo.Session, m *discordgo.MessageCreate) {
	// 多重保存を防止
	bm.mu.RLock()
	if bm.lastId == m.ID {
		return
	}
	bm.lastId = m.ID
	bm.mu.RUnlock()

	// メンションをユーザー名に変更->引用の文字を追加->
	cleanMessage := convertMentionsToNames(s, m)
	cleanMessage = addReplyContextToMessage(s, m) + truncateString(cleanMessage, 500)

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
			Content:   cleanMessage,
		},
	}

	// Convert the message to JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshalling message: %v", err)
		return
	}

	log.Printf("sand: %v", message)
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

func convertMentionsToNames(s *discordgo.Session, m *discordgo.MessageCreate) string {
	// メンションを検出する正規表現パターン
	mentionPattern := regexp.MustCompile(`<@!?(\d+)>`)

	// メッセージ内のすべてのメンションを処理
	convertedContent := mentionPattern.ReplaceAllStringFunc(m.Content, func(mention string) string {
		// メンションからユーザーIDを抽出
		userID := strings.Trim(mention, "<@!>")

		// ユーザー情報を取得
		user, err := s.User(userID)
		if err != nil {
			return mention // エラーが発生した場合は元のメンションを返す
		}

		// メンションをユーザー名に置き換え
		return "@" + user.Username
	})

	return convertedContent
}

func addReplyContextToMessage(s *discordgo.Session, m *discordgo.MessageCreate) string {
	// メッセージに返信情報がない場合は空文字を返す
	if m.MessageReference == nil {
		return ""
	}

	// 返信元のメッセージを取得
	referencedMessage, err := s.ChannelMessage(m.MessageReference.ChannelID, m.MessageReference.MessageID)
	if err != nil {
		fmt.Printf("Error fetching referenced message: %v\n", err)
		return ""
	}

	// 返信元のメッセージ内容を取得（長い場合は省略）
	referencedContent := truncateString(referencedMessage.Content, 300)

	// 返信元のユーザー名を取得
	referencedUser, err := s.User(referencedMessage.Author.ID)
	if err != nil {
		fmt.Printf("Error fetching referenced user: %v\n", err)
		return ""
	}

	// 返信コンテキストを作成（改行対応）
	lines := strings.Split(referencedContent, "\n")
	quotedLines := make([]string, len(lines))
	for i, line := range lines {
		quotedLines[i] = "> " + line
	}
	replyContext := fmt.Sprintf("> %s:\n%s\n\n", referencedUser.Username, strings.Join(quotedLines, "\n"))

	return replyContext
}

func truncateString(s string, maxLength int) string {
	if maxLength <= 3 {
		return "..."
	}

	runes := []rune(s)
	if utf8.RuneCountInString(s) <= maxLength {
		return s
	}

	return string(runes[:maxLength-3]) + "..."
}
