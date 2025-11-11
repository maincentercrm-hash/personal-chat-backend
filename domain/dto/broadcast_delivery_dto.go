package dto

import (
	"time"

	"github.com/google/uuid"
)

// ============ Request DTOs ============

// BroadcastDeliveryTrackRequest สำหรับการบันทึกการเปิดอ่านหรือคลิก broadcast
type BroadcastDeliveryTrackRequest struct {
	DeliveryID uuid.UUID `json:"delivery_id" validate:"required"`
}

// BroadcastDeliveryQueryRequest สำหรับการดึงรายการ broadcast deliveries
type BroadcastDeliveryQueryRequest struct {
	Limit int `json:"limit" validate:"omitempty,min=1,max=100"`
	Page  int `json:"page" validate:"omitempty,min=1"`
}

// ============ Response DTOs ============

// BroadcastDeliveryTrackResponse สำหรับผลลัพธ์การบันทึกการเปิดอ่านหรือคลิก
type BroadcastDeliveryTrackResponse struct {
	GenericResponse
}

// BroadcastDeliveryItem รายการ broadcast delivery
type BroadcastDeliveryItem struct {
	ID           uuid.UUID  `json:"id"`
	BroadcastID  uuid.UUID  `json:"broadcast_id"`
	UserID       uuid.UUID  `json:"user_id"`
	DeliveredAt  *time.Time `json:"delivered_at,omitempty"`
	OpenedAt     *time.Time `json:"opened_at,omitempty"`
	ClickedAt    *time.Time `json:"clicked_at,omitempty"`
	Status       string     `json:"status"`
	ErrorMessage *string    `json:"error_message,omitempty"`

	// ข้อมูลเพิ่มเติมที่เกี่ยวข้องกับ broadcast
	BroadcastTitle       string     `json:"broadcast_title,omitempty"`
	BroadcastMessageType string     `json:"broadcast_message_type,omitempty"`
	BroadcastContent     string     `json:"broadcast_content,omitempty"`
	BroadcastMediaURL    *string    `json:"broadcast_media_url,omitempty"`
	BroadcastSentAt      *time.Time `json:"broadcast_sent_at,omitempty"`
	BusinessID           uuid.UUID  `json:"business_id"`
	BusinessName         string     `json:"business_name,omitempty"`
}

// BroadcastDeliveryListResponse สำหรับผลลัพธ์การดึงรายการ broadcast deliveries
type BroadcastDeliveryListResponse struct {
	GenericResponse
	Deliveries []BroadcastDeliveryItem `json:"deliveries"`
	Pagination PaginationData          `json:"pagination"`
}
