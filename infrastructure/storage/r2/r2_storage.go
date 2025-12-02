// infrastructure/storage/r2/r2_storage.go
package r2

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/service"
)

// r2Storage จัดการการเก็บไฟล์ด้วย Cloudflare R2
type r2Storage struct {
	client *s3.Client
	config *R2Config
	ctx    context.Context
}

// NewR2Storage สร้าง FileStorageService ที่ใช้ Cloudflare R2
func NewR2Storage(cfg *R2Config) (service.FileStorageService, error) {
	ctx := context.Background()

	// สร้าง AWS SDK config สำหรับ R2
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:           cfg.GetEndpoint(),
			SigningRegion: cfg.GetRegion(),
		}, nil
	})

	awsConfig, err := config.LoadDefaultConfig(ctx,
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
		config.WithRegion(cfg.GetRegion()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load R2 config: %w", err)
	}

	// สร้าง S3 client
	client := s3.NewFromConfig(awsConfig)

	return &r2Storage{
		client: client,
		config: cfg,
		ctx:    ctx,
	}, nil
}

// UploadImage อัปโหลดรูปภาพไปยัง R2
func (r *r2Storage) UploadImage(file *multipart.FileHeader, folder string) (*service.FileUploadResult, error) {
	return r.uploadFile(file, folder, "image")
}

// UploadFile อัปโหลดไฟล์ทั่วไปไปยัง R2
func (r *r2Storage) UploadFile(file *multipart.FileHeader, folder string) (*service.FileUploadResult, error) {
	return r.uploadFile(file, folder, "auto")
}

// uploadFile ฟังก์ชันช่วยสำหรับอัปโหลดไฟล์
func (r *r2Storage) uploadFile(file *multipart.FileHeader, folder string, resourceType string) (*service.FileUploadResult, error) {
	// เปิดไฟล์
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// อ่านไฟล์เข้า buffer
	buf := new(bytes.Buffer)
	size, err := io.Copy(buf, src)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// สร้าง unique filename
	ext := filepath.Ext(file.Filename)
	nameWithoutExt := strings.TrimSuffix(file.Filename, ext)
	uniqueID := uuid.New().String()[:8]
	filename := fmt.Sprintf("%s_%s%s", nameWithoutExt, uniqueID, ext)

	// สร้าง path
	var path string
	if folder != "" {
		path = filepath.Join(folder, filename)
	} else {
		path = filename
	}
	// แปลง backslash เป็น forward slash สำหรับ S3
	path = filepath.ToSlash(path)

	// กำหนด content type
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// อัปโหลดไปยัง R2
	ctx, cancel := context.WithTimeout(r.ctx, 30*time.Second)
	defer cancel()

	_, err = r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.config.Bucket),
		Key:         aws.String(path),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to R2: %w", err)
	}

	// สร้าง public URL
	publicURL := r.GetPublicURL(path)

	// แปลงผลลัพธ์เป็น domain model
	return &service.FileUploadResult{
		URL:          publicURL,
		Path:         path,
		PublicID:     path, // ใช้ path เป็น PublicID สำหรับ R2
		ResourceType: resourceType,
		Format:       strings.TrimPrefix(ext, "."),
		Size:         int(size),
		Metadata:     map[string]string{},
	}, nil
}

// DeleteFile ลบไฟล์จาก R2
func (r *r2Storage) DeleteFile(path string) error {
	ctx, cancel := context.WithTimeout(r.ctx, 10*time.Second)
	defer cancel()

	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.config.Bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from R2: %w", err)
	}

	return nil
}

// GetPublicURL สร้าง public URL สำหรับไฟล์
func (r *r2Storage) GetPublicURL(path string) string {
	// ใช้ Public URL ที่ตั้งค่าไว้
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(r.config.PublicURL, "/"), path)
}

// GeneratePresignedUploadURL สร้าง presigned URL สำหรับให้ client upload ตรง
func (r *r2Storage) GeneratePresignedUploadURL(path string, contentType string, expiry time.Duration) (*service.PresignedURLResult, error) {
	// สร้าง presigned client
	presignClient := s3.NewPresignClient(r.client)

	// กำหนด expiry time
	expiresAt := time.Now().Add(expiry)

	// สร้าง presigned URL สำหรับ PUT
	ctx, cancel := context.WithTimeout(r.ctx, 5*time.Second)
	defer cancel()

	presignedReq, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.config.Bucket),
		Key:         aws.String(path),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(expiry))

	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}

	return &service.PresignedURLResult{
		URL:       presignedReq.URL,
		Path:      path,
		ExpiresAt: expiresAt,
		Method:    presignedReq.Method,
		Fields:    map[string]string{},
	}, nil
}

// GeneratePresignedDownloadURL สร้าง presigned URL สำหรับ download ไฟล์
func (r *r2Storage) GeneratePresignedDownloadURL(path string, expiry time.Duration) (string, error) {
	// สร้าง presigned client
	presignClient := s3.NewPresignClient(r.client)

	// สร้าง presigned URL สำหรับ GET
	ctx, cancel := context.WithTimeout(r.ctx, 5*time.Second)
	defer cancel()

	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.config.Bucket),
		Key:    aws.String(path),
	}, s3.WithPresignExpires(expiry))

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned download URL: %w", err)
	}

	return presignedReq.URL, nil
}
