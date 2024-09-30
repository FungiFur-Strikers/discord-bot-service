package service

import (
	"context"
	"discord-bot-service/internal/models"
	"discord-bot-service/internal/repository/mongodb"
)

type NodeDifyService struct {
	repo mongodb.NodeDifyRepository
}

func NewNodeDifyService(repo *mongodb.Repository) *NodeDifyService {
	return &NodeDifyService{
		repo: repo.NodeDify,
	}
}

func (s *NodeDifyService) GetAllNodeDifys(ctx context.Context) ([]models.NodeDify, error) {
	return s.repo.GetAll(ctx)
}
func (s *NodeDifyService) GetNodeDifyByName(ctx context.Context, name string) (*models.NodeDify, error) {
	return s.repo.GetByName(ctx, name)
}

func (s *NodeDifyService) AddNodeDify(ctx context.Context, name, token, url string) (*models.NodeDify, error) {
	dify := &models.NodeDify{
		Name:  name,
		Token: token,
		Url:   url,
	}

	if err := s.repo.Create(ctx, dify); err != nil {
		return nil, err
	}

	return dify, nil
}
