package service

import (
	"context"
	"discord-bot-service/internal/models"
	"discord-bot-service/internal/repository/mongodb"
)

type FlowDataService struct {
	repo mongodb.FlowDataRepository
}

func NewFlowDataService(repo *mongodb.Repository) *FlowDataService {
	return &FlowDataService{repo: repo.FlowData}
}

func (s *FlowDataService) SaveFlowData(ctx context.Context, flowData *models.FlowData) error {
	return s.repo.Save(ctx, flowData)
}

func (s *FlowDataService) GetFlowData(ctx context.Context, key string) (*models.FlowData, error) {
	return s.repo.GetByKey(ctx, key)
}

func (s *FlowDataService) GetAllFlowData(ctx context.Context) ([]models.FlowData, error) {
	return s.repo.GetAll(ctx)
}

func (s *FlowDataService) DeleteFlowData(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
