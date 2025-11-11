package dto

import (
	"time"

	"github.com/google/uuid"
)

// ============ Request DTOs ============

// BusinessFollowRequest สำหรับการติดตามธุรกิจ
type BusinessFollowRequest struct {
	Source string `json:"source,omitempty"`
}

// BusinessFollowersQueryRequest สำหรับการดึงรายชื่อผู้ติดตาม
type BusinessFollowersQueryRequest struct {
	Limit  int `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Offset int `json:"offset,omitempty" validate:"omitempty,min=0"`
}

// UserFollowedBusinessesQueryRequest สำหรับการดึงรายชื่อธุรกิจที่ผู้ใช้ติดตาม
type UserFollowedBusinessesQueryRequest struct {
	UserID *uuid.UUID `json:"user_id,omitempty"`
	Limit  int        `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Offset int        `json:"offset,omitempty" validate:"omitempty,min=0"`
}

// ============ Response DTOs ============

// BusinessFollowerItem ข้อมูลผู้ติดตาม
type BusinessFollowerItem struct {
	UserID          uuid.UUID  `json:"user_id"`
	Username        string     `json:"username"`
	DisplayName     string     `json:"display_name"`
	ProfileImageURL *string    `json:"profile_image_url,omitempty"`
	FollowedAt      time.Time  `json:"followed_at"`
	Source          string     `json:"source,omitempty"`
	LastActiveAt    *time.Time `json:"last_active_at,omitempty"`
}

// BusinessFollowersResponse สำหรับรายชื่อผู้ติดตาม
type BusinessFollowersResponse struct {
	Success bool                  `json:"success"`
	Data    BusinessFollowersData `json:"data"`
}

// BusinessFollowersData ข้อมูลผู้ติดตาม
type BusinessFollowersData struct {
	Followers []BusinessFollowerItem `json:"followers"`
	Total     int64                  `json:"total"`
	Limit     int                    `json:"limit"`
	Offset    int                    `json:"offset"`
}

// FollowedBusinessItem ข้อมูลธุรกิจที่ติดตาม
type FollowedBusinessItem struct {
	BusinessID      uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Username        string    `json:"username"`
	Description     string    `json:"description,omitempty"`
	ProfileImageURL *string   `json:"profile_image_url,omitempty"`
	CoverImageURL   *string   `json:"cover_image_url,omitempty"`
	FollowedAt      time.Time `json:"followed_at"`
	Source          string    `json:"source,omitempty"`
	FollowerCount   int       `json:"follower_count"`
}

// UserFollowedBusinessesResponse สำหรับรายชื่อธุรกิจที่ผู้ใช้ติดตาม
type UserFollowedBusinessesResponse struct {
	Success bool                       `json:"success"`
	Data    UserFollowedBusinessesData `json:"data"`
}

// UserFollowedBusinessesData ข้อมูลธุรกิจที่ผู้ใช้ติดตาม
type UserFollowedBusinessesData struct {
	Businesses []FollowedBusinessItem `json:"businesses"`
	Total      int64                  `json:"total"`
	Limit      int                    `json:"limit"`
	Offset     int                    `json:"offset"`
}

// FollowStatusResponse สำหรับสถานะการติดตาม
type FollowStatusResponse struct {
	Success bool             `json:"success"`
	Data    FollowStatusData `json:"data"`
}

// FollowStatusData ข้อมูลสถานะการติดตาม
type FollowStatusData struct {
	IsFollowing bool `json:"is_following"`
}
