// domain/dto/boardcast_dto.go

package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/types"
)

// ============ Request DTOs ============

// BroadcastTargetCriteria โครงสร้างข้อมูลสำหรับกำหนดกลุ่มเป้าหมาย
type BroadcastTargetCriteria struct {
	// สำหรับเป้าหมายประเภท tags
	IncludeTags  []uuid.UUID `json:"include_tags,omitempty"`
	ExcludeTags  []uuid.UUID `json:"exclude_tags,omitempty"`
	TagMatchType string      `json:"tag_match_type,omitempty"` // all, any

	// สำหรับเป้าหมายประเภท specific_users
	UserIDs []uuid.UUID `json:"user_ids,omitempty"`

	// สำหรับเป้าหมายประเภท customer_profile
	CustomerTypes   []string    `json:"customer_types,omitempty"`
	LastContactFrom *time.Time  `json:"last_contact_from,omitempty"`
	LastContactTo   *time.Time  `json:"last_contact_to,omitempty"`
	Statuses        []string    `json:"statuses,omitempty"`
	CustomQuery     types.JSONB `json:"custom_query,omitempty"` // สำหรับ query ซับซ้อนอื่นๆ
}

// BroadcastCreateRequest สำหรับการสร้าง broadcast ใหม่
type BroadcastCreateRequest struct {
	Title       string                 `json:"title" validate:"required"`
	MessageType string                 `json:"message_type" validate:"required,oneof=text image carousel flex"`
	Content     string                 `json:"content"`
	MediaURL    string                 `json:"media_url"`
	BubbleType  string                 `json:"bubble_type"`
	BubbleData  map[string]interface{} `json:"bubble_data"`
	TargetType  string                 `json:"target_type" validate:"omitempty,oneof=all tags specific_users customer_profile"`
	TargetData  map[string]interface{} `json:"target_data"`
	ScheduleAt  *time.Time             `json:"schedule_at"`
	SendNow     bool                   `json:"send_now"`
}

// BroadcastUpdateRequest สำหรับการอัปเดต broadcast
type BroadcastUpdateRequest struct {
	Title      string                 `json:"title"`
	Content    string                 `json:"content"`
	MediaURL   string                 `json:"media_url"`
	BubbleType string                 `json:"bubble_type"`
	BubbleData map[string]interface{} `json:"bubble_data"`
	TargetType string                 `json:"target_type" validate:"omitempty,oneof=all tags specific_users customer_profile"`
	TargetData map[string]interface{} `json:"target_data"`
	ScheduleAt *time.Time             `json:"schedule_at"`
}

// BroadcastQueryRequest สำหรับการดึงรายการ broadcasts
type BroadcastQueryRequest struct {
	Status string `json:"status" validate:"omitempty,oneof=draft scheduled sending completed failed"`
	Limit  int    `json:"limit" validate:"omitempty,min=1,max=100"`
	Page   int    `json:"page" validate:"omitempty,min=1"`
}

// BroadcastEstimateTargetRequest สำหรับการประมาณจำนวนผู้รับ
type BroadcastEstimateTargetRequest struct {
	TargetType     string                 `json:"target_type" validate:"required,oneof=all tags specific_users customer_profile"`
	TargetCriteria map[string]interface{} `json:"target_criteria"`
}

// ============ Response DTOs ============

// BroadcastItem รายการ broadcast
type BroadcastItem struct {
	ID           uuid.UUID   `json:"id"`
	BusinessID   uuid.UUID   `json:"business_id"`
	Title        string      `json:"title"`
	MessageType  string      `json:"message_type"`
	Content      string      `json:"content"`
	MediaURL     string      `json:"media_url,omitempty"`
	ScheduledAt  *time.Time  `json:"scheduled_at,omitempty"`
	SentAt       *time.Time  `json:"sent_at,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	CreatedBy    uuid.UUID   `json:"created_by"`
	Status       string      `json:"status"`
	TargetType   string      `json:"target_type,omitempty"`
	TargetData   types.JSONB `json:"target_data,omitempty"`
	BubbleType   string      `json:"bubble_type,omitempty"`
	BubbleData   types.JSONB `json:"bubble_data,omitempty"`
	Metrics      types.JSONB `json:"metrics,omitempty"`
	ErrorMessage string      `json:"error_message,omitempty"`
}

// BroadcastResponse สำหรับผลลัพธ์การดำเนินการกับ broadcast (สร้าง อัปเดต ดึงข้อมูล)
type BroadcastResponse struct {
	GenericResponse
	Broadcast BroadcastItem `json:"broadcast"`
}

// BroadcastListResponse สำหรับผลลัพธ์การดึงรายการ broadcasts
type BroadcastListResponse struct {
	GenericResponse
	Broadcasts []BroadcastItem `json:"broadcasts"`
	Pagination PaginationData  `json:"pagination"`
}

// BroadcastPreviewResponse สำหรับผลลัพธ์การดูตัวอย่าง broadcast
type BroadcastPreviewResponse struct {
	GenericResponse
	Preview interface{} `json:"preview"`
}

// BroadcastEstimateResponse สำหรับผลลัพธ์การประมาณจำนวนผู้รับ
type BroadcastEstimateResponse struct {
	GenericResponse
	Count int64 `json:"count"`
}

// BroadcastStats สถิติของ broadcast
type BroadcastStats struct {
	TotalTargeted int64   `json:"total_targeted"`
	Pending       int64   `json:"pending"`
	Delivered     int64   `json:"delivered"`
	Failed        int64   `json:"failed"`
	Opened        int64   `json:"opened"`
	Clicked       int64   `json:"clicked"`
	DeliveryRate  float64 `json:"delivery_rate"`
	OpenRate      float64 `json:"open_rate"`
	ClickRate     float64 `json:"click_rate"`
}

// BroadcastStatsResponse สำหรับผลลัพธ์การดึงสถิติของ broadcast
type BroadcastStatsResponse struct {
	GenericResponse
	Stats BroadcastStats `json:"stats"`
}

// BroadcastDeliveriesResponse สำหรับผลลัพธ์การดึงรายการการส่ง broadcast
type BroadcastDeliveriesResponse struct {
	GenericResponse
	Deliveries []BroadcastDeliveryItem `json:"deliveries"`
	Pagination PaginationData          `json:"pagination"`
}
