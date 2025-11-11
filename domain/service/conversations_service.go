// domain/service/conversation_service.go
package service

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/dto"
	"github.com/thizplus/gofiber-chat-api/domain/types"
)

type ConversationService interface {
	// CreateDirectConversation สร้างการสนทนาแบบส่วนตัวระหว่างผู้ใช้สองคน
	CreateDirectConversation(userID, friendID uuid.UUID) (*dto.ConversationDTO, error)

	// CreateGroupConversation สร้างการสนทนาแบบกลุ่ม
	CreateGroupConversation(userID uuid.UUID, title, iconURL string, memberIDs []uuid.UUID) (*dto.ConversationDTO, error)

	// CreateBusinessConversation สร้างการสนทนากับธุรกิจ
	CreateBusinessConversation(userID, businessID uuid.UUID) (*dto.ConversationDTO, error)

	// GetUserConversations ดึงรายการการสนทนาทั้งหมดของผู้ใช้
	GetUserConversations(userID uuid.UUID, limit, offset int, convType string, pinned bool) ([]*dto.ConversationDTO, int, error)

	// GetConversationMessages ดึงข้อความทั้งหมดในการสนทนา
	GetConversationMessages(conversationID, userID uuid.UUID, limit, offset int) ([]*dto.MessageDTO, int64, error)

	// SetPinStatus กำหนดสถานะการปักหมุดของการสนทนา
	SetPinStatus(conversationID, userID uuid.UUID, isPinned bool) error

	// SetMuteStatus กำหนดสถานะการปิดเสียงของการสนทนา
	SetMuteStatus(conversationID, userID uuid.UUID, isMuted bool) error

	// CheckMembership ตรวจสอบว่าผู้ใช้เป็นสมาชิกของการสนทนาหรือไม่
	CheckMembership(userID, conversationID uuid.UUID) (bool, error)

	// GetMessageContext ดึงข้อความเป้าหมายพร้อมข้อความก่อนหน้าและถัดไป
	GetMessageContext(conversationID, userID uuid.UUID, targetID string,
		beforeCount, afterCount int) ([]*dto.MessageDTO, bool, bool, error)

	// GetMessagesBeforeID ดึงข้อความที่เก่ากว่า ID ที่ระบุ
	GetMessagesBeforeID(conversationID, userID uuid.UUID, beforeID string,
		limit int) ([]*dto.MessageDTO, int64, error)

	// GetMessagesAfterID ดึงข้อความที่ใหม่กว่า ID ที่ระบุ
	GetMessagesAfterID(conversationID, userID uuid.UUID, afterID string,
		limit int) ([]*dto.MessageDTO, int64, error)

	// GetConversationsBeforeTime ดึงการสนทนาที่เก่ากว่าเวลาที่ระบุ
	GetConversationsBeforeTime(userID uuid.UUID, beforeTime string, limit int, convType string, pinned bool) ([]*dto.ConversationDTO, int, error)

	// GetConversationsAfterTime ดึงการสนทนาที่ใหม่กว่าเวลาที่ระบุ
	GetConversationsAfterTime(userID uuid.UUID, afterTime string, limit int, convType string, pinned bool) ([]*dto.ConversationDTO, int, error)

	// GetConversationsBeforeID ดึงการสนทนาที่เก่ากว่า ID ที่ระบุ
	GetConversationsBeforeID(userID, beforeID uuid.UUID, limit int, convType string, pinned bool) ([]*dto.ConversationDTO, int, error)

	// GetConversationsAfterID ดึงการสนทนาที่ใหม่กว่า ID ที่ระบุ
	GetConversationsAfterID(userID, afterID uuid.UUID, limit int, convType string, pinned bool) ([]*dto.ConversationDTO, int, error)

	// UpdateConversation อัปเดตข้อมูลการสนทนา
	UpdateConversation(id uuid.UUID, updateData types.JSONB) error

	/** ################### FOR BUSINESS ONLY ################## */

	// GetBusinessConversations ดึงการสนทนาทั้งหมดของธุรกิจ (โหมดปกติ)
	GetBusinessConversations(businessID uuid.UUID, adminID uuid.UUID, limit, offset int) ([]*dto.ConversationDTO, int, error)

	// GetBusinessConversationsBeforeTime ดึงการสนทนาธุรกิจที่เก่ากว่าเวลาที่ระบุ
	GetBusinessConversationsBeforeTime(businessID uuid.UUID, adminID uuid.UUID, beforeTime string, limit int) ([]*dto.ConversationDTO, int, error)

	// GetBusinessConversationsAfterTime ดึงการสนทนาธุรกิจที่ใหม่กว่าเวลาที่ระบุ
	GetBusinessConversationsAfterTime(businessID uuid.UUID, adminID uuid.UUID, afterTime string, limit int) ([]*dto.ConversationDTO, int, error)

	// GetBusinessConversationsBeforeID ดึงการสนทนาธุรกิจที่เก่ากว่า ID ที่ระบุ
	GetBusinessConversationsBeforeID(businessID uuid.UUID, adminID uuid.UUID, beforeID uuid.UUID, limit int) ([]*dto.ConversationDTO, int, error)

	// GetBusinessConversationsAfterID ดึงการสนทนาธุรกิจที่ใหม่กว่า ID ที่ระบุ
	GetBusinessConversationsAfterID(businessID uuid.UUID, adminID uuid.UUID, afterID uuid.UUID, limit int) ([]*dto.ConversationDTO, int, error)

	// GetBusinessConversationMessages ดึงข้อความในการสนทนาธุรกิจ (โหมดปกติ)
	GetBusinessConversationMessages(conversationID, businessID uuid.UUID, limit, offset int) ([]*dto.MessageDTO, int64, error)

	// GetBusinessMessageContext ดึงข้อความเป้าหมายพร้อมบริบทสำหรับธุรกิจ (Jump to Message)
	GetBusinessMessageContext(conversationID, businessID uuid.UUID, targetID string, beforeCount, afterCount int) ([]*dto.MessageDTO, bool, bool, error)

	// GetBusinessMessagesBeforeID ดึงข้อความธุรกิจที่เก่ากว่า ID ที่ระบุ
	GetBusinessMessagesBeforeID(conversationID, businessID uuid.UUID, beforeID string, limit int) ([]*dto.MessageDTO, int64, error)

	// GetBusinessMessagesAfterID ดึงข้อความธุรกิจที่ใหม่กว่า ID ที่ระบุ
	GetBusinessMessagesAfterID(conversationID, businessID uuid.UUID, afterID string, limit int) ([]*dto.MessageDTO, int64, error)

	// CheckConversationBelongsToBusiness ตรวจสอบว่าการสนทนาเป็นของธุรกิจ
	CheckConversationBelongsToBusiness(conversationID, businessID uuid.UUID) (bool, error)
}
