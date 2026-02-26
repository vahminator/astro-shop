package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sellflow/backend/internal/config"
)

type StorageService struct {
	client *minio.Client
	bucket string
}

func NewStorageService(cfg *config.Config) (*StorageService, error) {
	client, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.MinioBucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.MinioBucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &StorageService{client: client, bucket: cfg.MinioBucket}, nil
}

func (s *StorageService) UploadFromURL(ctx context.Context, imageURL string) (storageKey string, objectURL string, err error) {
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	ext := ".jpg"
	if strings.Contains(contentType, "png") {
		ext = ".png"
	} else if strings.Contains(contentType, "webp") {
		ext = ".webp"
	} else if strings.Contains(contentType, "gif") {
		ext = ".gif"
	}

	key := fmt.Sprintf("products/%s%s", uuid.New().String(), ext)

	_, err = s.client.PutObject(ctx, s.bucket, key, resp.Body, resp.ContentLength, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to upload to storage: %w", err)
	}

	objectURL = fmt.Sprintf("http://%s/%s/%s", s.client.EndpointURL().Host, s.bucket, key)
	return key, objectURL, nil
}

func (s *StorageService) Upload(ctx context.Context, reader io.Reader, size int64, contentType string) (storageKey string, objectURL string, err error) {
	ext := ".jpg"
	if strings.Contains(contentType, "png") {
		ext = ".png"
	} else if strings.Contains(contentType, "webp") {
		ext = ".webp"
	} else if strings.Contains(contentType, "gif") {
		ext = ".gif"
	}

	key := fmt.Sprintf("products/%s%s", uuid.New().String(), ext)

	_, err = s.client.PutObject(ctx, s.bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to upload to storage: %w", err)
	}

	objectURL = fmt.Sprintf("http://%s/%s/%s", s.client.EndpointURL().Host, s.bucket, key)
	return key, objectURL, nil
}

func (s *StorageService) Delete(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}

func (s *StorageService) GetURL(key string) string {
	return fmt.Sprintf("http://%s/%s/%s", s.client.EndpointURL().Host, s.bucket, key)
}
