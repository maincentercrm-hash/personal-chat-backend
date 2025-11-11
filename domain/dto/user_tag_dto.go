// domain/dto/user_tag_dto.go
package dto

import (
	"time"

	"github.com/google/uuid"
)

// ============ Constants ============

// TagMatchType ประเภทของการจับคู่แท็ก
type TagMatchType string

const (
	TagMatchAll TagMatchType = "all" // ต้องมีทุกแท็กที่ระบุ
	TagMatchAny TagMatchType = "any" // มีแท็กใดแท็กหนึ่งที่ระบุ
)

// ============ Request DTOs ============

// AddTagToUserParam พารามิเตอร์สำหรับการเพิ่มแท็กให้ผู้ใช้
type AddTagToUserParam struct {
	BusinessID uuid.UUID `json:"business_id" validate:"required"`
	UserID     uuid.UUID `json:"user_id" validate:"required"`
	TagID      uuid.UUID `json:"tag_id" validate:"required"`
}

// RemoveTagFromUserParam พารามิเตอร์สำหรับการลบแท็กออกจากผู้ใช้
type RemoveTagFromUserParam struct {
	BusinessID uuid.UUID `json:"business_id" validate:"required"`
	UserID     uuid.UUID `json:"user_id" validate:"required"`
	TagID      uuid.UUID `json:"tag_id" validate:"required"`
}

// CheckUserHasTagParam พารามิเตอร์สำหรับการตรวจสอบว่าผู้ใช้มีแท็กหรือไม่
type CheckUserHasTagParam struct {
	BusinessID uuid.UUID `json:"business_id" validate:"required"`
	UserID     uuid.UUID `json:"user_id" validate:"required"`
	TagID      uuid.UUID `json:"tag_id" validate:"required"`
}

// ReplaceUserTagsRequest สำหรับการแทนที่แท็กทั้งหมดของผู้ใช้
type ReplaceUserTagsRequest struct {
	TagIDs []string `json:"tag_ids" validate:"required"`
}

// BulkAddTagToUsersRequest สำหรับการเพิ่มแท็กให้ผู้ใช้หลายคน
type BulkAddTagToUsersRequest struct {
	UserIDs []string `json:"user_ids" validate:"required,min=1"`
}

// BulkRemoveTagFromUsersRequest สำหรับการลบแท็กจากผู้ใช้หลายคน
type BulkRemoveTagFromUsersRequest struct {
	UserIDs []string `json:"user_ids" validate:"required,min=1"`
}

// SearchUsersByTagsRequest สำหรับการค้นหาผู้ใช้ตามแท็ก
type SearchUsersByTagsRequest struct {
	IncludeTags []string     `json:"include_tags"` // แท็กที่ต้องมี
	ExcludeTags []string     `json:"exclude_tags"` // แท็กที่ไม่ต้องมี
	MatchType   TagMatchType `json:"match_type"`   // "all" หรือ "any"
	Limit       int          `json:"limit" query:"limit"`
	Offset      int          `json:"offset" query:"offset"`
}

// GetUsersWithMultipleTagsParam พารามิเตอร์สำหรับการดึงผู้ใช้ที่มีหลายแท็ก
type GetUsersWithMultipleTagsParam struct {
	TagIDs    string       `json:"tag_ids" query:"tag_ids" validate:"required"` // คั่นด้วยเครื่องหมายจุลภาค
	MatchType TagMatchType `json:"match_type" query:"match_type"`               // "all" หรือ "any"
	Limit     int          `json:"limit" query:"limit"`
	Offset    int          `json:"offset" query:"offset"`
}

// GetTagStatisticsParam พารามิเตอร์สำหรับการดึงสถิติการใช้แท็ก
type GetTagStatisticsParam struct {
	BusinessID uuid.UUID `json:"business_id" validate:"required"`
}

// GetUserTagAnalyticsParam พารามิเตอร์สำหรับการดึงข้อมูลวิเคราะห์ UserTag
type GetUserTagAnalyticsParam struct {
	Type string `json:"type" query:"type"` // "overview", "trends", "distribution"
	Days int    `json:"days" query:"days"` // สำหรับ trends
}

// ExportUserTagsParam พารามิเตอร์สำหรับการส่งออกข้อมูล UserTags
type ExportUserTagsParam struct {
	Format string     `json:"format" query:"format"` // "json" หรือ "csv"
	TagID  *uuid.UUID `json:"tag_id" query:"tag_id"` // Optional
}

// GetRecentTagActivityParam พารามิเตอร์สำหรับการดึงกิจกรรมแท็กล่าสุด
type GetRecentTagActivityParam struct {
	Limit int `json:"limit" query:"limit"`
}

// ValidateTaggingRequest สำหรับการตรวจสอบความถูกต้องของการแท็ก
type ValidateTaggingRequest struct {
	UserID string   `json:"user_id" validate:"required"`
	TagIDs []string `json:"tag_ids" validate:"required,min=1"`
}

