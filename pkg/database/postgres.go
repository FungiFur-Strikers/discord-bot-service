package database

import (
	"discord-bot-service/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgresDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto Migrate the schema
	err = db.AutoMigrate(&models.Bot{}, &models.Guild{}, &models.Channel{}, &models.NodeDify{}, &models.FlowData{})

	if err != nil {
		return nil, err
	}

	return db, nil
}
