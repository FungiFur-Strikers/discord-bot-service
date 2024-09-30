package mongodb

import (
	"context"
	"discord-bot-service/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type BotRepository struct {
	collection *mongo.Collection
}

func NewBotRepository(db *mongo.Database) BotRepository {
	return BotRepository{
		collection: db.Collection("bots"),
	}
}

func (r *BotRepository) Create(ctx context.Context, bot *models.Bot) error {
	_, err := r.collection.InsertOne(ctx, bot)
	return err
}

func (r *BotRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}

func (r *BotRepository) GetAll(ctx context.Context) ([]models.Bot, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bots []models.Bot
	if err = cursor.All(ctx, &bots); err != nil {
		return nil, err
	}
	return bots, nil
}

func (r *BotRepository) GetByID(ctx context.Context, id string) (*models.Bot, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var bot models.Bot
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&bot)
	if err != nil {
		return nil, err
	}
	return &bot, nil
}
