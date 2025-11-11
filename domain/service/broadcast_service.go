// domain/service/broadcast_service.go
package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/dto"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/types"
)

// BroadcastService interface สำหรับจัดการบริการ broadcast
type BroadcastService interface {
	// CreateBroadcast สร้าง broadcast ใหม่
	CreateBroadcast(
		businessID, createdByID uuid.UUID,
		title, messageType, content string,
		mediaURL string,
		bubbleType string,
		bubbleData types.JSONB,
	) (*models.Broadcast, error)

	// GetBroadcastByID ดึงข้อมูล broadcast ตาม ID
	GetBroadcastByID(id, businessID, requestedByID uuid.UUID) (*models.Broadcast, error)

	// GetBusinessBroadcasts ดึงข้อมูล broadcasts ทั้งหมดของธุรกิจ
	GetBusinessBroadcasts(businessID, requestedByID uuid.UUID, status string, limit, offset int) ([]*models.Broadcast, int64, error)

	// UpdateBroadcast อัพเดทข้อมูล broadcast
	UpdateBroadcast(id, businessID, requestedByID uuid.UUID, updateData types.JSONB) (*models.Broadcast, error)

	// DeleteBroadcast ลบ broadcast
	DeleteBroadcast(id, businessID, requestedByID uuid.UUID) error

	// ScheduleBroadcast กำหนดเวลาส่ง broadcast
	ScheduleBroadcast(id, businessID, requestedByID uuid.UUID, scheduledAt time.Time) error

	// CancelScheduledBroadcast ยกเลิกการกำหนดเวลาส่ง
	CancelScheduledBroadcast(id, businessID, requestedByID uuid.UUID) error

	// SendBroadcast ส่ง broadcast ทันที
	SendBroadcast(id, businessID, requestedByID uuid.UUID) error

	// GetBroadcastStats ดึงสถิติของ broadcast
	GetBroadcastStats(id, businessID, requestedByID uuid.UUID) (*dto.BroadcastStats, error)

	// GetBroadcastDeliveries ดึงรายการส่ง broadcast
	GetBroadcastDeliveries(id, businessID, requestedByID uuid.UUID, status string, limit, offset int) ([]*models.BroadcastDelivery, int64, error)

	// SearchBroadcasts ค้นหา broadcasts ตามเงื่อนไข
	SearchBroadcasts(businessID, requestedByID uuid.UUID, query, messageType, status string, startDate, endDate time.Time, limit, offset int) ([]*models.Broadcast, int64, error)

	// SetTargetAll กำหนดให้ส่งถึงผู้ติดตามทุกคน
	SetTargetAll(id, businessID, requestedByID uuid.UUID) error

	// SetTargetTags กำหนดให้ส่งถึงผู้ใช้ตาม tags
	SetTargetTags(id, businessID, requestedByID uuid.UUID, includeTags, excludeTags []uuid.UUID, matchType string) error

	// SetTargetUsers กำหนดให้ส่งถึงผู้ใช้เฉพาะราย
	SetTargetUsers(id, businessID, requestedByID uuid.UUID, userIDs []uuid.UUID) error

	// SetTargetCustomerProfile กำหนดให้ส่งถึงผู้ใช้ตามข้อมูล customer profile
	SetTargetCustomerProfile(id, businessID, requestedByID uuid.UUID, criteria *dto.BroadcastTargetCriteria) error

	// GetEstimatedTargetCount ประมาณจำนวนผู้รับตามเงื่อนไขที่กำหนด
	GetEstimatedTargetCount(businessID, requestedByID uuid.UUID, targetType string, targetCriteria *dto.BroadcastTargetCriteria) (int64, error)

	// PreviewBroadcast ดูตัวอย่างข้อความก่อนส่ง
	PreviewBroadcast(id, businessID, requestedByID uuid.UUID) (types.JSONB, error)

	// TrackBroadcastOpen บันทึกการเปิดอ่าน broadcast
	TrackBroadcastOpen(broadcastID, userID uuid.UUID) error

	// TrackBroadcastClick บันทึกการคลิก broadcast
	TrackBroadcastClick(broadcastID, userID uuid.UUID) error

	// ValidateBroadcastContent ตรวจสอบความถูกต้องของเนื้อหา broadcast
	ValidateBroadcastContent(messageType string, content string, mediaURL string, bubbleType string, bubbleData types.JSONB) error

	// DuplicateBroadcast สร้าง broadcast ใหม่โดยคัดลอกจาก broadcast เดิม
	DuplicateBroadcast(sourceID, businessID, requestedByID uuid.UUID) (*models.Broadcast, error)
}
