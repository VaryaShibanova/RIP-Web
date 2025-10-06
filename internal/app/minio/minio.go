package minio

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewMinioClient(endpoint, accessKey, secretKey string, useSSL bool) (*minio.Client, error) {
	return minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
}

func UploadImage(ctx context.Context, client *minio.Client, bucket string, file *multipart.FileHeader, anomalyID uint) (string, error) {
	f, err := file.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	contentType := file.Header.Get("Content-Type")
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		switch contentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		default:
			ext = ".bin"
		}
	}

	// Генерируем название на латинице
	timestamp := time.Now().Unix()
	objectName := fmt.Sprintf("anomaly_%d_%d%s", anomalyID, timestamp, ext)

	_, err = client.PutObject(ctx, bucket, objectName, f, file.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}
	return objectName, nil
}

func DeleteObject(ctx context.Context, client *minio.Client, bucket, objectName string) error {
	return client.RemoveObject(ctx, bucket, objectName, minio.RemoveObjectOptions{})
}

func InitMinio() (*minio.Client, error) {
	host := os.Getenv("MINIO_HOST")
	port := os.Getenv("MINIO_PORT")
	user := os.Getenv("MINIO_USER")
	pass := os.Getenv("MINIO_PASS")

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "9000"
	}
	if user == "" {
		user = "minio"
	}
	if pass == "" {
		pass = "minio124"
	}

	mc, err := NewMinioClient(host+":"+port, user, pass, false)
	if err != nil {
		return nil, err
	}

	// Создаем бакет если не существует
	bucketName := "images"
	exists, err := mc.BucketExists(context.Background(), bucketName)
	if err == nil && !exists {
		err = mc.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
	}

	return mc, nil
}

func GetImageURL(objectName string) string {
	host := os.Getenv("MINIO_HOST")
	port := os.Getenv("MINIO_PORT")

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "9000"
	}

	return fmt.Sprintf("http://%s:%s/images/%s", host, port, objectName)
}

func ExtractObjectNameFromURL(url string) string {
	parts := strings.Split(url, "/")
	return parts[len(parts)-1]
}

// minio.go - функция с кастомным названием
func UploadImageWithName(ctx context.Context, client *minio.Client, bucket string, file *multipart.FileHeader, anomalyID uint, customName string) (string, error) {
	f, err := file.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	contentType := file.Header.Get("Content-Type")
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		switch contentType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		default:
			ext = ".bin"
		}
	}

	// ИСПОЛЬЗУЕМ кастомное название + ID для уникальности
	objectName := fmt.Sprintf("img/%s_%d%s", customName, anomalyID, ext)

	_, err = client.PutObject(ctx, bucket, objectName, f, file.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}
	return objectName, nil
}
