package mongodb

import (
	"context"
	"discord-bot-service/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type NodeDifyRepository struct {
	collection *mongo.Collection
}

func NewNodeDifyRepository(db *mongo.Database) NodeDifyRepository {
	return NodeDifyRepository{
		collection: db.Collection("node_dify"),
	}
}

func (r NodeDifyRepository) Create(ctx context.Context, dify *models.NodeDify) error {
	if dify.ID.IsZero() {
		dify.ID = primitive.NewObjectID()
	}
	_, err := r.collection.InsertOne(ctx, dify)
	return err
}

func (r NodeDifyRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}

func (r NodeDifyRepository) GetAll(ctx context.Context) ([]models.NodeDify, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var difys []models.NodeDify
	if err = cursor.All(ctx, &difys); err != nil {
		return nil, err
	}
	return difys, nil
}

func (r NodeDifyRepository) GetByID(ctx context.Context, id string) (*models.NodeDify, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var dify models.NodeDify
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&dify)
	if err != nil {
		return nil, err
	}
	return &dify, nil
}

func (r *NodeDifyRepository) GetByName(ctx context.Context, name string) (*models.NodeDify, error) {

	var dify models.NodeDify
	err := r.collection.FindOne(ctx, bson.M{"name": name}).Decode(&dify)
	if err == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	}
	return &dify, err

}
