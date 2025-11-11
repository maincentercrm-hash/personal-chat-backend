// domain/service/user_tag_service.go
package service

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

type UserTagService interface {
	// AddTagToUser เพิ่มแท็กให้กับผู้ใช้
	AddTagToUser(businessID, userID, tagID, addedByID uuid.UUID) (*models.UserTag, error)

	// RemoveTagFromUser ลบแท็กออกจากผู้ใช้
	RemoveTagFromUser(businessID, userID, tagID, requestedByID uuid.UUID) error

	// GetUserTags ดึงแท็กทั้งหมดของผู้ใช้ในธุรกิจ
	GetUserTags(businessID, userID uuid.UUID, requestedByID uuid.UUID) ([]*models.UserTag, error)

	// GetUsersByTag ดึงรายชื่อผู้ใช้ที่มีแท็กนี้
	GetUsersByTag(businessID, tagID uuid.UUID, requestedByID uuid.UUID, limit, offset int) ([]*models.UserTag, int64, error)

	// BulkAddTagToUsers เพิ่มแท็กให้กับผู้ใช้หลายคน
	BulkAddTagToUsers(businessID, tagID uuid.UUID, userIDs []uuid.UUID, addedByID uuid.UUID) ([]*models.UserTag, error)

	// BulkRemoveTagFromUsers ลบแท็กออกจากผู้ใช้หลายคน
	BulkRemoveTagFromUsers(businessID, tagID uuid.UUID, userIDs []uuid.UUID, requestedByID uuid.UUID) error

	// ReplaceUserTags แทนที่แท็กทั้งหมดของผู้ใช้ด้วยแท็กใหม่
	ReplaceUserTags(businessID, userID uuid.UUID, tagIDs []uuid.UUID, updatedByID uuid.UUID) ([]*models.UserTag, error)

	// GetTagStatistics ดึงสถิติการใช้แท็ก
	GetTagStatistics(businessID uuid.UUID, requestedByID uuid.UUID) ([]TagStatistic, error)

	// SearchUsersByTags ค้นหาผู้ใช้ที่มีแท็กตามเงื่อนไข
	SearchUsersByTags(businessID uuid.UUID, criteria TagSearchCriteria, requestedByID uuid.UUID, limit, offset int) ([]*models.UserTag, int64, error)

	// CheckUserHasTag ตรวจสอบว่าผู้ใช้มีแท็กนี้หรือไม่
	CheckUserHasTag(businessID, userID, tagID uuid.UUID) (bool, error)

	// GetUsersWithMultipleTags ดึงผู้ใช้ที่มีแท็กตามที่กำหนด (AND/OR logic)
	GetUsersWithMultipleTags(businessID uuid.UUID, tagIDs []uuid.UUID, matchType TagMatchType, requestedByID uuid.UUID, limit, offset int) ([]*models.UserTag, int64, error)
}

// TagStatistic สถิติการใช้แท็ก
type TagStatistic struct {
	TagID     uuid.UUID `json:"tag_id"`
	TagName   string    `json:"tag_name"`
	TagColor  string    `json:"tag_color"`
	UserCount int64     `json:"user_count"`
	CreatedAt string    `json:"created_at"`
}

// TagSearchCriteria เงื่อนไขในการค้นหา
type TagSearchCriteria struct {
	IncludeTags []uuid.UUID  `json:"include_tags,omitempty"` // แท็กที่ต้องมี
	ExcludeTags []uuid.UUID  `json:"exclude_tags,omitempty"` // แท็กที่ไม่ต้องมี
	MatchType   TagMatchType `json:"match_type"`             // วิธีการจับคู่
}

// TagMatchType ประเภทการจับคู่แท็ก
type TagMatchType string

const (
	TagMatchAll TagMatchType = "all" // ต้องมีทุกแท็กที่ระบุ (AND)
	TagMatchAny TagMatchType = "any" // มีแท็กใดแท็กหนึ่งที่ระบุ (OR)
)
