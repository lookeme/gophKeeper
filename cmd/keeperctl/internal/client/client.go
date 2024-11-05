package client

import (
	"GophKeeper/internal/proto/gkeeper/pb"
	"GophKeeper/internal/security"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
)

type FileManagerClient struct {
	CashedToken string
	Client      pb.FileManagerServiceClient
	Close       func() error
}

func NewFMClient(target string) *FileManagerClient {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		log.Fatalf("failed to connect to auth service: %v", err)
	}
	client := pb.NewFileManagerServiceClient(conn)
	return &FileManagerClient{
		Client: client,
		Close: func() error {
			return conn.Close()
		},
	}

}

// GetAuthToken retrieves a token from an authentication service
func (c *FileManagerClient) GetAuthToken(username string, password string) error {
	ctx := context.Background()
	md := metadata.New(nil)
	ctx = metadata.NewOutgoingContext(ctx, md)
	headers := metadata.MD{}
	_, err := c.Client.Login(ctx, &pb.LoginRequest{
		Username: username,
		Password: password,
	}, grpc.Header(&headers))
	if err != nil {
		return fmt.Errorf("failed to login: %v", err)
	}
	if token, ok := headers["authorization"]; ok {
		c.CashedToken = token[0]
		return nil
	} else {
		return fmt.Errorf("authorization header not found in response")
	}
}

func (c *FileManagerClient) CreateUser(username string, password string, email string) (string, error) {
	user, err := c.Client.CreateUser(context.Background(), &pb.CreateUserRequest{
		Username: username,
		Password: password,
		Email:    email,
	})
	if err != nil {
		return "", err
	}
	return user.Username, nil
}

func (c *FileManagerClient) UploadFile(ctx context.Context) (grpc.ClientStreamingClient[pb.FileChunk, pb.UploadStatus], error) {
	return c.Client.UploadFile(ctx)
}

func (c *FileManagerClient) DownloadFile(ctx context.Context, in *pb.DownloadRequest) (grpc.ServerStreamingClient[pb.DownloadResponse], error) {
	return c.Client.DownloadFile(ctx, in)
}
func (c *FileManagerClient) ListUserFiles(ctx context.Context) (*pb.ListUserFileResponse, error) {
	return c.Client.ListUserFiles(ctx, &emptypb.Empty{})
}

//func (c *FileManagerClient) UploadFileByChunks(ctx context.Context) (grpc.ClientStreamingClient[pb.FileChunk, pb.UploadStatus], error) {
//	return c.Client.UploadFileByChunks(ctx)
//}

func (c *FileManagerClient) IsAuthorized() bool {
	expired, err := security.IsExpired(c.CashedToken)
	if err != nil {
		return false
	}
	return c.CashedToken != "" || !expired
}
