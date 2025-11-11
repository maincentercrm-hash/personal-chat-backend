// domain/dto/auth_dto.go

package dto

import (
	"time"

	"github.com/google/uuid"
)

// ============ Request DTOs ============

// RegisterRequest สำหรับรับข้อมูลการลงทะเบียน
type RegisterRequest struct {
	Username    string `json:"username" validate:"required"`
	Password    string `json:"password" validate:"required,min=8"`
	Email       string `json:"email" validate:"required,email"`
	DisplayName string `json:"display_name" validate:"required"`
}

// LoginRequest สำหรับรับข้อมูลการเข้าสู่ระบบ
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// RefreshTokenRequest สำหรับรับ refresh token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ============ Response DTOs ============

// UserResponse สำหรับข้อมูลผู้ใช้ที่ส่งกลับ (ใช้ร่วมกันในหลาย endpoint)
type UserResponse struct {
	ID              uuid.UUID  `json:"id"`
	Username        string     `json:"username"`
	Email           string     `json:"email"`
	DisplayName     string     `json:"display_name"`
	ProfileImageURL *string    `json:"profile_image_url,omitempty"`
	Bio             *string    `json:"bio,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	LastActiveAt    *time.Time `json:"last_active_at,omitempty"`
	Status          string     `json:"status"`
}

// AuthResponse สำหรับผลลัพธ์การยืนยันตัวตน (ใช้ร่วมกันในการ register และ login)
type AuthResponse struct {
	GenericResponse
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

// RefreshTokenResponse สำหรับผลลัพธ์การรีเฟรช token
type RefreshTokenResponse struct {
	GenericResponse
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// LogoutResponse สำหรับผลลัพธ์การออกจากระบบ
type LogoutResponse struct {
	GenericResponse
}

// UserProfileResponse สำหรับผลลัพธ์การดึงข้อมูลผู้ใช้ปัจจุบัน
type UserProfileResponse struct {
	Success bool         `json:"success"`
	User    UserResponse `json:"user"`
}
