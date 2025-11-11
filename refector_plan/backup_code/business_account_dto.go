package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/types"
)

// ============ Request DTOs ============

// BusinessCreateRequest สำหรับการสร้างธุรกิจใหม่
type BusinessCreateRequest struct {
	Name           string `json:"name" validate:"required"`
	Username       string `json:"username" validate:"required,min=3,max=30,alphanum"`
	Description    string `json:"description"`
	WelcomeMessage string `json:"welcome_message"`
}

// BusinessUpdateRequest สำหรับการอัปเดตข้อมูลธุรกิจ
type BusinessUpdateRequest struct {
	Name            string `json:"name,omitempty"`
	Description     string `json:"description,omitempty"`
	ProfileImageURL string `json:"profile_image_url,omitempty"`
	CoverImageURL   string `json:"cover_image_url,omitempty"`
	WelcomeMessage  string `json:"welcome_message,omitempty"`
}

// BusinessSearchRequest สำหรับการค้นหาธุรกิจ
type BusinessSearchRequest struct {
	Query  string `json:"q" validate:"required"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}

// BusinessUsernameSearchRequest สำหรับการค้นหาธุรกิจด้วย username
type BusinessUsernameSearchRequest struct {
	Username string `json:"username" validate:"required"`
}

// ============ Response DTOs ============

// BusinessItem ข้อมูลของธุรกิจ
type BusinessItem struct {
	ID              uuid.UUID   `json:"id"`
	Name            string      `json:"name"`
	Username        string      `json:"username"`
	Description     string      `json:"description,omitempty"`
	ProfileImageURL *string     `json:"profile_image_url,omitempty"`
	CoverImageURL   *string     `json:"cover_image_url,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	OwnerID         *uuid.UUID  `json:"owner_id,omitempty"`
	Status          string      `json:"status"`
	Settings        types.JSONB `json:"settings,omitempty"`
	FollowerCount   int         `json:"follower_count"`
	IsFollowing     bool        `json:"is_following"`
	IsAdmin         bool        `json:"is_admin"`
	AdminRole       string      `json:"admin_role,omitempty"`
}

// BusinessDetailResponse สำหรับรายละเอียดของธุรกิจ
type BusinessDetailResponse struct {
	GenericResponse
	Business BusinessItem `json:"business"`
}

// BusinessListResponse สำหรับรายการธุรกิจ
type BusinessListResponse struct {
	GenericResponse
	Data []BusinessItem `json:"data"`
}

// BusinessSearchResponse สำหรับผลลัพธ์การค้นหาธุรกิจ
type BusinessSearchResponse struct {
	Success bool               `json:"success"`
	Data    BusinessSearchData `json:"data"`
}

// BusinessSearchData ข้อมูลผลการค้นหาธุรกิจ
type BusinessSearchData struct {
	Businesses []BusinessItem `json:"businesses"`
	Total      int64          `json:"total"`
	Limit      int            `json:"limit"`
	Offset     int            `json:"offset"`
	Query      string         `json:"query"`
}

// BusinessImageUploadResponse สำหรับผลลัพธ์การอัปโหลดรูปภาพธุรกิจ
type BusinessImageUploadResponse struct {
	GenericResponse
	Data BusinessImageResult `json:"data"`
}

// BusinessImageResult ข้อมูลผลลัพธ์การอัปโหลดรูปภาพ
type BusinessImageResult struct {
	ProfileImageURL *string `json:"profile_image_url,omitempty"`
	CoverImageURL   *string `json:"cover_image_url,omitempty"`
	PublicID        string  `json:"public_id"`
}
