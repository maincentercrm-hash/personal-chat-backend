// domain/repository/tag_repository.go

package repository

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/dto"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

type TagRepository interface {
	Create(tag *models.Tag) error
	GetByBusinessID(businessID uuid.UUID) ([]*models.Tag, error)
	GetBusinessTagsWithUserCount(businessID uuid.UUID) ([]dto.TagInfo, error)
	GetByID(id uuid.UUID) (*models.Tag, error)
	Update(tag *models.Tag) error
	Delete(id uuid.UUID) error
}
