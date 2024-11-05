package service

import (
	"GophKeeper/internal/models"
	"GophKeeper/internal/security"
	db "GophKeeper/internal/storage"
	"GophKeeper/utils"
	"context"
	"github.com/jackc/pgerrcode"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserServiceServer is the server that provides user services
type UserServiceServer struct {
	storage *db.Storage
	logger  *zap.Logger
}

func NewUserServiceServer(storage *db.Storage, logger *zap.Logger) *UserServiceServer {
	return &UserServiceServer{storage: storage, logger: logger}
}

// CreateUsr handles creating a new user
func (s *UserServiceServer) CreateUsr(ctx context.Context, name string, pass string, email string) (*models.UserDTO, error) {
	if name == "" || email == "" || pass == "" {
		return nil, status.Error(codes.InvalidArgument, "please provide username, email and password")
	}
	password, err := security.EncodePass(pass)
	if err != nil {
		return nil, status.Error(codes.Internal, "something went wrong")
	}
	userID, err := s.storage.UserRepository.SaveUser(ctx, name, string(password), email)
	if err != nil {
		code := utils.ErrorCode(err)
		if code == pgerrcode.UniqueViolation {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "something went wrong")
	}
	return &models.UserDTO{
		ID:       userID.String(),
		Username: name,
	}, nil
}
