package bot

// Bot represents a Discord bot
type Bot struct {
	ID     string           `json:"id"`
	Name   string           `json:"name"`
	Avatar string           `json:"avatar"`
	Guilds map[string]Guild `json:"guilds"`
}

type BotInput struct {
	Token string `json:"token" binding:"required"`
}

// Guild represents a Discord guild
type Guild struct {
	Name     string             `json:"name"`
	ID       string             `json:"id"`
	Icon     string             `json:"icon"`
	Channels map[string]Channel `json:"channels"`
}

// Channel represents a Discord channel
type Channel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
