// domain/dto/tag_dto.go
package dto

import (
	"time"
)

// =====================================
// Request DTOs
// =====================================

// CreateTagRequest สำหรับการสร้างแท็กใหม่
type CreateTagRequest struct {
	Name  string `json:"name" validate:"required,max=50"`
	Color string `json:"color"`
}

// UpdateTagRequest สำหรับการอัปเดตแท็ก
type UpdateTagRequest struct {
	Name  string `json:"name" validate:"required"`
	Color string `json:"color"`
}

// AddTagToUserRequest สำหรับการเพิ่มแท็กให้ผู้ใช้
// ไม่มี body เนื่องจากใช้ข้อมูลจาก URL parameters
type AddTagToUserRequest struct{}

// =====================================
// Response DTOs
// =====================================

// TagInfo ข้อมูลของแท็ก
type TagInfo struct {
	ID         string    `json:"id"`
	BusinessID string    `json:"business_id"`
	Name       string    `json:"name"`
	Color      string    `json:"color"`
	CreatedAt  time.Time `json:"created_at"`
	CreatedBy  string    `json:"created_by,omitempty"`
	UserCount  int       `json:"user_count,omitempty"` // จำนวนผู้ใช้ที่มีแท็กนี้ (ถ้ามี)
}

// UserTagInfo ข้อมูลของแท็กที่เชื่อมโยงกับผู้ใช้
type UserTagInfo struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	TagID      string    `json:"tag_id"`
	BusinessID string    `json:"business_id"`
	AddedAt    time.Time `json:"added_at"`
	AddedBy    string    `json:"added_by,omitempty"`
	TagName    string    `json:"tag_name,omitempty"`
	TagColor   string    `json:"tag_color,omitempty"`
}

// CustomerTagInfo ข้อมูลลูกค้าพร้อมกับแท็ก
type CustomerTagInfo struct {
	UserID          string    `json:"user_id"`
	DisplayName     string    `json:"display_name"`
	ProfileImageURL string    `json:"profile_image_url,omitempty"`
	Username        string    `json:"username,omitempty"`
	Email           string    `json:"email,omitempty"`
	Tags            []TagInfo `json:"tags,omitempty"`
	LastActive      time.Time `json:"last_active,omitempty"`
}

// BulkAddTagResult ผลลัพธ์การเพิ่มแท็กแบบหลายรายการ
type BulkAddTagResult struct {
	SuccessfulCount int           `json:"successful_count"`
	RequestedCount  int           `json:"requested_count"`
	AddedUserTags   []UserTagInfo `json:"added_user_tags"`
	FailedUserIDs   []string      `json:"failed_user_ids,omitempty"`
}

// =====================================
// Response Wrapper DTOs
// =====================================

// CreateTagResponse การตอบกลับสำหรับการสร้างแท็ก
type CreateTagResponse struct {
	GenericResponse
	Data TagInfo `json:"data"`
}

// GetBusinessTagsResponse การตอบกลับสำหรับการดึงแท็กทั้งหมดของธุรกิจ
type GetBusinessTagsResponse struct {
	GenericResponse
	Data struct {
		Tags  []TagInfo `json:"tags"`
		Count int       `json:"count"`
	} `json:"data"`
}

// UpdateTagResponse การตอบกลับสำหรับการอัปเดตแท็ก
type UpdateTagResponse struct {
	GenericResponse
	Data TagInfo `json:"data"`
}

// DeleteTagResponse การตอบกลับสำหรับการลบแท็ก
type DeleteTagResponse struct {
	GenericResponse
}
