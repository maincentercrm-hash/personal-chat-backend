// domain/models/file_upload.go
package models

import (
	"time"

	"github.com/google/uuid"
)

// FileUploadStatus represents the status of a file upload
type FileUploadStatus string

const (
	FileUploadStatusPending   FileUploadStatus = "pending"
	FileUploadStatusUploading FileUploadStatus = "uploading"
	FileUploadStatusCompleted FileUploadStatus = "completed"
	FileUploadStatusFailed    FileUploadStatus = "failed"
	FileUploadStatusBlocked   FileUploadStatus = "blocked" // For virus-infected files
)

// FileUpload tracks the lifecycle of file uploads
type FileUpload struct {
	ID          uuid.UUID        `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID      uuid.UUID        `json:"user_id" gorm:"type:uuid;not null;index"`
	Filename    string           `json:"filename" gorm:"type:varchar(255);not null"`
	ContentType string           `json:"content_type" gorm:"type:varchar(100);not null"`
	Size        int64            `json:"size" gorm:"not null"`
	Status      FileUploadStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending';index"`
	Path        string           `json:"path" gorm:"type:text;not null"` // R2/Storage path
	URL         string           `json:"url" gorm:"type:text"`           // Public URL (after completion)
	ExpiresAt   time.Time        `json:"expires_at" gorm:"type:timestamp with time zone;not null;index"`
	CreatedAt   time.Time        `json:"created_at" gorm:"type:timestamp with time zone;default:now()"`
	CompletedAt *time.Time       `json:"completed_at,omitempty" gorm:"type:timestamp with time zone"`

	// Relations
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for FileUpload
func (FileUpload) TableName() string {
	return "file_uploads"
}
