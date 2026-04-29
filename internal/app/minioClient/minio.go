package minioClient

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewMinioClient(endpoint, accessKey, secretKey string, useSSL bool) (*minio.Client, error) {
	return minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
}

func InitMinio() (*minio.Client, error) {
	host := os.Getenv("MINIO_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("MINIO_PORT")
	if port == "" {
		port = "9000"
	}
	user := os.Getenv("MINIO_USER")
	if user == "" {
		// Default credentials match docker-compose.yml (MINIO_ROOT_USER/PASSWORD).
		user = "rootminio"
	}
	pass := os.Getenv("MINIO_PASS")
	if pass == "" {
		pass = "rootminio"
	}
	return NewMinioClient(host+":"+port, user, pass, false)
}

func Bucket() string {
	b := os.Getenv("MINIO_BUCKET")
	if b == "" {
		return "sql-index"
	}
	return b
}

func UploadFromReader(ctx context.Context, client *minio.Client, bucket, objectName string, r io.Reader, size int64, contentType string) (minio.UploadInfo, error) {
	return client.PutObject(ctx, bucket, objectName, r, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
}

func UploadFile(ctx context.Context, client *minio.Client, bucket, prefix string, file *multipart.FileHeader) (string, error) {
	f, err := file.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	contentType := file.Header.Get("Content-Type")
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		switch {
		case strings.HasPrefix(contentType, "image/"):
			ext = ".png"
		case strings.HasPrefix(contentType, "video/"):
			ext = ".mp4"
		default:
			ext = ".bin"
		}
	}

	objectName := fmt.Sprintf("%s_%s%s", prefix, uuid.NewString(), ext)
	if _, err := UploadFromReader(ctx, client, bucket, objectName, f, file.Size, contentType); err != nil {
		return "", err
	}
	return objectName, nil
}

