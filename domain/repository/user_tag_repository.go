// domain/repository/user_tag_repository.go

package repository

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

type UserTagRepository interface {
	Create(userTag *models.UserTag) error
	Delete(businessID, userID, tagID uuid.UUID) error
	GetUserTags(businessID, userID uuid.UUID) ([]*models.UserTag, error)
	GetUsersByTag(businessID, tagID uuid.UUID) ([]*models.UserTag, error)
	Exists(businessID, userID, tagID uuid.UUID) (bool, error) // เพิ่มฟังก์ชันนี้
}
