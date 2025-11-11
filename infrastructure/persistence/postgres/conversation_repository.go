// infrastructure/persistence/postgres/conversation_repository.go
package postgres

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"github.com/thizplus/gofiber-chat-api/domain/types"
	"gorm.io/gorm"
)

type conversationRepository struct {
	db *gorm.DB
}

// NewConversationRepository สร้าง repository ใหม่
func NewConversationRepository(db *gorm.DB) repository.ConversationRepository {
	return &conversationRepository{
		db: db,
	}
}

// GetByID ดึงข้อมูลการสนทนาตาม ID
func (r *conversationRepository) GetByID(id uuid.UUID) (*models.Conversation, error) {
	var conversation models.Conversation
	if err := r.db.First(&conversation, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &conversation, nil
}

// Create สร้างการสนทนาใหม่
func (r *conversationRepository) Create(conversation *models.Conversation) error {
	return r.db.Create(conversation).Error
}

// AddMember เพิ่มสมาชิกในการสนทนา
func (r *conversationRepository) AddMember(member *models.ConversationMember) error {
	return r.db.Create(member).Error
}

// GetMembers ดึงรายการสมาชิกทั้งหมดในการสนทนา
func (r *conversationRepository) GetMembers(conversationID uuid.UUID) ([]*models.ConversationMember, error) {
	var members []*models.ConversationMember
	if err := r.db.Where("conversation_id = ?", conversationID).Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

// GetMember ดึงข้อมูลสมาชิกในการสนทนา
func (r *conversationRepository) GetMember(conversationID, userID uuid.UUID) (*models.ConversationMember, error) {
	var member models.ConversationMember
	if err := r.db.Where("conversation_id = ? AND user_id = ?", conversationID, userID).First(&member).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

// FindDirectConversation หาการสนทนาโดยตรงระหว่างผู้ใช้สองคน
func (r *conversationRepository) FindDirectConversation(user1ID, user2ID uuid.UUID) (*models.Conversation, error) {
	// หา conversation IDs ที่ user1 เป็นสมาชิก
	var user1Memberships []models.ConversationMember
	if err := r.db.Where("user_id = ?", user1ID).Find(&user1Memberships).Error; err != nil {
		return nil, err
	}

	if len(user1Memberships) == 0 {
		return nil, nil
	}

	// สร้าง slice ของ conversation IDs
	var convIDs []uuid.UUID
	for _, m := range user1Memberships {
		convIDs = append(convIDs, m.ConversationID)
	}

	// หา conversation IDs ที่ user2 เป็นสมาชิก และมี type เป็น direct
	var directConversations []models.Conversation
	if err := r.db.Where("id IN ? AND type = ?", convIDs, "direct").Find(&directConversations).Error; err != nil {
		return nil, err
	}

	// ตรวจสอบแต่ละการสนทนาว่ามี user2 เป็นสมาชิกหรือไม่
	for _, conv := range directConversations {
		var members []models.ConversationMember
		if err := r.db.Where("conversation_id = ?", conv.ID).Find(&members).Error; err != nil {
			continue
		}

		// ตรวจสอบว่ามีแค่ 2 คนและมี user2 เป็นสมาชิก
		if len(members) == 2 {
			for _, member := range members {
				if member.UserID == user2ID {
					return &conv, nil
				}
			}
		}
	}

	return nil, nil
}

// GetUserConversations ดึงการสนทนาทั้งหมดของผู้ใช้ (ยกเว้น business conversations ที่ user เป็น admin)
func (r *conversationRepository) GetUserConversations(userID uuid.UUID, limit, offset int) ([]*models.Conversation, int, error) {
	// 1. หา conversation IDs ที่ผู้ใช้เป็นสมาชิก
	var memberships []models.ConversationMember
	if err := r.db.Where("user_id = ?", userID).Find(&memberships).Error; err != nil {
		return nil, 0, err
	}

	if len(memberships) == 0 {
		return []*models.Conversation{}, 0, nil
	}

	// 2. สร้าง slice ของ conversation IDs
	var convIDs []uuid.UUID
	for _, m := range memberships {
		convIDs = append(convIDs, m.ConversationID)
	}

	// 3. ✨ แก้ไขตรงนี้: ไม่รวม business conversations ที่ user เป็น admin/owner
	baseQuery := r.db.Model(&models.Conversation{}).
		Where("id IN ? AND is_active = ?", convIDs, true)

	// ✨ เพิ่มเงื่อนไข: กรอง business conversations ที่ user เป็น owner หรือ admin
	baseQuery = baseQuery.Where(`
		NOT (
			type = 'business' AND (
				creator_id = ? OR 
				business_id IN (
					SELECT id FROM business_accounts 
					WHERE creator_id = ? OR 
					id IN (
						SELECT business_id FROM business_admins 
						WHERE user_id = ? AND is_active = true
					)
				)
			)
		)
	`, userID, userID, userID)

	// 4. นับจำนวนทั้งหมด
	var count int64
	if err := baseQuery.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// 5. ดึงข้อมูลการสนทนา
	var conversations []*models.Conversation
	err := baseQuery.
		Order("COALESCE(last_message_at, updated_at) DESC").
		Limit(limit).
		Offset(offset).
		Find(&conversations).Error
	if err != nil {
		return nil, 0, err
	}

	return conversations, int(count), nil
}

// GetUserMemberships ดึงข้อมูลสมาชิกการสนทนาทั้งหมดของผู้ใช้
func (r *conversationRepository) GetUserMemberships(userID uuid.UUID) ([]*models.ConversationMember, error) {
	var memberships []*models.ConversationMember
	if err := r.db.Where("user_id = ?", userID).Find(&memberships).Error; err != nil {
		return nil, err
	}
	return memberships, nil
}

// GetConversationsByIDs ดึงการสนทนาจาก IDs
func (r *conversationRepository) GetConversationsByIDs(ids []uuid.UUID) ([]*models.Conversation, error) {
	if len(ids) == 0 {
		return []*models.Conversation{}, nil
	}

	var conversations []*models.Conversation
	if err := r.db.Where("id IN ?", ids).Find(&conversations).Error; err != nil {
		return nil, err
	}
	return conversations, nil
}

// UpdateLastMessage อัพเดต last_message สำหรับการสนทนา
func (r *conversationRepository) UpdateLastMessage(conversationID uuid.UUID, text string, messageTime time.Time) error {
	return r.db.Model(&models.Conversation{}).
		Where("id = ?", conversationID).
		Updates(types.JSONB{
			"last_message_text": text,
			"last_message_at":   messageTime,
			"updated_at":        time.Now(),
		}).Error
}

// SetPinStatus กำหนดสถานะการปักหมุดของการสนทนา
func (r *conversationRepository) SetPinStatus(conversationID, userID uuid.UUID, isPinned bool) error {
	result := r.db.Model(&models.ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Update("is_pinned", isPinned)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("conversation member not found")
	}
	return nil
}

// SetMuteStatus กำหนดสถานะการปิดเสียงของการสนทนา
func (r *conversationRepository) SetMuteStatus(conversationID, userID uuid.UUID, isMuted bool) error {
	result := r.db.Model(&models.ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Update("is_muted", isMuted)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("conversation member not found")
	}
	return nil
}

// MarkAllMessagesAsRead มาร์คข้อความทั้งหมดในการสนทนาว่าอ่านแล้ว
func (r *conversationRepository) MarkAllMessagesAsRead(conversationID, userID uuid.UUID) error {
	// อัพเดท last_read_at ในตาราง conversation_members
	now := time.Now()
	err := r.db.Model(&models.ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Update("last_read_at", now).Error

	return err
}

// RemoveMember ลบสมาชิกออกจากการสนทนา
func (r *conversationRepository) RemoveMember(conversationID, userID uuid.UUID) error {
	result := r.db.Delete(&models.ConversationMember{}, "conversation_id = ? AND user_id = ?", conversationID, userID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("conversation member not found")
	}
	return nil
}

// UpdateMemberAdmin อัพเดตสถานะแอดมินของสมาชิก
func (r *conversationRepository) UpdateMemberAdmin(conversationID, userID uuid.UUID, isAdmin bool) error {
	result := r.db.Model(&models.ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Update("is_admin", isAdmin)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("conversation member not found")
	}
	return nil
}

// Delete ลบการสนทนา (soft delete - เปลี่ยนสถานะเป็น inactive)
func (r *conversationRepository) Delete(id uuid.UUID) error {
	result := r.db.Model(&models.Conversation{}).
		Where("id = ?", id).
		Update("is_active", false)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("conversation not found")
	}
	return nil
}

// UpdateConversation อัปเดตข้อมูลการสนทนา
func (r *conversationRepository) UpdateConversation(id uuid.UUID, updateData types.JSONB) error {
	// ใช้ฟังก์ชันใหม่เพื่อแปลงข้อมูลให้ปลอดภัยสำหรับ GORM
	updates := updateData.SafeForGorm()

	result := r.db.Model(&models.Conversation{}).Where("id = ?", id).Updates(updates)
	return result.Error
}

func (r *conversationRepository) UpdateMemberLastRead(conversationID uuid.UUID, userID uuid.UUID, readTime time.Time) error {
	// สร้างคำสั่ง SQL หรือใช้ ORM เพื่ออัปเดตข้อมูล
	result := r.db.Model(&models.ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Update("last_read_at", readTime)

	if result.Error != nil {
		return result.Error
	}

	// ตรวจสอบว่ามีการอัปเดตจริงหรือไม่
	if result.RowsAffected == 0 {
		// อาจจำเป็นต้องสร้างบันทึกใหม่ถ้ายังไม่มี
		member := &models.ConversationMember{
			ID:             uuid.New(),
			ConversationID: conversationID,
			UserID:         userID,
			LastReadAt:     &readTime,
			JoinedAt:       time.Now(),
		}

		if err := r.db.Create(member).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *conversationRepository) IsMember(conversationID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", conversationID, userID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *conversationRepository) GetLastMessage(conversationID uuid.UUID) (*models.Message, error) {
	var message models.Message
	err := r.db.Where("conversation_id = ?", conversationID).
		Order("created_at DESC").
		First(&message).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &message, nil
}

func (r *conversationRepository) GetLastNonDeletedMessage(conversationID uuid.UUID) (*models.Message, error) {
	var message models.Message
	err := r.db.Where("conversation_id = ? AND is_deleted = ?", conversationID, false).
		Order("created_at DESC").
		First(&message).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &message, nil
}

// GetConversationsAfterID ดึงการสนทนาที่ใหม่กว่า ID ที่ระบุ
func (r *conversationRepository) GetConversationsAfterID(userID, afterID uuid.UUID, limit int, convType string, pinned bool) ([]*models.Conversation, int, error) {
	// ดึงการสนทนาเป้าหมายเพื่อดูเวลาของมัน
	var targetConversation models.Conversation
	err := r.db.First(&targetConversation, "id = ?", afterID).Error
	if err != nil {
		return nil, 0, fmt.Errorf("error fetching target conversation: %w", err)
	}

	// ดึงรายการการสนทนาที่ผู้ใช้เป็นสมาชิก
	var memberIDs []uuid.UUID
	err = r.db.Model(&models.ConversationMember{}).
		Select("conversation_id").
		Where("user_id = ?", userID).
		Find(&memberIDs).Error
	if err != nil {
		return nil, 0, err
	}

	if len(memberIDs) == 0 {
		return []*models.Conversation{}, 0, nil
	}

	// เริ่มสร้าง query
	baseQuery := r.db.Model(&models.Conversation{}).
		Where("id IN (?) AND is_active = ?", memberIDs, true)

	// ใช้ LastMessageAt หรือ UpdatedAt เพื่อเปรียบเทียบ
	// ถ้า LastMessageAt มีค่า ใช้มัน แต่ถ้าไม่มี ใช้ UpdatedAt แทน
	var timeCondition string
	var args []interface{}

	if targetConversation.LastMessageAt != nil {
		timeCondition = "(COALESCE(last_message_at, updated_at) > ? OR (COALESCE(last_message_at, updated_at) = ? AND id > ?))"
		args = []interface{}{targetConversation.LastMessageAt, targetConversation.LastMessageAt, afterID}
	} else {
		timeCondition = "(COALESCE(last_message_at, updated_at) > ? OR (COALESCE(last_message_at, updated_at) = ? AND id > ?))"
		args = []interface{}{targetConversation.UpdatedAt, targetConversation.UpdatedAt, afterID}
	}

	baseQuery = baseQuery.Where(timeCondition, args...)

	// เพิ่มเงื่อนไขเพิ่มเติมตามพารามิเตอร์
	if convType != "" {
		baseQuery = baseQuery.Where("type = ?", convType)
	}

	// สร้าง subquery สำหรับการดึงการสนทนาที่ปักหมุด
	if pinned {
		pinnedIDs := []uuid.UUID{}
		subQuery := r.db.Model(&models.ConversationMember{}).
			Select("conversation_id").
			Where("user_id = ? AND is_pinned = ?", userID, true)

		if err := subQuery.Find(&pinnedIDs).Error; err != nil {
			return nil, 0, err
		}

		if len(pinnedIDs) > 0 {
			baseQuery = baseQuery.Where("id IN ?", pinnedIDs)
		} else {
			// ถ้าไม่มีการสนทนาที่ปักหมุด และต้องการเฉพาะที่ปักหมุด
			return []*models.Conversation{}, 0, nil
		}
	}

	// นับจำนวนทั้งหมด
	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลตามเงื่อนไข
	var conversations []*models.Conversation
	err = baseQuery.
		Order("COALESCE(last_message_at, updated_at) ASC").
		Limit(limit).
		Find(&conversations).Error
	if err != nil {
		return nil, 0, err
	}

	// กลับลำดับให้เป็น DESC (จากใหม่ไปเก่า)
	for i := 0; i < len(conversations)/2; i++ {
		j := len(conversations) - i - 1
		conversations[i], conversations[j] = conversations[j], conversations[i]
	}

	return conversations, int(total), nil
}

// GetConversationsBeforeID ดึงการสนทนาที่เก่ากว่า ID ที่ระบุ
func (r *conversationRepository) GetConversationsBeforeID(userID, beforeID uuid.UUID, limit int, convType string, pinned bool) ([]*models.Conversation, int, error) {
	// ดึงการสนทนาเป้าหมายเพื่อดูเวลาของมัน
	var targetConversation models.Conversation
	err := r.db.First(&targetConversation, "id = ?", beforeID).Error
	if err != nil {
		return nil, 0, fmt.Errorf("error fetching target conversation: %w", err)
	}

	// ดึงรายการการสนทนาที่ผู้ใช้เป็นสมาชิก
	var memberIDs []uuid.UUID
	err = r.db.Model(&models.ConversationMember{}).
		Select("conversation_id").
		Where("user_id = ?", userID).
		Find(&memberIDs).Error
	if err != nil {
		return nil, 0, err
	}

	if len(memberIDs) == 0 {
		return []*models.Conversation{}, 0, nil
	}

	// เริ่มสร้าง query
	baseQuery := r.db.Model(&models.Conversation{}).
		Where("id IN (?) AND is_active = ?", memberIDs, true)

	// ใช้ LastMessageAt หรือ UpdatedAt เพื่อเปรียบเทียบ
	var timeCondition string
	var args []interface{}

	if targetConversation.LastMessageAt != nil {
		timeCondition = "(COALESCE(last_message_at, updated_at) < ? OR (COALESCE(last_message_at, updated_at) = ? AND id < ?))"
		args = []interface{}{targetConversation.LastMessageAt, targetConversation.LastMessageAt, beforeID}
	} else {
		timeCondition = "(COALESCE(last_message_at, updated_at) < ? OR (COALESCE(last_message_at, updated_at) = ? AND id < ?))"
		args = []interface{}{targetConversation.UpdatedAt, targetConversation.UpdatedAt, beforeID}
	}

	baseQuery = baseQuery.Where(timeCondition, args...)

	// เพิ่มเงื่อนไขเพิ่มเติมตามพารามิเตอร์
	if convType != "" {
		baseQuery = baseQuery.Where("type = ?", convType)
	}

	// สร้าง subquery สำหรับการดึงการสนทนาที่ปักหมุด
	if pinned {
		pinnedIDs := []uuid.UUID{}
		subQuery := r.db.Model(&models.ConversationMember{}).
			Select("conversation_id").
			Where("user_id = ? AND is_pinned = ?", userID, true)

		if err := subQuery.Find(&pinnedIDs).Error; err != nil {
			return nil, 0, err
		}

		if len(pinnedIDs) > 0 {
			baseQuery = baseQuery.Where("id IN ?", pinnedIDs)
		} else {
			// ถ้าไม่มีการสนทนาที่ปักหมุด และต้องการเฉพาะที่ปักหมุด
			return []*models.Conversation{}, 0, nil
		}
	}

	// นับจำนวนทั้งหมด
	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลตามเงื่อนไข
	var conversations []*models.Conversation
	err = baseQuery.
		Order("COALESCE(last_message_at, updated_at) DESC").
		Limit(limit).
		Find(&conversations).Error
	if err != nil {
		return nil, 0, err
	}

	return conversations, int(total), nil
}

// GetConversationsBeforeTime ดึงการสนทนาที่เก่ากว่าเวลาที่ระบุ
func (r *conversationRepository) GetConversationsBeforeTime(userID uuid.UUID, beforeTime time.Time, limit int, convType string, pinned bool) ([]*models.Conversation, int, error) {
	// ดึงรายการการสนทนาที่ผู้ใช้เป็นสมาชิก
	var memberIDs []uuid.UUID
	err := r.db.Model(&models.ConversationMember{}).
		Select("conversation_id").
		Where("user_id = ?", userID).
		Find(&memberIDs).Error
	if err != nil {
		return nil, 0, err
	}

	if len(memberIDs) == 0 {
		return []*models.Conversation{}, 0, nil
	}

	// เริ่มสร้าง query
	baseQuery := r.db.Model(&models.Conversation{}).
		Where("id IN (?) AND is_active = ?", memberIDs, true)

	// เพิ่มเงื่อนไขเวลา - ใช้ COALESCE เพื่อใช้ last_message_at ถ้ามี แต่ถ้าไม่มีให้ใช้ updated_at
	baseQuery = baseQuery.Where("COALESCE(last_message_at, updated_at) < ?", beforeTime)

	// เพิ่มเงื่อนไขเพิ่มเติมตามพารามิเตอร์
	if convType != "" {
		baseQuery = baseQuery.Where("type = ?", convType)
	}

	// สร้าง subquery สำหรับการดึงการสนทนาที่ปักหมุด
	if pinned {
		pinnedIDs := []uuid.UUID{}
		subQuery := r.db.Model(&models.ConversationMember{}).
			Select("conversation_id").
			Where("user_id = ? AND is_pinned = ?", userID, true)

		if err := subQuery.Find(&pinnedIDs).Error; err != nil {
			return nil, 0, err
		}

		if len(pinnedIDs) > 0 {
			baseQuery = baseQuery.Where("id IN ?", pinnedIDs)
		} else {
			// ถ้าไม่มีการสนทนาที่ปักหมุด และต้องการเฉพาะที่ปักหมุด
			return []*models.Conversation{}, 0, nil
		}
	}

	// นับจำนวนทั้งหมด
	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลตามเงื่อนไข
	var conversations []*models.Conversation
	err = baseQuery.
		Order("COALESCE(last_message_at, updated_at) DESC").
		Limit(limit).
		Find(&conversations).Error
	if err != nil {
		return nil, 0, err
	}

	return conversations, int(total), nil
}

// GetConversationsAfterTime ดึงการสนทนาที่ใหม่กว่าเวลาที่ระบุ
func (r *conversationRepository) GetConversationsAfterTime(userID uuid.UUID, afterTime time.Time, limit int, convType string, pinned bool) ([]*models.Conversation, int, error) {
	// ดึงรายการการสนทนาที่ผู้ใช้เป็นสมาชิก
	var memberIDs []uuid.UUID
	err := r.db.Model(&models.ConversationMember{}).
		Select("conversation_id").
		Where("user_id = ?", userID).
		Find(&memberIDs).Error
	if err != nil {
		return nil, 0, err
	}

	if len(memberIDs) == 0 {
		return []*models.Conversation{}, 0, nil
	}

	// เริ่มสร้าง query
	baseQuery := r.db.Model(&models.Conversation{}).
		Where("id IN (?) AND is_active = ?", memberIDs, true)

	// เพิ่มเงื่อนไขเวลา - ใช้ COALESCE เพื่อใช้ last_message_at ถ้ามี แต่ถ้าไม่มีให้ใช้ updated_at
	baseQuery = baseQuery.Where("COALESCE(last_message_at, updated_at) > ?", afterTime)

	// เพิ่มเงื่อนไขเพิ่มเติมตามพารามิเตอร์
	if convType != "" {
		baseQuery = baseQuery.Where("type = ?", convType)
	}

	// สร้าง subquery สำหรับการดึงการสนทนาที่ปักหมุด
	if pinned {
		pinnedIDs := []uuid.UUID{}
		subQuery := r.db.Model(&models.ConversationMember{}).
			Select("conversation_id").
			Where("user_id = ? AND is_pinned = ?", userID, true)

		if err := subQuery.Find(&pinnedIDs).Error; err != nil {
			return nil, 0, err
		}

		if len(pinnedIDs) > 0 {
			baseQuery = baseQuery.Where("id IN ?", pinnedIDs)
		} else {
			// ถ้าไม่มีการสนทนาที่ปักหมุด และต้องการเฉพาะที่ปักหมุด
			return []*models.Conversation{}, 0, nil
		}
	}

	// นับจำนวนทั้งหมด
	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลตามเงื่อนไข
	var conversations []*models.Conversation
	err = baseQuery.
		Order("COALESCE(last_message_at, updated_at) ASC").
		Limit(limit).
		Find(&conversations).Error
	if err != nil {
		return nil, 0, err
	}

	// กลับลำดับให้เป็น DESC (จากใหม่ไปเก่า)
	for i := 0; i < len(conversations)/2; i++ {
		j := len(conversations) - i - 1
		conversations[i], conversations[j] = conversations[j], conversations[i]
	}

	return conversations, int(total), nil
}

// GetUserConversationsWithFilter ดึงการสนทนาทั้งหมดของผู้ใช้พร้อมตัวกรอง
func (r *conversationRepository) GetUserConversationsWithFilter(userID uuid.UUID, limit, offset int, convType string, pinned bool) ([]*models.Conversation, int, error) {
	// ดึงรายการการสนทนาที่ผู้ใช้เป็นสมาชิก
	var memberIDs []uuid.UUID
	err := r.db.Model(&models.ConversationMember{}).
		Select("conversation_id").
		Where("user_id = ?", userID).
		Find(&memberIDs).Error
	if err != nil {
		return nil, 0, err
	}

	if len(memberIDs) == 0 {
		return []*models.Conversation{}, 0, nil
	}

	// เริ่มสร้าง query
	baseQuery := r.db.Model(&models.Conversation{}).
		Where("conversations.id IN (?) AND conversations.is_active = ?", memberIDs, true) // เพิ่ม conversations. นำหน้า id

	// เพิ่มเงื่อนไขเพิ่มเติมตามพารามิเตอร์
	if convType != "" {
		baseQuery = baseQuery.Where("conversations.type = ?", convType) // เพิ่ม conversations. นำหน้า type
	}

	// สร้าง subquery สำหรับการดึงการสนทนาที่ปักหมุด
	if pinned {
		pinnedIDs := []uuid.UUID{}
		subQuery := r.db.Model(&models.ConversationMember{}).
			Select("conversation_id").
			Where("user_id = ? AND is_pinned = ?", userID, true)

		if err := subQuery.Find(&pinnedIDs).Error; err != nil {
			return nil, 0, err
		}

		if len(pinnedIDs) > 0 {
			baseQuery = baseQuery.Where("conversations.id IN ?", pinnedIDs) // เพิ่ม conversations. นำหน้า id
		} else {
			// ถ้าไม่มีการสนทนาที่ปักหมุด และต้องการเฉพาะที่ปักหมุด
			return []*models.Conversation{}, 0, nil
		}
	}

	// นับจำนวนทั้งหมด
	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// เตรียม query สำหรับดึงข้อมูลพร้อมจัดเรียง
	// 1. ดึงการสนทนาที่ปักหมุดก่อน (ถ้าไม่ได้กรองเฉพาะที่ปักหมุด)
	// 2. จัดเรียงตาม last_message_at หรือ updated_at (จากใหม่ไปเก่า)
	queryWithOrder := baseQuery.Session(&gorm.Session{})

	// เพิ่มการจัดเรียงตาม is_pinned (ถ้าไม่ได้กรองเฉพาะที่ปักหมุด)
	if !pinned {
		queryWithOrder = queryWithOrder.Joins(`
            LEFT JOIN conversation_members cm ON conversations.id = cm.conversation_id AND cm.user_id = ?
        `, userID).
			Order("cm.is_pinned DESC")
	}

	// จัดเรียงตาม last_message_at หรือ updated_at (จากใหม่ไปเก่า)
	queryWithOrder = queryWithOrder.
		Order("COALESCE(conversations.last_message_at, conversations.updated_at) DESC"). // เพิ่ม conversations. นำหน้า
		Limit(limit).
		Offset(offset)

	// ดึงข้อมูล
	var conversations []*models.Conversation
	err = queryWithOrder.Find(&conversations).Error
	if err != nil {
		return nil, 0, err
	}

	return conversations, int(total), nil
}

// infrastructure/persistence/postgres/conversation_repository.go
// เพิ่มเมธอดเหล่านี้ใน conversationRepository struct ที่มีอยู่แล้ว

// ========================================
// 🏢 BUSINESS CONVERSATION REPOSITORY IMPLEMENTATIONS
// ========================================

// GetBusinessConversations ดึงการสนทนาทั้งหมดของธุรกิจ
func (r *conversationRepository) GetBusinessConversations(businessID uuid.UUID, limit, offset int) ([]*models.Conversation, int, error) {
	// นับจำนวนทั้งหมด
	var count int64
	if err := r.db.Model(&models.Conversation{}).
		Where("business_id = ? AND is_active = ?", businessID, true).
		Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลการสนทนาของธุรกิจ
	var conversations []*models.Conversation
	err := r.db.Where("business_id = ? AND is_active = ?", businessID, true).
		Order("COALESCE(last_message_at, updated_at) DESC").
		Limit(limit).
		Offset(offset).
		Find(&conversations).Error

	if err != nil {
		return nil, 0, err
	}

	return conversations, int(count), nil
}

// GetBusinessConversationsBeforeTime ดึงการสนทนาธุรกิจที่เก่ากว่าเวลาที่ระบุ
func (r *conversationRepository) GetBusinessConversationsBeforeTime(businessID uuid.UUID, beforeTime time.Time, limit int) ([]*models.Conversation, int, error) {
	// เริ่มสร้าง query
	baseQuery := r.db.Model(&models.Conversation{}).
		Where("business_id = ? AND is_active = ?", businessID, true)

	// เพิ่มเงื่อนไขเวลา
	baseQuery = baseQuery.Where("COALESCE(last_message_at, updated_at) < ?", beforeTime)

	// นับจำนวนทั้งหมด
	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลตามเงื่อนไข
	var conversations []*models.Conversation
	err := baseQuery.
		Order("COALESCE(last_message_at, updated_at) DESC").
		Limit(limit).
		Find(&conversations).Error
	if err != nil {
		return nil, 0, err
	}

	return conversations, int(total), nil
}

// GetBusinessConversationsAfterTime ดึงการสนทนาธุรกิจที่ใหม่กว่าเวลาที่ระบุ
func (r *conversationRepository) GetBusinessConversationsAfterTime(businessID uuid.UUID, afterTime time.Time, limit int) ([]*models.Conversation, int, error) {
	// เริ่มสร้าง query
	baseQuery := r.db.Model(&models.Conversation{}).
		Where("business_id = ? AND is_active = ?", businessID, true)

	// เพิ่มเงื่อนไขเวลา
	baseQuery = baseQuery.Where("COALESCE(last_message_at, updated_at) > ?", afterTime)

	// นับจำนวนทั้งหมด
	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลตามเงื่อนไข
	var conversations []*models.Conversation
	err := baseQuery.
		Order("COALESCE(last_message_at, updated_at) ASC").
		Limit(limit).
		Find(&conversations).Error
	if err != nil {
		return nil, 0, err
	}

	// กลับลำดับให้เป็น DESC (จากใหม่ไปเก่า)
	for i := 0; i < len(conversations)/2; i++ {
		j := len(conversations) - i - 1
		conversations[i], conversations[j] = conversations[j], conversations[i]
	}

	return conversations, int(total), nil
}

// GetBusinessConversationsBeforeID ดึงการสนทนาธุรกิจที่เก่ากว่า ID ที่ระบุ
func (r *conversationRepository) GetBusinessConversationsBeforeID(businessID, beforeID uuid.UUID, limit int) ([]*models.Conversation, int, error) {
	// ดึงการสนทนาเป้าหมายเพื่อดูเวลาของมัน
	var targetConversation models.Conversation
	err := r.db.First(&targetConversation, "id = ?", beforeID).Error
	if err != nil {
		return nil, 0, fmt.Errorf("error fetching target conversation: %w", err)
	}

	// เริ่มสร้าง query
	baseQuery := r.db.Model(&models.Conversation{}).
		Where("business_id = ? AND is_active = ?", businessID, true)

	// ใช้ LastMessageAt หรือ UpdatedAt เพื่อเปรียบเทียบ
	var timeCondition string
	var args []interface{}

	if targetConversation.LastMessageAt != nil {
		timeCondition = "(COALESCE(last_message_at, updated_at) < ? OR (COALESCE(last_message_at, updated_at) = ? AND id < ?))"
		args = []interface{}{targetConversation.LastMessageAt, targetConversation.LastMessageAt, beforeID}
	} else {
		timeCondition = "(COALESCE(last_message_at, updated_at) < ? OR (COALESCE(last_message_at, updated_at) = ? AND id < ?))"
		args = []interface{}{targetConversation.UpdatedAt, targetConversation.UpdatedAt, beforeID}
	}

	baseQuery = baseQuery.Where(timeCondition, args...)

	// นับจำนวนทั้งหมด
	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลตามเงื่อนไข
	var conversations []*models.Conversation
	err = baseQuery.
		Order("COALESCE(last_message_at, updated_at) DESC").
		Limit(limit).
		Find(&conversations).Error
	if err != nil {
		return nil, 0, err
	}

	return conversations, int(total), nil
}

// GetBusinessConversationsAfterID ดึงการสนทนาธุรกิจที่ใหม่กว่า ID ที่ระบุ
func (r *conversationRepository) GetBusinessConversationsAfterID(businessID, afterID uuid.UUID, limit int) ([]*models.Conversation, int, error) {
	// ดึงการสนทนาเป้าหมายเพื่อดูเวลาของมัน
	var targetConversation models.Conversation
	err := r.db.First(&targetConversation, "id = ?", afterID).Error
	if err != nil {
		return nil, 0, fmt.Errorf("error fetching target conversation: %w", err)
	}

	// เริ่มสร้าง query
	baseQuery := r.db.Model(&models.Conversation{}).
		Where("business_id = ? AND is_active = ?", businessID, true)

	// ใช้ LastMessageAt หรือ UpdatedAt เพื่อเปรียบเทียบ
	var timeCondition string
	var args []interface{}

	if targetConversation.LastMessageAt != nil {
		timeCondition = "(COALESCE(last_message_at, updated_at) > ? OR (COALESCE(last_message_at, updated_at) = ? AND id > ?))"
		args = []interface{}{targetConversation.LastMessageAt, targetConversation.LastMessageAt, afterID}
	} else {
		timeCondition = "(COALESCE(last_message_at, updated_at) > ? OR (COALESCE(last_message_at, updated_at) = ? AND id > ?))"
		args = []interface{}{targetConversation.UpdatedAt, targetConversation.UpdatedAt, afterID}
	}

	baseQuery = baseQuery.Where(timeCondition, args...)

	// นับจำนวนทั้งหมด
	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลตามเงื่อนไข
	var conversations []*models.Conversation
	err = baseQuery.
		Order("COALESCE(last_message_at, updated_at) ASC").
		Limit(limit).
		Find(&conversations).Error
	if err != nil {
		return nil, 0, err
	}

	// กลับลำดับให้เป็น DESC (จากใหม่ไปเก่า)
	for i := 0; i < len(conversations)/2; i++ {
		j := len(conversations) - i - 1
		conversations[i], conversations[j] = conversations[j], conversations[i]
	}

	return conversations, int(total), nil
}

// CheckConversationBelongsToBusiness ตรวจสอบว่าการสนทนาเป็นของธุรกิจ
func (r *conversationRepository) CheckConversationBelongsToBusiness(conversationID, businessID uuid.UUID) (bool, error) {
	var conversation models.Conversation
	err := r.db.Select("business_id").First(&conversation, "id = ?", conversationID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	// ตรวจสอบว่า business_id ตรงกันหรือไม่
	return conversation.BusinessID != nil && *conversation.BusinessID == businessID, nil
}

// ========================================
// 🔍 ADDITIONAL HELPER METHODS สำหรับ Business Context
// ========================================

// GetBusinessConversationCount นับจำนวนการสนทนาทั้งหมดของธุรกิจ
func (r *conversationRepository) GetBusinessConversationCount(businessID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Conversation{}).
		Where("business_id = ? AND is_active = ?", businessID, true).
		Count(&count).Error
	return count, err
}

// GetBusinessActiveConversations ดึงการสนทนาที่มีข้อความล่าสุดในช่วงเวลาที่กำหนด
func (r *conversationRepository) GetBusinessActiveConversations(businessID uuid.UUID, since time.Time, limit int) ([]*models.Conversation, error) {
	var conversations []*models.Conversation
	err := r.db.Where("business_id = ? AND is_active = ? AND last_message_at >= ?", businessID, true, since).
		Order("last_message_at DESC").
		Limit(limit).
		Find(&conversations).Error
	return conversations, err
}

// GetBusinessConversationWithCustomer ดึงการสนทนาระหว่างธุรกิจกับลูกค้าเฉพาะ
func (r *conversationRepository) GetBusinessConversationWithCustomer(businessID, customerID uuid.UUID) (*models.Conversation, error) {
	// หา conversation ที่เป็นของธุรกิจ
	var businessConversations []models.Conversation
	err := r.db.Where("business_id = ? AND type = ? AND is_active = ?", businessID, "business", true).
		Find(&businessConversations).Error
	if err != nil {
		return nil, err
	}

	// ตรวจสอบแต่ละการสนทนาว่ามีลูกค้าคนนี้เป็นสมาชิกหรือไม่
	for _, conv := range businessConversations {
		var memberCount int64
		err := r.db.Model(&models.ConversationMember{}).
			Where("conversation_id = ? AND user_id = ?", conv.ID, customerID).
			Count(&memberCount).Error
		if err == nil && memberCount > 0 {
			return &conv, nil
		}
	}

	return nil, nil // ไม่พบการสนทนา
}

// FindBusinessUserConversation หาการสนทนาระหว่างธุรกิจและผู้ใช้
func (r *conversationRepository) FindBusinessUserConversation(businessID, userID uuid.UUID) (uuid.UUID, error) {
	// ใช้ model ของ Conversation โดยตรง
	var conversation models.Conversation

	err := r.db.Joins("JOIN conversation_members m ON conversations.id = m.conversation_id").
		Where("conversations.business_id = ? AND conversations.type = 'business' AND m.user_id = ? AND conversations.is_active = true",
			businessID, userID).
		First(&conversation).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, nil // ไม่พบการสนทนา
		}
		return uuid.Nil, err
	}

	return conversation.ID, nil
}