// TagSearchCriteria เกณฑ์การค้นหาแท็ก
type TagSearchCriteria struct {
	IncludeTags []uuid.UUID  `json:"include_tags"`
	ExcludeTags []uuid.UUID  `json:"exclude_tags"`
	MatchType   TagMatchType `json:"match_type"`
}

// ============ Response Data DTOs ============

// UserTagItem ข้อมูลแท็กของผู้ใช้
type UserTagItem struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	TagID       uuid.UUID `json:"tag_id"`
	BusinessID  uuid.UUID `json:"business_id"`
	AddedAt     time.Time `json:"added_at"`
	AddedBy     uuid.UUID `json:"added_by"`
	TagName     string    `json:"tag_name"`
	TagColor    string    `json:"tag_color"`
	UserName    *string   `json:"user_name,omitempty"`     // ชื่อผู้ใช้ (ถ้ามี)
	DisplayName *string   `json:"display_name,omitempty"`  // ชื่อที่แสดงของผู้ใช้ (ถ้ามี)
	AddedByName *string   `json:"added_by_name,omitempty"` // ชื่อผู้เพิ่มแท็ก (ถ้ามี)
}

// TaggedUserItem ข้อมูลผู้ใช้ที่มีแท็ก
type TaggedUserItem struct {
	UserID          uuid.UUID     `json:"user_id"`
	Username        string        `json:"username"`
	DisplayName     string        `json:"display_name"`
	ProfileImageURL *string       `json:"profile_image_url,omitempty"`
	Email           *string       `json:"email,omitempty"`
	LastActiveAt    *time.Time    `json:"last_active_at,omitempty"`
	Tags            []TagInfoItem `json:"tags,omitempty"`
}

// TagInfoItem ข้อมูลแท็ก
type TagInfoItem struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Color       string    `json:"color"`
	UserTagID   uuid.UUID `json:"user_tag_id,omitempty"` // ID ของการเชื่อมโยงแท็กกับผู้ใช้
	BusinessID  uuid.UUID `json:"business_id"`
	AddedAt     time.Time `json:"added_at,omitempty"`
	AddedBy     uuid.UUID `json:"added_by,omitempty"`
	AddedByName *string   `json:"added_by_name,omitempty"`
}

// BulkTagResultItem ผลลัพธ์ของการเพิ่มแท็กแบบหลายรายการ
type BulkTagResultItem struct {
	SuccessfulCount int           `json:"successful_count"`
	RequestedCount  int           `json:"requested_count"`
	AddedUserTags   []UserTagItem `json:"added_user_tags"`
}

// TagStatisticItem สถิติของแท็ก
type TagStatisticItem struct {
	TagID      uuid.UUID `json:"tag_id"`
	TagName    string    `json:"tag_name"`
	TagColor   string    `json:"tag_color"`
	UserCount  int       `json:"user_count"`
	CreatedAt  time.Time `json:"created_at"`
	BusinessID uuid.UUID `json:"business_id"`
}

// TagValidationResultItem ผลลัพธ์การตรวจสอบแท็ก
type TagValidationResultItem struct {
	TagID  uuid.UUID `json:"tag_id"`
	Valid  bool      `json:"valid"`
	HasTag bool      `json:"has_tag,omitempty"`
	Error  string    `json:"error,omitempty"`
}

// ============ Response Wrapper DTOs ============

// AddTagToUserResponse การตอบกลับสำหรับการเพิ่มแท็กให้ผู้ใช้
type AddTagToUserResponse struct {
	GenericResponse
	Data UserTagItem `json:"data"`
}

// RemoveTagFromUserResponse การตอบกลับสำหรับการลบแท็กออกจากผู้ใช้
type RemoveTagFromUserResponse struct {
	GenericResponse
}

// GetUserTagsResponse การตอบกลับสำหรับการดึงแท็กของผู้ใช้
type GetUserTagsResponse struct {
	GenericResponse
	Data struct {
		UserTags []UserTagItem `json:"user_tags"`
		Count    int           `json:"count"`
	} `json:"data"`
}

// GetUsersByTagResponse การตอบกลับสำหรับการดึงผู้ใช้ตามแท็ก
type GetUsersByTagResponse struct {
	GenericResponse
	Data struct {
		UserTags []UserTagItem `json:"user_tags"`
		Total    int           `json:"total"`
		Limit    int           `json:"limit"`
		Offset   int           `json:"offset"`
	} `json:"data"`
}

// CheckUserHasTagResponse การตอบกลับสำหรับการตรวจสอบว่าผู้ใช้มีแท็กหรือไม่
type CheckUserHasTagResponse struct {
	GenericResponse
	Data struct {
		HasTag     bool      `json:"has_tag"`
		UserID     uuid.UUID `json:"user_id"`
		TagID      uuid.UUID `json:"tag_id"`
		BusinessID uuid.UUID `json:"business_id"`
	} `json:"data"`
}

