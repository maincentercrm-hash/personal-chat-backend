package dto

import (
	"time"

	"github.com/google/uuid"
)

// ============ Request DTOs ============

// AddMemberRequest สำหรับการเพิ่มสมาชิกในการสนทนา
type AddMemberRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}

// ToggleAdminRequest สำหรับการเปลี่ยนสถานะแอดมินของสมาชิก
type ToggleAdminRequest struct {
	IsAdmin bool `json:"is_admin"`
}

// MemberQueryRequest สำหรับการดึงรายการสมาชิก
type MemberQueryRequest struct {
	Page  int `json:"page,omitempty" validate:"omitempty,min=1"`
	Limit int `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// ============ Response DTOs ============

// MemberDTO โครงสร้างข้อมูลสมาชิกในการสนทนา
type MemberDTO struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	Username       string    `json:"username"`
	DisplayName    string    `json:"display_name"`
	ProfilePicture string    `json:"profile_picture"`
	Role           string    `json:"role"` // "admin" หรือ "member"
	JoinedAt       time.Time `json:"joined_at"`
	IsOnline       bool      `json:"is_online"`
}

// ConversationMemberDTO ข้อมูลสมาชิกในการสนทนา
type ConversationMemberDTO struct {
	ID                   uuid.UUID   `json:"id"`
	ConversationID       uuid.UUID   `json:"conversation_id"`
	UserID               uuid.UUID   `json:"user_id"`
	Username             string      `json:"username"`
	DisplayName          string      `json:"display_name"`
	ProfileImageURL      *string     `json:"profile_image_url,omitempty"`
	IsAdmin              bool        `json:"is_admin"`
	JoinedAt             time.Time   `json:"joined_at"`
	LastReadAt           *time.Time  `json:"last_read_at,omitempty"`
	IsMuted              bool        `json:"is_muted"`
	IsPinned             bool        `json:"is_pinned"`
	Nickname             string      `json:"nickname,omitempty"`
	NotificationSettings interface{} `json:"notification_settings,omitempty"`
	Status               string      `json:"status,omitempty"`
	LastActiveAt         *time.Time  `json:"last_active_at,omitempty"`
}

// MemberListData ข้อมูลรายการสมาชิก
type MemberListData struct {
	Members    []ConversationMemberDTO `json:"members"`
	Total      int                     `json:"total"`
	Page       int                     `json:"page"`
	Limit      int                     `json:"limit"`
	TotalPages int                     `json:"total_pages"`
}

// MemberResponse สำหรับผลลัพธ์การดึงข้อมูลสมาชิก
type MemberResponse struct {
	GenericResponse
	Data ConversationMemberDTO `json:"data"`
}

// MemberListResponse สำหรับผลลัพธ์การดึงรายการสมาชิก
type MemberListResponse struct {
	GenericResponse
	Data MemberListData `json:"data"`
}

// AdminStatusResponse สำหรับผลลัพธ์การเปลี่ยนสถานะแอดมิน
type AdminStatusResponse struct {
	GenericResponse
	IsAdmin bool `json:"is_admin"`
}
