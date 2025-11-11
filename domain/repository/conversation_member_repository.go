// domain/repository/conversation_member_repository.go
package repository

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

// ConversationMemberRepository อินเตอร์เฟซสำหรับการจัดการสมาชิกในการสนทนา
type ConversationMemberRepository interface {
	Create(member *models.ConversationMember) error
	GetByID(id uuid.UUID) (*models.ConversationMember, error)
	GetByConversationAndUserID(conversationID, userID uuid.UUID) (*models.ConversationMember, error)
	GetMembersByConversationID(conversationID uuid.UUID, page, limit int) ([]*models.ConversationMember, error)
	UpdateRole(id uuid.UUID, role string) error
	Delete(id uuid.UUID) error
	DeleteByConversationAndUserID(conversationID, userID uuid.UUID) error
	GetConversationAdmins(conversationID uuid.UUID) ([]*models.ConversationMember, error)
	CountConversationMembers(conversationID uuid.UUID) (int64, error)
}
