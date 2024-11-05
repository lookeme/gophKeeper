package service

import (
	"GophKeeper/internal/models"
	db "GophKeeper/internal/storage"
	"context"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UserCredService struct {
	storage *db.Storage
	logger  *zap.Logger
}

func NewUserCredService(storage *db.Storage, logger *zap.Logger) *UserCredService {
	return &UserCredService{storage: storage, logger: logger}
}

func (s *UserCredService) GetCreds(ctx context.Context, userID string, credName string) (models.UserCredentials, error) {
	return s.storage.CredRepository.GetLastUserCreds(ctx, userID, credName)
}

func (s *UserCredService) SaveCreds(ctx context.Context, userID string, credName string, data string, dataType db.DataType) (uuid.UUID, error) {
	return s.storage.CredRepository.SaveUserCreds(ctx, credName, userID, data, dataType)
}

func (s *UserCredService) GetAllCreds(ctx context.Context, userID string) ([]models.UserCredentials, error) {
	return s.storage.CredRepository.FindAll(ctx, userID)
}
