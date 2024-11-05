package service

import (
	"GophKeeper/internal/proto/gkeeper/pb"
	"bytes"
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
	"io"
	"log"
	"path/filepath"
)

const MinPartSize = 5 * 1024 * 1024

// S3Service struct holds the MinIO minIOCore
type S3Service struct {
	minIOCore *minio.Core
	log       *zap.Logger
	bucket    string
}

// NewS3Service initializes a new S3 minIOCore
func NewS3Service(log *zap.Logger, endpoint string, accessKey string, secretKey string, bucket string, useSSL bool) (*S3Service, error) {
	minIOCore, err := minio.NewCore(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	return &S3Service{
		minIOCore: minIOCore,
		bucket:    bucket,
		log:       log,
	}, nil
}

func (s *S3Service) InitBucket() error {
	exists, err := s.minIOCore.BucketExists(context.Background(), s.bucket)
	if err != nil {
		return err
	}
	if exists {
		s.log.Info("Bucket already exists", zap.String("name", s.bucket))
	} else {
		err = s.minIOCore.MakeBucket(context.Background(), s.bucket, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
		err = s.minIOCore.SetBucketPolicy(context.Background(), s.bucket, "public-read")
		if err != nil {
			s.log.Error("Bucket policy has been set", zap.String("name", s.bucket), zap.Error(err))
		}
		s.log.Info("Bucket has been created", zap.String("name", s.bucket))
	}
	versioning, err := s.minIOCore.GetBucketVersioning(context.Background(), s.bucket)
	if err != nil {
		return err
	}
	if versioning.Status != minio.Enabled {
		err = s.minIOCore.EnableVersioning(context.Background(), s.bucket)
		if err != nil {
			s.log.Info("Bucket versioning has been enabled", zap.String("name", s.bucket))
		}
	} else {
		s.log.Info("Bucket versioning is enabled", zap.String("name", s.bucket))
	}
	return nil
}

func (s *S3Service) UploadFile(stream pb.FileManagerService_UploadFileServer, userID string) error {
	firstChunk, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("error receiving first chunk: %v", err)
	}
	fileName := firstChunk.GetFilename()
	fileName = fmt.Sprintf("%s/%s", userID, fileName)
	UploadID, err := s.minIOCore.NewMultipartUpload(context.Background(), s.bucket, fileName, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to initialize multipart upload: %v", err)
	}
	var partNumber int
	var parts []minio.CompletePart
	buffer := bytes.NewBuffer(nil) // Accumulate chunks here
	buffer.Write(firstChunk.GetChunk())
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error receiving chunk: %v", err)
		}
		buffer.Write(chunk.GetChunk())
		if buffer.Len() >= MinPartSize {
			partNumber++
			part, err := s.uploadPart(fileName, UploadID, partNumber, buffer.Len(), buffer)
			if err != nil {
				return fmt.Errorf("failed to upload part %d: %v", partNumber, err)
			}
			parts = append(parts, minio.CompletePart{PartNumber: partNumber, ETag: part.ETag})
			buffer.Reset() // Clear buffer for next accumulation
		}
	}

	if buffer.Len() > 0 {
		partNumber++
		part, err := s.uploadPart(fileName, UploadID, partNumber, buffer.Len(), buffer)
		if err != nil {
			return fmt.Errorf("failed to upload final part %d: %v", partNumber, err)
		}
		parts = append(parts, minio.CompletePart{PartNumber: partNumber, ETag: part.ETag})
	}

	_, err = s.minIOCore.CompleteMultipartUpload(context.Background(), s.bucket, fileName, UploadID, parts, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to complete multipart upload: %v", err)
	}

	return stream.SendAndClose(&pb.UploadStatus{
		Success: true,
		Message: "File uploaded successfully!",
	})
}

func (s *S3Service) uploadPart(fileName, uploadID string, partNumber int, size int, data io.Reader) (minio.ObjectPart, error) {
	part, err := s.minIOCore.PutObjectPart(context.Background(), s.bucket, fileName, uploadID, partNumber, data, int64(size), minio.PutObjectPartOptions{})
	if err != nil {
		return minio.ObjectPart{}, err
	}
	return part, nil
}

func (s *S3Service) DownloadFile(ctx context.Context, userID string, req *pb.DownloadRequest, stream pb.FileManagerService_DownloadFileServer) error {
	fileName := req.GetFilename()
	fileName = fmt.Sprintf("%s/%s", userID, fileName)
	reader, _, _, readErr := s.minIOCore.GetObject(ctx, s.bucket, fileName, minio.GetObjectOptions{})
	if readErr != nil {
		return fmt.Errorf("failed to get object from MinIO: %v", readErr)
	}
	defer reader.Close()
	buffer := make([]byte, MinPartSize)
	for {
		n, readErr := reader.Read(buffer)
		if n > 0 {
			if sendErr := stream.Send(&pb.DownloadResponse{Chunk: buffer[:n]}); sendErr != nil {
				return fmt.Errorf("failed to send chunk: %v", sendErr)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("error reading from MinIO object: %v", readErr)
		}
	}
	return nil
}

// DeleteFile deletes a file from S3
func (s *S3Service) DeleteFile(ctx context.Context, objectName string) error {
	err := s.minIOCore.RemoveObject(ctx, s.bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("unable to delete %q from bucket %q: %w", objectName, s.bucket, err)
	}
	log.Printf("Successfully deleted %q from %q\n", objectName, s.bucket)
	return nil
}

// ListFiles lists all files in the S3 bucket
func (s *S3Service) ListFiles(userName string) (*pb.ListUserFileResponse, error) {
	var continuationToken string
	var result []*pb.FileObject
	for {
		listObjectsV2Result, err := s.minIOCore.ListObjectsV2(s.bucket, userName+"/", "", continuationToken, "", 1000)
		if err != nil {
			return &pb.ListUserFileResponse{}, err
		}

		for _, object := range listObjectsV2Result.Contents {
			fileName := filepath.Base(object.Key)
			result = append(result, &pb.FileObject{
				FileName:  fileName,
				Key:       object.Key,
				VersionID: object.VersionID,
				IsLatest:  true,
				Size:      object.Size,
			})
		}
		if !listObjectsV2Result.IsTruncated {
			break
		}
		continuationToken = listObjectsV2Result.NextContinuationToken
	}
	return &pb.ListUserFileResponse{
		Objects: result,
	}, nil
}
