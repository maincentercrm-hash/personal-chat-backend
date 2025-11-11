// domain/dto/message_read_dto.go
package dto

import (
	"time"

	"github.com/google/uuid"
)

// =====================================
// Request DTOs
// =====================================

// MarkMessageAsReadRequest เป็น DTO สำหรับคำขอมาร์คข้อความเป็นอ่านแล้ว
type MarkMessageAsReadRequest struct {
	MessageID uuid.UUID `json:"message_id" validate:"required"`
}

// GetMessageReadsRequest เป็น DTO สำหรับคำขอดูรายชื่อผู้ที่อ่านข้อความแล้ว
type GetMessageReadsRequest struct {
	MessageID uuid.UUID `json:"message_id" validate:"required"`
}

// MarkAllMessagesAsReadRequest เป็น DTO สำหรับคำขอมาร์คข้อความทั้งหมดในการสนทนาเป็นอ่านแล้ว
type MarkAllMessagesAsReadRequest struct {
	ConversationID uuid.UUID `json:"conversation_id" validate:"required"`
}

// GetUnreadCountRequest เป็น DTO สำหรับคำขอดูจำนวนข้อความที่ยังไม่ได้อ่านในการสนทนา
type GetUnreadCountRequest struct {
	ConversationID uuid.UUID `json:"conversation_id" validate:"required"`
}

// =====================================
// Response DTOs
// =====================================

// MessageReadInfo เป็นข้อมูลเกี่ยวกับการอ่านข้อความของผู้ใช้
type MessageReadInfo struct {
	UserID    string    `json:"user_id"`
	ReadAt    time.Time `json:"read_at"`
	UserName  string    `json:"user_name,omitempty"`  // ชื่อผู้ใช้ (ถ้ามี)
	AvatarURL string    `json:"avatar_url,omitempty"` // URL รูปโปรไฟล์ (ถ้ามี)
	IsOnline  bool      `json:"is_online,omitempty"`  // สถานะออนไลน์ (ถ้ามี)
	LastSeen  time.Time `json:"last_seen,omitempty"`  // เวลาล่าสุดที่ออนไลน์ (ถ้ามี)
}

// MarkMessageAsReadResponse เป็น DTO สำหรับการตอบกลับการมาร์คข้อความเป็นอ่านแล้ว
type MarkMessageAsReadResponse struct {
	GenericResponse
}

// GetMessageReadsResponse เป็น DTO สำหรับการตอบกลับรายชื่อผู้ที่อ่านข้อความแล้ว
type GetMessageReadsResponse struct {
	GenericResponse
	Data struct {
		Reads []MessageReadInfo `json:"reads"`
		// สามารถเพิ่มข้อมูลเพิ่มเติมได้ เช่น
		TotalReads int `json:"total_reads"`
	} `json:"data"`
}

// MarkAllMessagesAsReadResponse เป็น DTO สำหรับการตอบกลับการมาร์คข้อความทั้งหมดเป็นอ่านแล้ว
type MarkAllMessagesAsReadResponse struct {
	GenericResponse
	Data struct {
		MarkedCount int `json:"marked_count"` // จำนวนข้อความที่ถูกมาร์คว่าอ่านแล้ว
	} `json:"data"`
}

// GetUnreadCountResponse เป็น DTO สำหรับการตอบกลับจำนวนข้อความที่ยังไม่ได้อ่าน
type GetUnreadCountResponse struct {
	GenericResponse
	Data struct {
		UnreadCount int `json:"unread_count"` // จำนวนข้อความที่ยังไม่ได้อ่าน
	} `json:"data"`
}

// MessageReadSummary เป็น DTO สำหรับสรุปข้อมูลการอ่านของแต่ละข้อความ
type MessageReadSummary struct {
	MessageID   string    `json:"message_id"`
	TotalReads  int       `json:"total_reads"`
	LatestRead  time.Time `json:"latest_read,omitempty"`
	IsReadByAll bool      `json:"is_read_by_all"`
}

// ConversationReadStatus เป็น DTO สำหรับสรุปสถานะการอ่านของการสนทนา
type ConversationReadStatus struct {
	ConversationID string    `json:"conversation_id"`
	TotalMessages  int       `json:"total_messages"`
	ReadMessages   int       `json:"read_messages"`
	UnreadMessages int       `json:"unread_messages"`
	LastReadAt     time.Time `json:"last_read_at,omitempty"`
}
