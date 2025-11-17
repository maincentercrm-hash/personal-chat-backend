package dto

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/types"
)

// ============ Request DTOs ============

// WelcomeMessageCreateRequest สำหรับการสร้างข้อความต้อนรับใหม่
type WelcomeMessageCreateRequest struct {
	MessageType   string          `json:"message_type" validate:"required,oneof=text image card carousel flex"`
	Title         string          `json:"title" validate:"required"`
	Content       string          `json:"content"`
	ImageURL      string          `json:"image_url"`
	ThumbnailURL  string          `json:"thumbnail_url"`
	ActionButtons json.RawMessage `json:"action_buttons"`
	Components    json.RawMessage `json:"components"`
	TriggerType   string          `json:"trigger_type" validate:"required,oneof=follow inactive schedule command conversation_start location event"`
	TriggerParams json.RawMessage `json:"trigger_params"`
	SortOrder     int             `json:"sort_order"`
}

// WelcomeMessageUpdateRequest สำหรับการอัปเดตข้อความต้อนรับ
type WelcomeMessageUpdateRequest struct {
	MessageType   string                 `json:"message_type,omitempty" validate:"omitempty,oneof=text image card carousel flex"`
	Title         string                 `json:"title,omitempty"`
	Content       string                 `json:"content,omitempty"`
	ImageURL      string                 `json:"image_url,omitempty"`
	ThumbnailURL  string                 `json:"thumbnail_url,omitempty"`
	ActionButtons map[string]interface{} `json:"action_buttons,omitempty"`
	Components    map[string]interface{} `json:"components,omitempty"`
	TriggerType   string                 `json:"trigger_type,omitempty" validate:"omitempty,oneof=follow inactive schedule command conversation_start location event"`
	TriggerParams map[string]interface{} `json:"trigger_params,omitempty"`
	SortOrder     *int                   `json:"sort_order,omitempty"`
	IsActive      *bool                  `json:"is_active,omitempty"`
}

// WelcomeMessageStatusRequest สำหรับการอัปเดตสถานะข้อความต้อนรับ
type WelcomeMessageStatusRequest struct {
	IsActive bool `json:"is_active" validate:"required"`
}

// WelcomeMessageSortOrderRequest สำหรับการอัปเดตลำดับการแสดงผลข้อความต้อนรับ
type WelcomeMessageSortOrderRequest struct {
	SortOrder int `json:"sort_order" validate:"required,min=0"`
}

// WelcomeMessageTrackClickRequest สำหรับการบันทึกการคลิก
type WelcomeMessageTrackClickRequest struct {
	ActionType string          `json:"action_type" validate:"required"`
	ActionData json.RawMessage `json:"action_data"`
}

// WelcomeMessageQueryParams สำหรับการดึงข้อมูลข้อความต้อนรับ
type WelcomeMessageQueryParams struct {
	IncludeInactive bool   `json:"include_inactive,omitempty"`
	TriggerType     string `json:"trigger_type,omitempty" validate:"omitempty,oneof=follow inactive schedule command conversation_start location event"`
}

// ============ Response DTOs ============

// WelcomeMessageItem ข้อมูลข้อความต้อนรับ
type WelcomeMessageItem struct {
	ID            uuid.UUID   `json:"id"`
	BusinessID    uuid.UUID   `json:"business_id"`
	IsActive      bool        `json:"is_active"`
	MessageType   string      `json:"message_type"`
	Title         string      `json:"title"`
	Content       string      `json:"content"`
	ImageURL      string      `json:"image_url,omitempty"`
	ThumbnailURL  string      `json:"thumbnail_url,omitempty"`
	ActionButtons types.JSONB `json:"action_buttons,omitempty"`
	Components    types.JSONB `json:"components,omitempty"`
	TriggerType   string      `json:"trigger_type"`
	TriggerParams types.JSONB `json:"trigger_params,omitempty"`
	SortOrder     int         `json:"sort_order"`
	SentCount     int         `json:"sent_count"`
	ClickCount    int         `json:"click_count"`
	ReplyCount    int         `json:"reply_count"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	CreatedByID   uuid.UUID   `json:"created_by_id"`
	UpdatedByID   uuid.UUID   `json:"updated_by_id"`
}

// WelcomeMessageResponse สำหรับผลลัพธ์ของการดึงข้อมูลข้อความต้อนรับ
type WelcomeMessageResponse struct {
	GenericResponse
	WelcomeMessage WelcomeMessageItem `json:"welcome_message"`
}

// WelcomeMessageListResponse สำหรับผลลัพธ์ของการดึงรายการข้อความต้อนรับ
type WelcomeMessageListResponse struct {
	GenericResponse
	WelcomeMessages []WelcomeMessageItem `json:"welcome_messages"`
	TriggerType     string               `json:"trigger_type,omitempty"`
}

// WelcomeMessageStatusResponse สำหรับผลลัพธ์ของการอัปเดตสถานะข้อความต้อนรับ
type WelcomeMessageStatusResponse struct {
	GenericResponse
	IsActive bool `json:"is_active"`
}

// WelcomeMessageSortOrderResponse สำหรับผลลัพธ์ของการอัปเดตลำดับการแสดงผลข้อความต้อนรับ
type WelcomeMessageSortOrderResponse struct {
	GenericResponse
	SortOrder int `json:"sort_order"`
}

// ============ Constants ============

// WelcomeMessageType ประเภทของข้อความต้อนรับ
type WelcomeMessageType string

const (
	WelcomeMessageTypeText     WelcomeMessageType = "text"
	WelcomeMessageTypeImage    WelcomeMessageType = "image"
	WelcomeMessageTypeCard     WelcomeMessageType = "card"
	WelcomeMessageTypeCarousel WelcomeMessageType = "carousel"
	WelcomeMessageTypeFlex     WelcomeMessageType = "flex"
)

// WelcomeTriggerType ประเภทของทริกเกอร์
type WelcomeTriggerType string

const (
	WelcomeTriggerTypeFollow            WelcomeTriggerType = "follow"
	WelcomeTriggerTypeInactive          WelcomeTriggerType = "inactive"
	WelcomeTriggerTypeSchedule          WelcomeTriggerType = "schedule"
	WelcomeTriggerTypeCommand           WelcomeTriggerType = "command"
	WelcomeTriggerTypeConversationStart WelcomeTriggerType = "conversation_start"
	WelcomeTriggerTypeLocation          WelcomeTriggerType = "location"
	WelcomeTriggerTypeEvent             WelcomeTriggerType = "event"
)
