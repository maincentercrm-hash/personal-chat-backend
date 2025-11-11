// domain/service/business_welcome_message_service.go
package service

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/types"
)

// BusinessWelcomeMessageService คือ interface สำหรับบริการจัดการ welcome message
type BusinessWelcomeMessageService interface {
	// CreateWelcomeMessage สร้าง welcome message ใหม่
	CreateWelcomeMessage(
		businessID uuid.UUID,
		userID uuid.UUID,
		messageType string,
		title string,
		content string,
		imageURL string,
		thumbnailURL string,
		actionButtons types.JSONB,
		components types.JSONB,
		triggerType string,
		triggerParams types.JSONB,
		sortOrder int,
	) (*models.BusinessWelcomeMessage, error)

	// GetWelcomeMessageByID ดึงข้อมูล welcome message ตาม ID
	GetWelcomeMessageByID(id uuid.UUID, userID uuid.UUID) (*models.BusinessWelcomeMessage, error)

	// GetBusinessWelcomeMessages ดึงข้อมูล welcome message ทั้งหมดของธุรกิจ
	GetBusinessWelcomeMessages(businessID uuid.UUID, userID uuid.UUID, includeInactive bool) ([]*models.BusinessWelcomeMessage, error)

	// UpdateWelcomeMessage อัพเดทข้อมูล welcome message
	UpdateWelcomeMessage(id uuid.UUID, businessID uuid.UUID, userID uuid.UUID, updateData types.JSONB) (*models.BusinessWelcomeMessage, error)

	// DeleteWelcomeMessage ลบ welcome message
	DeleteWelcomeMessage(id uuid.UUID, businessID uuid.UUID, userID uuid.UUID) error

	// SetWelcomeMessageActive กำหนดสถานะการใช้งานของ welcome message
	SetWelcomeMessageActive(id uuid.UUID, businessID uuid.UUID, userID uuid.UUID, isActive bool) error

	// UpdateWelcomeMessageSortOrder อัพเดทลำดับการแสดงผลของ welcome message
	UpdateWelcomeMessageSortOrder(id uuid.UUID, businessID uuid.UUID, userID uuid.UUID, sortOrder int) error

	// GetWelcomeMessagesByTriggerType ดึงข้อมูล welcome message ตามประเภททริกเกอร์
	GetWelcomeMessagesByTriggerType(businessID uuid.UUID, userID uuid.UUID, triggerType string) ([]*models.BusinessWelcomeMessage, error)

	// ProcessFollowWelcomeMessages ประมวลผลและส่ง welcome message เมื่อผู้ใช้เริ่มติดตามธุรกิจ
	ProcessFollowWelcomeMessages(businessID uuid.UUID, targetUserID uuid.UUID) error

	// ProcessCommandWelcomeMessages ประมวลผลและส่ง welcome message เมื่อผู้ใช้ส่งคำสั่งเฉพาะ
	ProcessCommandWelcomeMessages(businessID uuid.UUID, targetUserID uuid.UUID, command string) ([]*models.BusinessWelcomeMessage, error)

	// ProcessConversationStartWelcomeMessages ประมวลผลและส่ง welcome message เมื่อผู้ใช้เริ่มต้นสนทนา
	ProcessConversationStartWelcomeMessages(businessID uuid.UUID, targetUserID uuid.UUID, conversationID uuid.UUID) error

	// ProcessInactiveWelcomeMessages ประมวลผลและส่ง welcome message เมื่อผู้ใช้กลับมาหลังจากไม่มีกิจกรรม
	ProcessInactiveWelcomeMessages(businessID uuid.UUID, targetUserID uuid.UUID, inactiveDays int) error

	// TrackMessageSent บันทึกการส่ง welcome message
	TrackMessageSent(messageID uuid.UUID) error

	// TrackMessageClick บันทึกการคลิกในแต่ละแอคชั่นของ welcome message
	TrackMessageClick(messageID uuid.UUID, actionType string, actionData types.JSONB) error

	// TrackMessageReply บันทึกการตอบกลับ welcome message
	TrackMessageReply(messageID uuid.UUID) error

	// RenderWelcomeMessageContent เรนเดอร์เนื้อหา welcome message พร้อมแทนที่ตัวแปร
	RenderWelcomeMessageContent(message *models.BusinessWelcomeMessage, targetUserID uuid.UUID) (types.JSONB, error)

	// ValidateTriggerParams ตรวจสอบความถูกต้องของพารามิเตอร์ทริกเกอร์
	ValidateTriggerParams(triggerType string, triggerParams types.JSONB) error

	// ValidateMessageComponents ตรวจสอบความถูกต้องของคอมโพเนนต์ข้อความ
	ValidateMessageComponents(messageType string, components types.JSONB) error

	// ValidateActionButtons ตรวจสอบความถูกต้องของปุ่มดำเนินการ
	ValidateActionButtons(actionButtons types.JSONB) error
}
