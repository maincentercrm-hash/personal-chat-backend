package dto

import (
	"time"

	"github.com/google/uuid"
)

// ============ Request DTOs ============

// BusinessAdminAddRequest สำหรับการเพิ่มแอดมินใหม่
type BusinessAdminAddRequest struct {
	Username string `json:"username,omitempty"`
	UserID   string `json:"user_id,omitempty" validate:"omitempty,uuid4"`
	Role     string `json:"role" validate:"required,oneof=owner admin editor viewer"`
}

// BusinessAdminRoleUpdateRequest สำหรับการอัปเดตบทบาทของแอดมิน
type BusinessAdminRoleUpdateRequest struct {
	Role string `json:"role" validate:"required,oneof=owner admin editor viewer"`
}

// ============ Response DTOs ============

// BusinessAdminItem ข้อมูลของแอดมิน
type BusinessAdminItem struct {
	ID         uuid.UUID  `json:"id"`
	BusinessID uuid.UUID  `json:"business_id"`
	UserID     uuid.UUID  `json:"user_id"`
	Role       string     `json:"role"`
	AddedAt    time.Time  `json:"added_at"`
	AddedBy    *uuid.UUID `json:"added_by,omitempty"`

	// ข้อมูลเพิ่มเติมของผู้ใช้
	Username        string  `json:"username"`
	DisplayName     string  `json:"display_name"`
	ProfileImageURL *string `json:"profile_image_url,omitempty"`
	Email           string  `json:"email,omitempty"` // แสดงเฉพาะกับเจ้าของหรือแอดมิน
}

// BusinessAdminListResponse สำหรับรายการแอดมิน
type BusinessAdminListResponse struct {
	GenericResponse
	Admins []BusinessAdminItem `json:"admins"`
}

// BusinessAdminDetailResponse สำหรับรายละเอียดของแอดมิน
type BusinessAdminDetailResponse struct {
	GenericResponse
	Admin BusinessAdminItem `json:"admin"`
}
