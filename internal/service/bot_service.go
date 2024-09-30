package service

import (
	"context"
	"discord-bot-service/internal/models"
	"discord-bot-service/internal/repository/mongodb"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type BotService struct {
	repo mongodb.BotRepository
}

func NewBotService(repo *mongodb.Repository) *BotService {
	return &BotService{
		repo: repo.Bot,
	}
}

func (bs *BotService) GetAllBots(ctx context.Context) ([]models.Bot, error) {
	return bs.repo.GetAll(ctx)
}

func (bs *BotService) AddBot(ctx context.Context, token string) (*models.Bot, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}
	defer dg.Close()

	// ボット情報を取得
	user, err := dg.User("@me")
	if err != nil {
		return nil, err
	}

	avatarURL := fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", user.ID, user.Avatar)

	bot := &models.Bot{
		ID:     user.ID,
		Name:   user.Username,
		Avatar: avatarURL,
		Token:  token,
		Guilds: []models.Guild{},
	}

	guilds, err := dg.UserGuilds(100, "", "", true)
	if err != nil {
		return nil, err
	}

	for _, g := range guilds {
		iconURL := discordgo.EndpointGuildIcon(g.ID, g.Icon)
		guild := models.Guild{
			ID:       g.ID,
			Name:     g.Name,
			Icon:     iconURL,
			Channels: []models.Channel{},
		}

		channels, err := dg.GuildChannels(g.ID)
		if err != nil {
			continue
		}

		for _, ch := range channels {
			guild.Channels = append(guild.Channels, models.Channel{
				ID:   ch.ID,
				Name: ch.Name,
			})
		}

		bot.Guilds = append(bot.Guilds, guild)
	}

	// データベースにボットを保存
	if err := bs.repo.Create(ctx, bot); err != nil {
		return nil, err
	}

	return bot, nil
}

func (bs *BotService) DeleteBot(ctx context.Context, id string) error {
	return bs.repo.Delete(ctx, id)
}
