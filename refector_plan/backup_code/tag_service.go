// domain/service/tag_service.go
package service

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/dto"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

type TagService interface {
	CreateTag(businessID, createdByID uuid.UUID, name, color string) (*models.Tag, error)
	GetBusinessTags(businessID uuid.UUID) ([]*models.Tag, error)
	GetBusinessTagsWithInfo(businessID uuid.UUID) ([]dto.TagInfo, error)
	AddTagToUser(businessID, userID, tagID, addedByID uuid.UUID) error
	RemoveTagFromUser(businessID, userID, tagID uuid.UUID) error
	GetUserTags(businessID, userID uuid.UUID) ([]*models.Tag, error)
	GetUsersByTag(businessID, tagID uuid.UUID) ([]*models.CustomerProfile, error)
	UpdateTag(businessID, tagID, updatedByID uuid.UUID, name, color string) (*models.Tag, error)
	DeleteTag(businessID, tagID, deletedByID uuid.UUID) error
}
