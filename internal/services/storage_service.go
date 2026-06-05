package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"jia/server/internal/repositories"
)

type StorageService struct {
	settingsRepo *repositories.SettingsRepository
	client       *minio.Client
	bucketName   string
	mu           sync.RWMutex
}

func NewStorageService(settingsRepo *repositories.SettingsRepository) *StorageService {
	s := &StorageService{
		settingsRepo: settingsRepo,
	}
	s.ReloadConfig()
	return s
}

func (s *StorageService) ReloadConfig() {
	s.mu.Lock()
	defer s.mu.Unlock()

	var endpoint, bucket, accessKey, secretKey string
	var useSSL bool

	// If setup isn't completed, S3 settings won't be in the DB. Skip client initialization.
	if !s.settingsRepo.IsSetupCompleted() {
		log.Println("S3 storage client bypass: setup is not yet completed")
		return
	}

	err := s.settingsRepo.Get("s3.endpoint", &endpoint)
	if err != nil {
		log.Printf("Storage config reload error (endpoint): %v", err)
		return
	}

	_ = s.settingsRepo.Get("s3.bucket", &bucket)
	_ = s.settingsRepo.Get("s3.access_key", &accessKey)
	_ = s.settingsRepo.Get("s3.secret_key", &secretKey)
	_ = s.settingsRepo.Get("s3.use_ssl", &useSSL)

	s.bucketName = bucket

	if endpoint == "" {
		log.Println("Storage config reload warning: empty endpoint")
		return
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Printf("Failed to initialize MinIO/S3 client: %v", err)
		return
	}

	s.client = client

	// Ensure bucket exists in a separate goroutine to not block startup
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		exists, err := client.BucketExists(ctx, bucket)
		if err != nil {
			log.Printf("Failed to check if S3 bucket exists: %v", err)
			return
		}
		if !exists {
			err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
			if err != nil {
				log.Printf("Failed to create S3 bucket %s: %v", bucket, err)
			} else {
				log.Printf("Successfully created S3 bucket: %s", bucket)
			}
		}
	}()

	log.Println("MinIO/S3 client configuration successfully reloaded")
}

func (s *StorageService) UploadFile(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) (string, error) {
	s.mu.RLock()
	client := s.client
	bucket := s.bucketName
	s.mu.RUnlock()

	if client == nil {
		return "", errors.New("S3 storage client is not configured")
	}

	opts := minio.PutObjectOptions{
		ContentType: contentType,
	}

	_, err := client.PutObject(ctx, bucket, objectKey, reader, size, opts)
	if err != nil {
		return "", fmt.Errorf("failed to upload object to S3: %w", err)
	}

	return objectKey, nil
}

func (s *StorageService) GetPresignedURL(ctx context.Context, objectKey string, expires time.Duration) (string, error) {
	s.mu.RLock()
	client := s.client
	bucket := s.bucketName
	s.mu.RUnlock()

	if client == nil {
		return "", errors.New("S3 storage client is not configured")
	}

	reqParams := make(url.Values)
	presignedURL, err := client.PresignedGetObject(ctx, bucket, objectKey, expires, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL.String(), nil
}

func (s *StorageService) DeleteFile(ctx context.Context, objectKey string) error {
	s.mu.RLock()
	client := s.client
	bucket := s.bucketName
	s.mu.RUnlock()

	if client == nil {
		return errors.New("S3 storage client is not configured")
	}

	err := client.RemoveObject(ctx, bucket, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object from S3: %w", err)
	}

	return nil
}
