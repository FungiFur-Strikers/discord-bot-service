package mongodb

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type Repository struct {
	db       *mongo.Database
	FlowData FlowDataRepository
	Bot      BotRepository
	NodeDify NodeDifyRepository
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{
		db:       db,
		FlowData: NewFlowDataRepository(db),
		NodeDify: NewNodeDifyRepository(db),
		Bot:      NewBotRepository(db),
	}
}
