// infrastructure/storage/cloudinary/cloudinary_storage.go
package cloudinary

import (
	"context"
	"mime/multipart"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/thizplus/gofiber-chat-api/domain/service"
)

// cloudinaryStorage จัดการการเก็บไฟล์ด้วย Cloudinary
type cloudinaryStorage struct {
	cld    *cloudinary.Cloudinary
	ctx    context.Context
	config *CloudinaryConfig
}

// NewCloudinaryStorage สร้าง FileStorageService ที่ใช้ Cloudinary
func NewCloudinaryStorage(config *CloudinaryConfig) (service.FileStorageService, error) {
	// สร้าง context
	ctx := context.Background()

	// สร้าง Cloudinary client
	cld, err := cloudinary.NewFromParams(config.CloudName, config.APIKey, config.APISecret)
	if err != nil {
		return nil, err
	}

	return &cloudinaryStorage{
		cld:    cld,
		ctx:    ctx,
		config: config,
	}, nil
}

// ใช้ฟังก์ชันช่วยสร้าง pointer to bool
func boolPtr(b bool) *bool {
	return &b
}

// UploadImage อัปโหลดรูปภาพไปยัง Cloudinary
func (c *cloudinaryStorage) UploadImage(file *multipart.FileHeader, folder string) (*service.FileUploadResult, error) {
	// เปิดไฟล์
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// กำหนดตัวเลือกในการอัปโหลด
	uploadParams := uploader.UploadParams{
		Folder:         folder,
		UseFilename:    boolPtr(true),
		UniqueFilename: boolPtr(true),
		ResourceType:   "image",
		Transformation: "q_auto:good",
	}

	// อัปโหลดไปยัง Cloudinary
	ctx, cancel := context.WithTimeout(c.ctx, 30*time.Second)
	defer cancel()

	result, err := c.cld.Upload.Upload(ctx, src, uploadParams)
	if err != nil {
		return nil, err
	}

	// แปลงผลลัพธ์เป็น domain model
	return &service.FileUploadResult{
		URL:          result.SecureURL,
		PublicID:     result.PublicID,
		ResourceType: result.ResourceType,
		Format:       result.Format,
		Size:         int(result.Bytes),
	}, nil
}

// UploadFile อัปโหลดไฟล์ทั่วไปไปยัง Cloudinary
func (c *cloudinaryStorage) UploadFile(file *multipart.FileHeader, folder string) (*service.FileUploadResult, error) {
	// เปิดไฟล์
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// กำหนดตัวเลือกในการอัปโหลด
	uploadParams := uploader.UploadParams{
		Folder:         folder,
		UseFilename:    boolPtr(true),
		UniqueFilename: boolPtr(true),
		ResourceType:   "auto",
	}

	// อัปโหลดไปยัง Cloudinary
	ctx, cancel := context.WithTimeout(c.ctx, 30*time.Second)
	defer cancel()

	result, err := c.cld.Upload.Upload(ctx, src, uploadParams)
	if err != nil {
		return nil, err
	}

	// แปลงผลลัพธ์เป็น domain model
	return &service.FileUploadResult{
		URL:          result.SecureURL,
		PublicID:     result.PublicID,
		ResourceType: result.ResourceType,
		Format:       result.Format,
		Size:         int(result.Bytes),
	}, nil
}
