package mongodb

import (
	"context"
	"discord-bot-service/internal/models"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrDuplicateKey = errors.New("duplicate key error")
	ErrNotFound     = errors.New("document not found")
)

type FlowDataRepository struct {
	collection *mongo.Collection
}

func NewFlowDataRepository(db *mongo.Database) FlowDataRepository {
	return FlowDataRepository{
		collection: db.Collection("flow_data"),
	}
}

func (r FlowDataRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "key", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	return err
}

func (r FlowDataRepository) Save(ctx context.Context, flow *models.FlowData) error {
	data, err := r.GetByKey(ctx, flow.Key)
	if err != nil && err != ErrNotFound {
		return err
	}

	if err == ErrNotFound {
		flow.ID = primitive.NewObjectID()
	} else {
		flow.ID = data.ID
	}

	_, err = r.collection.UpdateOne(ctx,
		bson.M{"key": flow.Key},
		bson.M{"$set": flow},
		options.Update().SetUpsert(true),
	)

	if mongo.IsDuplicateKeyError(err) {
		return ErrDuplicateKey
	}
	return err
}

func (r FlowDataRepository) GetByID(ctx context.Context, id string) (*models.FlowData, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var flow models.FlowData
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&flow)
	if err == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	}
	return &flow, err
}

func (r FlowDataRepository) GetByKey(ctx context.Context, key string) (*models.FlowData, error) {
	var flow models.FlowData
	err := r.collection.FindOne(ctx, bson.M{"key": key}).Decode(&flow)
	if err == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	}
	return &flow, err
}

func (r FlowDataRepository) GetAll(ctx context.Context) ([]models.FlowData, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var flows []models.FlowData
	if err = cursor.All(ctx, &flows); err != nil {
		return nil, err
	}

	return flows, nil
}

func (r FlowDataRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}