// ReplaceUserTagsResponse การตอบกลับสำหรับการแทนที่แท็กของผู้ใช้
type ReplaceUserTagsResponse struct {
	GenericResponse
	Data struct {
		NewUserTags []UserTagItem `json:"new_user_tags"`
		Count       int           `json:"count"`
	} `json:"data"`
}

// BulkAddTagToUsersResponse การตอบกลับสำหรับการเพิ่มแท็กให้ผู้ใช้หลายคน
type BulkAddTagToUsersResponse struct {
	GenericResponse
	Data BulkTagResultItem `json:"data"`
}

// BulkRemoveTagFromUsersResponse การตอบกลับสำหรับการลบแท็กจากผู้ใช้หลายคน
type BulkRemoveTagFromUsersResponse struct {
	GenericResponse
	Data struct {
		ProcessedCount int `json:"processed_count"`
	} `json:"data"`
}

// SearchUsersByTagsResponse การตอบกลับสำหรับการค้นหาผู้ใช้ตามแท็ก
type SearchUsersByTagsResponse struct {
	GenericResponse
	Data struct {
		UserTags []UserTagItem     `json:"user_tags"`
		Total    int               `json:"total"`
		Limit    int               `json:"limit"`
		Offset   int               `json:"offset"`
		Criteria TagSearchCriteria `json:"criteria"`
	} `json:"data"`
}

// GetUsersWithMultipleTagsResponse การตอบกลับสำหรับการดึงผู้ใช้ที่มีหลายแท็ก
type GetUsersWithMultipleTagsResponse struct {
	GenericResponse
	Data struct {
		UserTags  []UserTagItem `json:"user_tags"`
		Total     int           `json:"total"`
		Limit     int           `json:"limit"`
		Offset    int           `json:"offset"`
		TagIDs    []uuid.UUID   `json:"tag_ids"`
		MatchType TagMatchType  `json:"match_type"`
	} `json:"data"`
}

// GetTagStatisticsResponse การตอบกลับสำหรับการดึงสถิติการใช้แท็ก
type GetTagStatisticsResponse struct {
	GenericResponse
	Data struct {
		Statistics []TagStatisticItem `json:"statistics"`
		Count      int                `json:"count"`
	} `json:"data"`
}

// GetUserTagAnalyticsResponse การตอบกลับสำหรับการดึงข้อมูลวิเคราะห์ UserTag
type GetUserTagAnalyticsResponse struct {
	GenericResponse
	Data struct {
		Type       string             `json:"type"`
		Statistics []TagStatisticItem `json:"statistics,omitempty"`
		Days       int                `json:"days,omitempty"`
		// Trends     []TagTrendItem      `json:"trends,omitempty"` // ถ้ามีการเพิ่ม model นี้ในอนาคต
	} `json:"data"`
}

// ExportUserTagsResponse การตอบกลับสำหรับการส่งออกข้อมูล UserTags
type ExportUserTagsResponse struct {
	GenericResponse
	Data struct {
		Format     string             `json:"format"`
		UserTags   []UserTagItem      `json:"user_tags,omitempty"`
		Statistics []TagStatisticItem `json:"statistics,omitempty"`
		Count      int                `json:"count,omitempty"`
		ExportedAt string             `json:"exported_at"`
	} `json:"data"`
}

// GetRecentTagActivityResponse การตอบกลับสำหรับการดึงกิจกรรมแท็กล่าสุด
type GetRecentTagActivityResponse struct {
	GenericResponse
	Data struct {
		Limit int    `json:"limit"`
		Note  string `json:"note,omitempty"`
		// Activities []TagActivityItem `json:"activities,omitempty"` // ถ้ามีการเพิ่ม model นี้ในอนาคต
	} `json:"data"`
}

// ValidateTaggingResponse การตอบกลับสำหรับการตรวจสอบความถูกต้องของการแท็ก
type ValidateTaggingResponse struct {
	GenericResponse
	Data struct {
		UserID       uuid.UUID                 `json:"user_id"`
		BusinessID   uuid.UUID                 `json:"business_id"`
		TotalTags    int                       `json:"total_tags"`
		ValidTags    int                       `json:"valid_tags"`
		ExistingTags int                       `json:"existing_tags"`
		Results      []TagValidationResultItem `json:"results"`
	} `json:"data"`
}

// BulkAddTagToUsersPartialResponse การตอบกลับสำหรับการเพิ่มแท็กให้ผู้ใช้หลายคนแบบบางส่วนสำเร็จ
type BulkAddTagToUsersPartialResponse struct {
	GenericResponse
	Data BulkAddTagResult `json:"data"`
}
