// domain/service/storage_service.go
package service

import "mime/multipart"

// FileUploadResult เก็บผลลัพธ์จากการอัปโหลด
type FileUploadResult struct {
	URL          string
	PublicID     string
	ResourceType string
	Format       string
	Size         int
}

// FileStorageService กำหนด interface สำหรับบริการจัดเก็บไฟล์
type FileStorageService interface {
	UploadImage(file *multipart.FileHeader, folder string) (*FileUploadResult, error)
	UploadFile(file *multipart.FileHeader, folder string) (*FileUploadResult, error)
	// เพิ่มเมธอดอื่นๆ ตามต้องการ เช่น
	// DeleteFile(publicID string) error
}
