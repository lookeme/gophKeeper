// Package cmd /*
package cmd

import (
	"GophKeeper/internal/proto/gkeeper/pb"
	"GophKeeper/internal/security"
	"GophKeeper/internal/service"
	db "GophKeeper/internal/storage"
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

// LevelMap is a map that associates string keys with zapcore.Level values. It is used to map logging levels from string representations to their corresponding zapcore.Level constants
var (
	LevelMap = map[string]zapcore.Level{
		"debug": zapcore.DebugLevel,
		"info":  zapcore.InfoLevel,
		"warn":  zapcore.WarnLevel,
		"error": zapcore.ErrorLevel,
	}
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "A brief description of your command",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		configFile := viper.GetString("config")
		if configFile != "" {
			viper.SetConfigFile(configFile)
		} else {
			panic(fmt.Errorf("config file not exist"))
		}
		if err := viper.ReadInConfig(); err != nil {
			fmt.Printf(": %s\n", err)
		}
		logLevel := getLoggerLevel()
		config := zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(logLevel) // Устанавливаем уровень на Debug
		logger, err := config.Build()
		if err != nil {
			panic(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			if err := startGRPCServer(ctx, logger); err != nil {
				log.Fatalf("Failed to start gRPC server: %v", err)
			}
		}()
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		logger.Info("Stopping gRPC server...")
		cancel()
	},
}

func startGRPCServer(ctx context.Context, logger *zap.Logger) error {
	connString := viper.GetString("database.postgres.connection_string")
	postgres, err := db.New(ctx, logger, connString)
	if err != nil {
		logger.Fatal("Fatal error occurred",
			zap.String("operation", "database connection"),
			zap.Error(err),
		)
	}
	userRepo := db.NewUserRepository(postgres)
	settingsRepo := db.NewSettingsRepository(postgres)
	credRepo := db.NewCredRepository(postgres)
	storage := db.NewStorage(userRepo, settingsRepo, credRepo)
	endpoint := viper.GetString("blockstore.s3.endpoint")
	accessKey := viper.GetString("blockstore.s3.access_key_id")
	secretKey := viper.GetString("blockstore.s3.secret_access_key")
	bucket := viper.GetString("blockstore.s3.bucket")
	s3service, err := service.NewS3Service(logger, endpoint, accessKey, secretKey, bucket, false)
	if err != nil {
		logger.Fatal("Fatal error occurred",
			zap.String("operation", "s3 service creation"),
			zap.Error(err),
		)
	}
	err = s3service.InitBucket()
	if err != nil {
		panic(err)
	}
	serverAddress := getAddress()
	lis, err := net.Listen("tcp", serverAddress)
	if err != nil {
		logger.Fatal("failed to listen: %v", zap.String("error", err.Error()))
	}
	userService := service.NewUserServiceServer(storage, logger)
	s3service, err = service.NewS3Service(logger, endpoint, accessKey, secretKey, bucket, false)
	if err != nil {
		logger.Fatal("failed to create s3 service: %v", zap.String("error", err.Error()))
	}
	secureService := security.NewSecureService(storage, logger)
	err = secureService.Init(ctx)
	if err != nil {
		logger.Fatal("failed to init secure service: %v", zap.String("error", err.Error()))
	}
	authService := security.NewAuthService(storage, logger)
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(authService.GetAuthInterceptor()),
		grpc.StreamInterceptor(authService.GetAuthStreamInterceptor()))
	credService := service.NewUserCredService(storage, logger)
	fileManagerService := service.NewFileManagerService(s3service, userService, authService, credService, secureService)
	pb.RegisterFileManagerServiceServer(grpcServer, fileManagerService)
	go func() {
		logger.Info("starting credentials rotation ticker...")
		secureService.StartTickerRotation(ctx)
	}()
	go func() {
		<-ctx.Done()
		logger.Info("stopping gRPC server...")
		postgres.Close()
		grpcServer.GracefulStop()
	}()
	logger.Info(" app is starting on ",
		zap.String("port", serverAddress),
		zap.String("version", "1.0.0"),
	)
	return grpcServer.Serve(lis)
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func getLoggerLevel() zapcore.Level {
	val, ok := LevelMap[viper.GetString("logger.level")]
	if !ok {
		return zapcore.InfoLevel
	}
	return val
}

func getAddress() string {
	val := viper.GetString("listen_address")
	return val
}
