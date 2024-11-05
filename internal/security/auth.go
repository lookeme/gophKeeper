package security

import (
	"GophKeeper/internal/models"
	db "GophKeeper/internal/storage"
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
	"time"
)

// TokenExp is a constant that represents the expiration time for JWT tokens.
const TokenExp = time.Hour * 3
const LoginFullMethod = "/pb.FileManagerService/Login"
const CreateUserFullMethod = "/pb.FileManagerService/CreateUser"

var TokenExpiredErr = errors.New("token expired")

type contextKey string

const UserIDKey = contextKey("userID")

// AuthService сервер
type AuthService struct {
	storage *db.Storage
	log     *zap.Logger
}

func (auth *AuthService) Login(ctx context.Context, userName string, pass string) (string, error) {
	user, err := auth.storage.UserRepository.FindByName(ctx, userName)
	if err != nil {
		return "", status.Error(codes.Unauthenticated, "User not found")
	}
	if CheckPass(pass, user.Password) {
		token, err := auth.BuildJWTString(user.ID)
		if err != nil {
			return "", status.Error(codes.Internal, "Could not generate token")
		}
		return token, nil
	}
	return "", status.Error(codes.Unauthenticated, "Invalid credentials")
}

// NewAuthService constructs a new instance of Authorization with a user service and logger.
func NewAuthService(storage *db.Storage, logger *zap.Logger) *AuthService {
	return &AuthService{
		storage: storage,
		log:     logger,
	}
}

// BuildJWTString generates a JWT for a specified user ID.
func (auth *AuthService) BuildJWTString(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, models.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		UserID: userID,
	})
	tokenString, err := token.SignedString([]byte(getSecretKeyToken()))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// GetUserID retrieves a user ID from a given JWT.
func GetUserID(tokenString string) (string, error) {
	var claims models.Claims
	_, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(getSecretKeyToken()), nil
	})
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

// VerifyToken is a method that takes a token string as input and verifies its validity using JWT.
// It returns true if the token is valid, otherwise false.
// If there is an error during verification, it logs the error and returns false.
func (auth *AuthService) VerifyToken(tokenString string) (bool, error) {
	claims := &models.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(getSecretKeyToken()), nil
	})
	if err != nil {
		return false, err
	}
	expirationTime, err := claims.GetExpirationTime()
	if err != nil {
		return false, err
	}
	now := time.Now()
	if expirationTime.Before(now) {
		return false, TokenExpiredErr
	}
	return token.Valid, nil
}

// GetAuthInterceptor returns a grpc.UnaryServerInterceptor that enforces authentication for unary methods.
func (auth *AuthService) GetAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if info.FullMethod == LoginFullMethod || info.FullMethod == CreateUserFullMethod {
			return handler(ctx, req)
		}
		token, err := ExtractToken(ctx)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		if strings.Contains(token, "Bearer") {
			tokenArr := strings.Split(token, " ")
			if len(tokenArr) < 2 {
				return nil, status.Error(codes.Unauthenticated, "Invalid token")
			}
			token = tokenArr[1]
		}
		if ok, err := auth.VerifyToken(token); !ok || err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		userID, err := GetUserID(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		ctxWithVal := context.WithValue(ctx, UserIDKey, userID)
		return handler(ctxWithVal, req)
	}
}

// WrappedStream is a custom wrapper around grpc.ServerStream that allows modifying the context
type WrappedStream struct {
	grpc.ServerStream                 // Embedding grpc.ServerStream
	wrappedContext    context.Context // Store the modified context
}

// Context overrides the ServerStream.Context() method to return the modified context
func (w *WrappedStream) Context() context.Context {
	return w.wrappedContext
}

// GetAuthStreamInterceptor returns a grpc.StreamServerInterceptor that enforces authentication for stream methods.
func (auth *AuthService) GetAuthStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		token, err := ExtractToken(ss.Context())
		if err != nil {
			return status.Error(codes.Unauthenticated, err.Error())
		}
		if ok, err := auth.VerifyToken(token); !ok || err != nil {
			return status.Error(codes.Unauthenticated, err.Error())
		}
		userID, err := GetUserID(token)
		if err != nil {
			return status.Error(codes.Unauthenticated, err.Error())
		}
		ctx := ss.Context()
		ctxWithVal := context.WithValue(ctx, UserIDKey, userID)
		wrappedStream := &WrappedStream{
			ServerStream:   ss,
			wrappedContext: ctxWithVal,
		}
		return handler(srv, wrappedStream)
	}
}

func IsExpired(tokenString string) (bool, error) {
	var claims models.Claims
	_, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(getSecretKeyToken()), nil
	})
	if err != nil {
		return true, err
	}
	expirationTime, err := claims.GetExpirationTime()
	if err != nil {
		return true, err
	}
	now := time.Now()
	if expirationTime.Before(now) {
		return true, nil
	}
	return false, nil

}

// ExtractToken retrieves the 'authorization' token from the gRPC metadata in the given context. Returns an error if the token is missing.
func ExtractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("missing metadata")
	}
	tokens := md["authorization"]
	if len(tokens) == 0 {
		return "", fmt.Errorf("missing auth token")
	}
	return tokens[0], nil
}
