package models

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Bot struct {
	ID     string  `bson:"id" json:"id"`
	Name   string  `bson:"name" json:"name"`
	Avatar string  `bson:"avatar" json:"avatar"`
	Token  string  `bson:"token" json:"-"`
	Guilds []Guild `bson:"guilds" json:"guilds"`
}

type Guild struct {
	ID       string    `bson:"id" json:"id"`
	Name     string    `bson:"name" json:"name"`
	Icon     string    `bson:"icon" json:"icon"`
	Channels []Channel `bson:"channels" json:"channels"`
}

type Channel struct {
	ID   string `bson:"id" json:"id"`
	Name string `bson:"name" json:"name"`
}

type NodeDify struct {
	ID    primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name  string             `bson:"name" json:"name"`
	Token string             `bson:"token" json:"-"`
	Url   string             `bson:"url" json:"url"`
}

type FlowData struct {
	ID    primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Key   string             `bson:"key" json:"key"`
	Edges []Edge             `bson:"edges" json:"edges"`
	Nodes []Node             `bson:"nodes" json:"nodes"`
}

func EnsureIndexes(ctx context.Context, collection *mongo.Collection) error {
	_, err := collection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "key", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	return err
}

type Edge struct {
	ID        string `bson:"id" json:"id"`
	Source    string `bson:"source" json:"source"`
	Target    string `bson:"target" json:"target"`
	Deletable bool   `bson:"deletable" json:"deletable"`
}

type Node struct {
	ID       string       `bson:"id" json:"id"`
	Type     string       `bson:"type" json:"type"`
	Data     NodeData     `bson:"data" json:"data"`
	Position NodePosition `bson:"NodePosition" json:"position"`
	Measured interface{}  `bson:"measured" json:"measured"`
}

type NodeData struct {
	Label string `bson:"label" json:"label"`
}

type NodePosition struct {
	X int `bson:"x" json:"x"`
	Y int `bson:"y" json:"y"`
}
