package service

import (
	"GophKeeper/internal/models"
	"GophKeeper/internal/proto/gkeeper/pb"
	"GophKeeper/internal/security"
	db "GophKeeper/internal/storage"
	"context"
	"encoding/json"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"strconv"
)

type FileManagerService struct {
	s3Service     *S3Service
	userService   *UserServiceServer
	authService   *security.AuthService
	credService   *UserCredService
	secretService *security.SecureService
	pb.UnimplementedFileManagerServiceServer
}

func NewFileManagerService(s3Service *S3Service, userService *UserServiceServer, authService *security.AuthService,
	credService *UserCredService,
	secretService *security.SecureService) *FileManagerService {
	return &FileManagerService{
		s3Service:     s3Service,
		userService:   userService,
		authService:   authService,
		credService:   credService,
		secretService: secretService,
	}
}

func (s *FileManagerService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	name := req.GetUsername()
	password := req.GetPassword()
	email := req.GetEmail()
	usr, err := s.userService.CreateUsr(ctx, name, password, email)
	if err != nil {
		return nil, err
	}
	return &pb.CreateUserResponse{
		Id:       usr.ID,
		Username: usr.Username,
	}, nil

}

func (s *FileManagerService) ListUserFiles(ctx context.Context, _ *emptypb.Empty) (*pb.ListUserFileResponse, error) {
	userID, ok := ctx.Value(security.UserIDKey).(string)
	if !ok {
		return nil, status.Error(codes.Internal, "userID not found in context")
	}
	return s.s3Service.ListFiles(userID)

}

func (s *FileManagerService) DownloadFile(req *pb.DownloadRequest, stream pb.FileManagerService_DownloadFileServer) error {
	ctx := stream.Context()
	userID, ok := ctx.Value(security.UserIDKey).(string)
	if !ok {
		return status.Error(codes.Internal, "userID not found in context")
	}
	return s.s3Service.DownloadFile(ctx, userID, req, stream)
}
func (s *FileManagerService) UploadFile(stream pb.FileManagerService_UploadFileServer) error {
	ctx := stream.Context()
	userID, ok := ctx.Value(security.UserIDKey).(string)
	if !ok {
		return status.Error(codes.Internal, "userID not found in context")
	}
	return s.s3Service.UploadFile(stream, userID)
}

func (s *FileManagerService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	userName := req.GetUsername()
	pass := req.GetPassword()
	token, err := s.authService.Login(ctx, userName, pass)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	header := metadata.Pairs("authorization", token)
	err = grpc.SendHeader(ctx, header)
	if err != nil {
		return nil, status.Error(codes.Internal, "Error sending header")
	}
	return &pb.LoginResponse{
		Message: "Login successful",
	}, nil
}

func (s *FileManagerService) SaveCredentials(ctx context.Context, req *pb.SaveCredentialsRequest) (*pb.SaveCredentialsResponse, error) {
	userID, ok := ctx.Value(security.UserIDKey).(string)
	if !ok {
		return nil, status.Error(codes.Internal, "userID not found in context")
	}
	username := req.GetUsername()
	password := req.GetPassword()
	name := req.GetName()
	if username == "" || password == "" || name == "" {
		return nil, status.Error(codes.InvalidArgument, "parameters are empty")
	}
	data := models.CredData{
		Password: password,
		Username: username,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	stringData := string(jsonData)
	encryptData, err := s.secretService.EncryptData([]byte(stringData))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	_, err = s.credService.SaveCreds(ctx, userID, username, encryptData, db.Credentials)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.SaveCredentialsResponse{
		Message: "Credentials saved",
	}, nil

}

func (s *FileManagerService) GetAllCreds(ctx context.Context, _ *emptypb.Empty) (*pb.AllCredsResponse, error) {
	userID, ok := ctx.Value(security.UserIDKey).(string)
	if !ok {
		return nil, status.Error(codes.Internal, "userID not found in context")
	}
	creds, err := s.credService.GetAllCreds(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var userCreds []*pb.GetCredentialsResponse
	for _, cred := range creds {
		decryptedCred, err := s.secretService.DecryptData(cred.Data)
		if err != nil {
			continue
		}
		userCreds = append(userCreds, &pb.GetCredentialsResponse{
			Name:       cred.Name,
			Version:    strconv.FormatInt(cred.Version, 10),
			Data:       string(decryptedCred),
			CreateDate: cred.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return &pb.AllCredsResponse{
		Creds: userCreds,
	}, nil
}
