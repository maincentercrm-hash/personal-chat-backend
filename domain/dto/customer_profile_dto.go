package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/types"
)

// ============ Request DTOs ============

// CreateCustomerProfileRequest สำหรับการสร้างโปรไฟล์ลูกค้าใหม่
type CreateCustomerProfileRequest struct {
	Nickname     string `json:"nickname" validate:"required"`
	Notes        string `json:"notes"`
	CustomerType string `json:"customer_type" validate:"required"`
}

// UpdateCustomerProfileRequest สำหรับการอัปเดตโปรไฟล์ลูกค้า
type UpdateCustomerProfileRequest struct {
	Nickname     string   `json:"nickname,omitempty"`
	Notes        string   `json:"notes,omitempty"`
	CustomerType string   `json:"customer_type,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Status       string   `json:"status,omitempty"`
	// สามารถเพิ่ม field อื่นๆ ตามต้องการ
}

// CustomerSearchRequest สำหรับการค้นหาลูกค้า
type CustomerSearchRequest struct {
	Query  string `json:"q" validate:"required"`
	Limit  int    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Offset int    `json:"offset,omitempty" validate:"omitempty,min=0"`
}

// CustomerListRequest สำหรับการดึงรายชื่อลูกค้า
type CustomerListRequest struct {
	Limit  int `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Offset int `json:"offset,omitempty" validate:"omitempty,min=0"`
}

// ============ Response DTOs ============

// CustomerProfileDTO ข้อมูลโปรไฟล์ลูกค้า
type CustomerProfileDTO struct {
	ID           uuid.UUID   `json:"id"`
	BusinessID   uuid.UUID   `json:"business_id"`
	UserID       uuid.UUID   `json:"user_id"`
	Username     string      `json:"username"`
	DisplayName  string      `json:"display_name"`
	Nickname     string      `json:"nickname"`
	Notes        string      `json:"notes"`
	CustomerType string      `json:"customer_type"`
	Tags         []SimpleTag `json:"tags,omitempty"` // เปลี่ยนจาก []string เป็น []SimpleTag
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	CreatedBy    *uuid.UUID  `json:"created_by,omitempty"`
	UpdatedBy    *uuid.UUID  `json:"updated_by,omitempty"`
	LastContact  *time.Time  `json:"last_contact,omitempty"`
	Status       string      `json:"status,omitempty"`
	Metadata     types.JSONB `json:"metadata,omitempty"`
	// ข้อมูลเพิ่มเติมเกี่ยวกับลูกค้า

	ConversationID  *uuid.UUID `json:"conversation_id,omitempty"`
	ProfileImageURL *string    `json:"profile_image_url,omitempty"`
}

// CustomerListData ข้อมูลรายการลูกค้า
type CustomerListData struct {
	Customers []CustomerProfileDTO `json:"customers"`
	Total     int                  `json:"total"`
	Limit     int                  `json:"limit"`
	Offset    int                  `json:"offset"`
}

// CustomerSearchData ข้อมูลผลการค้นหาลูกค้า
type CustomerSearchData struct {
	Customers []CustomerProfileDTO `json:"customers"`
	Total     int                  `json:"total"`
	Limit     int                  `json:"limit"`
	Offset    int                  `json:"offset"`
	Query     string               `json:"query"`
}

// CustomerProfileResponse สำหรับผลลัพธ์การดึงข้อมูลโปรไฟล์ลูกค้า
type CustomerProfileResponse struct {
	GenericResponse
	Data CustomerProfileDTO `json:"data"`
}

// CustomerListResponse สำหรับผลลัพธ์การดึงรายการลูกค้า
type CustomerListResponse struct {
	GenericResponse
	Data CustomerListData `json:"data"`
}

// CustomerSearchResponse สำหรับผลลัพธ์การค้นหาลูกค้า
type CustomerSearchResponse struct {
	GenericResponse
	Data CustomerSearchData `json:"data"`
}

type SimpleTag struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Color string    `json:"color"`
}
