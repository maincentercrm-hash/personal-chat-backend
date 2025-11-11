// domain/dto/user_dto.go
package dto

import (
	"time"

	"github.com/google/uuid"
)

// ============ Constants ============

// UserStatus สถานะของผู้ใช้
type UserStatus string

const (
	UserStatusOnline  UserStatus = "online"
	UserStatusAway    UserStatus = "away"
	UserStatusOffline UserStatus = "offline"
	UserStatusBusy    UserStatus = "busy"
)

// ============ Request DTOs ============

// UpdateProfileRequest สำหรับการอัปเดตโปรไฟล์
type UpdateProfileRequest struct {
	DisplayName string  `json:"display_name" validate:"omitempty,min=2,max=50"`
	Bio         *string `json:"bio" validate:"omitempty,max=200"`
	Status      string  `json:"status" validate:"omitempty,oneof=online away offline busy"`
}

// UploadProfileImageRequest สำหรับการอัปโหลดรูปโปรไฟล์
// ใช้ multipart/form-data ไม่ใช่ JSON
type UploadProfileImageRequest struct {
	// ไม่มีฟิลด์ JSON เนื่องจากใช้ FormFile
}

// GetUserStatusRequest สำหรับการดึงสถานะผู้ใช้
type GetUserStatusRequest struct {
	IDs string `json:"ids" query:"ids" validate:"required"` // รูปแบบ comma-separated UUIDs
}

// ============ Response Data DTOs ============

// UserData ข้อมูลผู้ใช้
type UserData struct {
	ID                uuid.UUID  `json:"id"`
	Username          string     `json:"username"`
	DisplayName       string     `json:"display_name"`
	Email             *string    `json:"email,omitempty"` // ซ่อนเมื่อดูโปรไฟล์คนอื่น
	ProfileImageURL   *string    `json:"profile_image_url,omitempty"`
	Bio               *string    `json:"bio,omitempty"`
	Status            UserStatus `json:"status"`
	LastActiveAt      *time.Time `json:"last_active_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at,omitempty"`         // ซ่อนเมื่อดูโปรไฟล์คนอื่น
	IsEmailVerified   bool       `json:"is_email_verified,omitempty"`  // ซ่อนเมื่อดูโปรไฟล์คนอื่น
	IsPhoneVerified   bool       `json:"is_phone_verified,omitempty"`  // ซ่อนเมื่อดูโปรไฟล์คนอื่น
	Phone             *string    `json:"phone,omitempty"`              // ซ่อนเมื่อดูโปรไฟล์คนอื่น
	NotificationPrefs *string    `json:"notification_prefs,omitempty"` // ซ่อนเมื่อดูโปรไฟล์คนอื่น
}

// PublicUserData ข้อมูลผู้ใช้ที่เปิดเผยต่อสาธารณะ
type PublicUserData struct {
	ID              uuid.UUID  `json:"id"`
	Username        string     `json:"username"`
	DisplayName     string     `json:"display_name"`
	ProfileImageURL *string    `json:"profile_image_url,omitempty"`
	Bio             *string    `json:"bio,omitempty"`
	Status          UserStatus `json:"status"`
	LastActiveAt    *time.Time `json:"last_active_at,omitempty"`
}

// UserStatusItem ข้อมูลสถานะของผู้ใช้
type UserStatusItem struct {
	UserID       uuid.UUID  `json:"user_id"`
	Status       UserStatus `json:"status"`
	LastActiveAt *time.Time `json:"last_active_at,omitempty"`
}

// ProfileImageUploadResult ผลลัพธ์การอัปโหลดรูปโปรไฟล์
type ProfileImageUploadResult struct {
	ProfileImageURL string `json:"profile_image_url"`
	PublicID        string `json:"public_id"`
}

// SearchUserItem ผลลัพธ์การค้นหาผู้ใช้
type SearchUserItem struct {
	ID              uuid.UUID  `json:"id"`
	Username        string     `json:"username"`
	DisplayName     string     `json:"display_name"`
	ProfileImageURL *string    `json:"profile_image_url,omitempty"`
	Bio             *string    `json:"bio,omitempty"`
	Status          UserStatus `json:"status"`
}

// ============ Response Wrapper DTOs ============

// GetCurrentUserResponse การตอบกลับสำหรับการดึงข้อมูลผู้ใช้ปัจจุบัน
type GetCurrentUserResponse struct {
	GenericResponse
	User UserData `json:"user"`
}

// GetProfileResponse การตอบกลับสำหรับการดึงข้อมูลโปรไฟล์
type GetProfileResponse struct {
	GenericResponse
	Data interface{} `json:"data"` // อาจเป็น UserData หรือ PublicUserData
}

// UpdateProfileResponse การตอบกลับสำหรับการอัปเดตโปรไฟล์
type UpdateProfileResponse struct {
	GenericResponse
	Data UserData `json:"data"`
}

// UploadProfileImageResponse การตอบกลับสำหรับการอัปโหลดรูปโปรไฟล์
type UploadProfileImageResponse struct {
	GenericResponse
	Data ProfileImageUploadResult `json:"data"`
}

// SearchUsersResponse การตอบกลับสำหรับการค้นหาผู้ใช้
type SearchUsersResponse struct {
	GenericResponse
	Data struct {
		Users  []SearchUserItem `json:"users"`
		Count  int              `json:"count"`
		Limit  int              `json:"limit"`
		Offset int              `json:"offset"`
	} `json:"data"`
}

// GetUserStatusResponse การตอบกลับสำหรับการดึงสถานะผู้ใช้
type GetUserStatusResponse struct {
	GenericResponse
	Data []UserStatusItem `json:"data"`
}
