// infrastructure/persistence/postgres/file_upload_repository.go
package postgres

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"gorm.io/gorm"
)

type fileUploadRepository struct {
	db *gorm.DB
}

// NewFileUploadRepository creates a new file upload repository
func NewFileUploadRepository(db *gorm.DB) repository.FileUploadRepository {
	return &fileUploadRepository{db: db}
}

// Create creates a new file upload record
func (r *fileUploadRepository) Create(upload *models.FileUpload) error {
	return r.db.Create(upload).Error
}

// FindByID finds a file upload by ID
func (r *fileUploadRepository) FindByID(id uuid.UUID) (*models.FileUpload, error) {
	var upload models.FileUpload
	err := r.db.Where("id = ?", id).First(&upload).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("file upload not found")
		}
		return nil, err
	}
	return &upload, nil
}

// FindByUserID finds all file uploads by user ID
func (r *fileUploadRepository) FindByUserID(userID uuid.UUID, limit, offset int) ([]*models.FileUpload, error) {
	var uploads []*models.FileUpload
	query := r.db.Where("user_id = ?", userID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&uploads).Error
	return uploads, err
}

// Update updates a file upload record
func (r *fileUploadRepository) Update(upload *models.FileUpload) error {
	return r.db.Save(upload).Error
}

// UpdateStatus updates the status of a file upload
func (r *fileUploadRepository) UpdateStatus(id uuid.UUID, status models.FileUploadStatus) error {
	return r.db.Model(&models.FileUpload{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// MarkAsCompleted marks a file upload as completed
func (r *fileUploadRepository) MarkAsCompleted(id uuid.UUID, url string) error {
	now := time.Now()
	return r.db.Model(&models.FileUpload{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       models.FileUploadStatusCompleted,
			"url":          url,
			"completed_at": now,
		}).Error
}

// Delete deletes a file upload record
func (r *fileUploadRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.FileUpload{}, "id = ?", id).Error
}

// FindPendingOlderThan finds pending uploads older than the given time
func (r *fileUploadRepository) FindPendingOlderThan(cutoff time.Time) ([]*models.FileUpload, error) {
	var uploads []*models.FileUpload
	err := r.db.Where("status = ? AND created_at < ?", models.FileUploadStatusPending, cutoff).
		Find(&uploads).Error
	return uploads, err
}

// CountByUserID counts uploads by user ID since a given time
func (r *fileUploadRepository) CountByUserID(userID uuid.UUID, since time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&models.FileUpload{}).
		Where("user_id = ? AND created_at > ?", userID, since).
		Count(&count).Error
	return count, err
}
