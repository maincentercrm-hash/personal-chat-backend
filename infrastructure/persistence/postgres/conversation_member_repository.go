// infrastructure/persistence/postgres/conversation_member_repository.go
package postgres

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"gorm.io/gorm"
)

type conversationMemberRepository struct {
	db *gorm.DB
}

// NewConversationMemberRepository สร้าง repository instance สำหรับจัดการสมาชิกในการสนทนา
func NewConversationMemberRepository(db *gorm.DB) repository.ConversationMemberRepository {
	return &conversationMemberRepository{db}
}

// Create สร้างสมาชิกใหม่ในการสนทนา
func (r *conversationMemberRepository) Create(member *models.ConversationMember) error {
	return r.db.Create(member).Error
}

// GetByID ดึงข้อมูลสมาชิกตาม ID
func (r *conversationMemberRepository) GetByID(id uuid.UUID) (*models.ConversationMember, error) {
	var member models.ConversationMember
	err := r.db.Where("id = ?", id).First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// GetByConversationAndUserID ดึงข้อมูลสมาชิกตาม conversationID และ userID
func (r *conversationMemberRepository) GetByConversationAndUserID(conversationID, userID uuid.UUID) (*models.ConversationMember, error) {
	var member models.ConversationMember
	err := r.db.Where("conversation_id = ? AND user_id = ?", conversationID, userID).First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// GetMembersByConversationID ดึงรายการสมาชิกทั้งหมดในการสนทนา
func (r *conversationMemberRepository) GetMembersByConversationID(conversationID uuid.UUID, page, limit int) ([]*models.ConversationMember, error) {
	var members []*models.ConversationMember
	offset := (page - 1) * limit
	err := r.db.Where("conversation_id = ?", conversationID).Offset(offset).Limit(limit).Find(&members).Error
	if err != nil {
		return nil, err
	}
	return members, nil
}

// UpdateRole อัพเดทบทบาทของสมาชิก
func (r *conversationMemberRepository) UpdateRole(id uuid.UUID, role string) error {
	return r.db.Model(&models.ConversationMember{}).Where("id = ?", id).Update("role", role).Error
}

// Delete ลบสมาชิกตาม ID
func (r *conversationMemberRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.ConversationMember{}, id).Error
}

// DeleteByConversationAndUserID ลบสมาชิกตาม conversationID และ userID
func (r *conversationMemberRepository) DeleteByConversationAndUserID(conversationID, userID uuid.UUID) error {
	return r.db.Where("conversation_id = ? AND user_id = ?", conversationID, userID).Delete(&models.ConversationMember{}).Error
}

// GetConversationAdmins ดึงรายการ admin ทั้งหมดในการสนทนา
func (r *conversationMemberRepository) GetConversationAdmins(conversationID uuid.UUID) ([]*models.ConversationMember, error) {
	var admins []*models.ConversationMember
	err := r.db.Where("conversation_id = ? AND role = ?", conversationID, "admin").Find(&admins).Error
	if err != nil {
		return nil, err
	}
	return admins, nil
}

// CountConversationMembers นับจำนวนสมาชิกในการสนทนา
func (r *conversationMemberRepository) CountConversationMembers(conversationID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.ConversationMember{}).Where("conversation_id = ?", conversationID).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
