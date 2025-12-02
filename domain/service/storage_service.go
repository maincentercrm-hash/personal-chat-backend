// domain/service/storage_service.go
package service

import (
	"mime/multipart"
	"time"
)

// FileUploadResult เก็บผลลัพธ์จากการอัปโหลด
type FileUploadResult struct {
	URL          string            // Public URL ของไฟล์
	Path         string            // Path ของไฟล์ใน storage (สำหรับ delete/manage)
	PublicID     string            // Public ID (for backward compatibility with Cloudinary)
	ResourceType string            // ประเภทของไฟล์ (image, video, raw, auto)
	Format       string            // รูปแบบไฟล์ (jpg, png, mp4, etc.)
	Size         int               // ขนาดไฟล์ (bytes)
	Width        int               // ความกว้าง (สำหรับ image/video)
	Height       int               // ความสูง (สำหรับ image/video)
	Metadata     map[string]string // Metadata เพิ่มเติม
}

// PresignedURLResult เก็บผลลัพธ์จากการสร้าง presigned URL
type PresignedURLResult struct {
	URL        string            // Presigned URL สำหรับ upload/download
	Path       string            // Path ของไฟล์ใน storage
	ExpiresAt  time.Time         // เวลาที่ URL จะหมดอายุ
	Fields     map[string]string // Fields เพิ่มเติมสำหรับ POST upload (for S3/R2)
	Method     string            // HTTP method (GET, POST, PUT)
}

// FileStorageService กำหนด interface สำหรับบริการจัดเก็บไฟล์
// Interface นี้รองรับหลาย storage providers (Cloudinary, R2, S3, GCS, etc.)
type FileStorageService interface {
	// Upload Operations
	UploadImage(file *multipart.FileHeader, folder string) (*FileUploadResult, error)
	UploadFile(file *multipart.FileHeader, folder string) (*FileUploadResult, error)

	// Delete Operations
	DeleteFile(path string) error // ลบไฟล์ตาม path

	// URL Operations
	GetPublicURL(path string) string // แปลง path เป็น public URL
	GeneratePresignedUploadURL(path string, contentType string, expiry time.Duration) (*PresignedURLResult, error) // สร้าง URL สำหรับ client upload ตรง
	GeneratePresignedDownloadURL(path string, expiry time.Duration) (string, error) // สร้าง URL สำหรับ download ไฟล์ private
}
