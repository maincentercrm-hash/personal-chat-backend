// domain/service/conversation_member_service.go
package service

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/dto"
)

// ConversationMemberService interface สำหรับจัดการสมาชิกในการสนทนา
type ConversationMemberService interface {
	// AddMember เพิ่มสมาชิกในการสนทนากลุ่ม
	AddMember(userID, conversationID, newMemberID uuid.UUID) (*dto.MemberDTO, error)

	// GetMembers ดึงรายการสมาชิกในการสนทนา
	GetMembers(userID, conversationID uuid.UUID, page, limit int) ([]*dto.MemberDTO, int, error)

	// RemoveMember ลบสมาชิกออกจากการสนทนา
	RemoveMember(userID, conversationID, memberToRemoveID uuid.UUID) error

	// ToggleAdminStatus เปลี่ยนสถานะแอดมินของสมาชิก
	ToggleAdminStatus(userID, conversationID, targetUserID uuid.UUID, isAdmin bool) (bool, error)

	//ค้นหาการสนทนาแบบ direct ระหว่างผู้ใช้สองคน
	FindDirectConversationBetweenUsers(userID, friendID uuid.UUID) (uuid.UUID, error)
}
