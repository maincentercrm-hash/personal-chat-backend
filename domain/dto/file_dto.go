package dto

import (
	"time"

	"github.com/google/uuid"
)

// ============ Request DTOs ============

// FileUploadRequest สำหรับการอัปโหลดไฟล์
type FileUploadRequest struct {
	Folder string `form:"folder"`
	// ไฟล์จะถูกส่งมาในรูปแบบ multipart/form-data ซึ่งไม่สามารถระบุใน struct ได้โดยตรง
}

// ============ Response DTOs ============

// FileUploadDTO ข้อมูลผลลัพธ์การอัปโหลดไฟล์
type FileUploadDTO struct {
	ID          uuid.UUID              `json:"id"`
	FileName    string                 `json:"file_name"`
	FileSize    int64                  `json:"file_size"`
	FileType    string                 `json:"file_type"`
	MimeType    string                 `json:"mime_type"`
	URL         string                 `json:"url"`
	Path        string                 `json:"path"`
	Folder      string                 `json:"folder"`
	StorageType string                 `json:"storage_type"`
	UploadedAt  time.Time              `json:"uploaded_at"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`

	// สำหรับรูปภาพโดยเฉพาะ
	Width        *int    `json:"width,omitempty"`
	Height       *int    `json:"height,omitempty"`
	ThumbnailURL *string `json:"thumbnail_url,omitempty"`
}

// FileUploadResponse สำหรับผลลัพธ์การอัปโหลดไฟล์
type FileUploadResponse struct {
	GenericResponse
	Data FileUploadDTO `json:"data"`
}

// FileError สำหรับข้อผิดพลาดในการอัปโหลดไฟล์
type FileError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

// FileErrorResponse สำหรับการตอบกลับกรณีเกิดข้อผิดพลาด
type FileErrorResponse struct {
	GenericResponse
	Error FileError `json:"error"`
}

// FilesListDTO ข้อมูลรายการไฟล์
type FilesListDTO struct {
	Files  []FileUploadDTO `json:"files"`
	Total  int             `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

// FilesListResponse สำหรับผลลัพธ์การดึงรายการไฟล์
type FilesListResponse struct {
	GenericResponse
	Data FilesListDTO `json:"data"`
}
