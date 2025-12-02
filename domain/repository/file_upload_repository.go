// domain/repository/file_upload_repository.go
package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

// FileUploadRepository handles file upload records
type FileUploadRepository interface {
	// Create creates a new file upload record
	Create(upload *models.FileUpload) error

	// FindByID finds a file upload by ID
	FindByID(id uuid.UUID) (*models.FileUpload, error)

	// FindByUserID finds all file uploads by user ID
	FindByUserID(userID uuid.UUID, limit, offset int) ([]*models.FileUpload, error)

	// Update updates a file upload record
	Update(upload *models.FileUpload) error

	// UpdateStatus updates the status of a file upload
	UpdateStatus(id uuid.UUID, status models.FileUploadStatus) error

	// MarkAsCompleted marks a file upload as completed
	MarkAsCompleted(id uuid.UUID, url string) error

	// Delete deletes a file upload record
	Delete(id uuid.UUID) error

	// FindPendingOlderThan finds pending uploads older than the given time (for cleanup)
	FindPendingOlderThan(cutoff time.Time) ([]*models.FileUpload, error)

	// CountByUserID counts uploads by user ID (for rate limiting)
	CountByUserID(userID uuid.UUID, since time.Time) (int64, error)
}
