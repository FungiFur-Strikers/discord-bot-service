package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerAddress     string
	MongoDBURI        string
	MongoDBName       string
	MessageServiceURL string
	DifyURL           string
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	return &Config{
		ServerAddress:     os.Getenv("SERVER_ADDRESS"),
		MongoDBURI:        os.Getenv("MONGODB_URI"),
		MongoDBName:       os.Getenv("MONGODB_NAME"),
		MessageServiceURL: os.Getenv("DISCORD_MESSAGE_SERVICE_URL"),
		DifyURL:           os.Getenv("DIFY_URL"),
	}, nil
}
