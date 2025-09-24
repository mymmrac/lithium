package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage interface {
	Upload(ctx context.Context, bucket, path string, data []byte) error
	Download(ctx context.Context, bucket, path string) ([]byte, error)
}

type storage struct {
	client *minio.Client
}

func NewStorage(cfg Config) (Storage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.ID, cfg.Secret, ""),
		Secure: cfg.Secure,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio: %w", err)
	}

	return &storage{
		client: client,
	}, nil
}

func (s *storage) Upload(ctx context.Context, bucket, path string, data []byte) error {
	_, err := s.client.PutObject(ctx, bucket, path, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{})
	return err
}

func (s *storage) Download(ctx context.Context, bucket, path string) ([]byte, error) {
	obj, err := s.client.GetObject(ctx, bucket, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}
	defer func() { _ = obj.Close() }()

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("read object: %w", err)
	}

	return data, nil
}
