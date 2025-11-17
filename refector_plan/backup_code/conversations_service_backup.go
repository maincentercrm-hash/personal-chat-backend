// application/serviceimpl/conversation_service.go
package serviceimpl

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/dto"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/domain/types"
)

type conversationService struct {
	conversationRepo    repository.ConversationRepository
	userRepo            repository.UserRepository
	businessRepo        repository.BusinessAccountRepository
	messageRepo         repository.MessageRepository
	businessAdminRepo   repository.BusinessAdminRepository
	customerProfileRepo repository.CustomerProfileRepository
}

// NewConversationService สร้าง service ใหม่
func NewConversationService(
	conversationRepo repository.ConversationRepository,
	userRepo repository.UserRepository,
	businessRepo repository.BusinessAccountRepository,
	messageRepo repository.MessageRepository,
	businessAdminRepo repository.BusinessAdminRepository,
	customerProfileRepo repository.CustomerProfileRepository,

) service.ConversationService {
	return &conversationService{
		conversationRepo:    conversationRepo,
		userRepo:            userRepo,
		businessRepo:        businessRepo,
		messageRepo:         messageRepo,
		businessAdminRepo:   businessAdminRepo,
		customerProfileRepo: customerProfileRepo,
	}
}

// CreateDirectConversation สร้างการสนทนาแบบส่วนตัวระหว่างผู้ใช้สองคน
func (s *conversationService) CreateDirectConversation(userID, friendID uuid.UUID) (*dto.ConversationDTO, error) {

	// 1. ตรวจสอบว่าผู้ใช้ที่จะสนทนาด้วยมีอยู่จริงไหม
	friend, err := s.userRepo.FindByID(friendID)
	if err != nil || friend == nil {
		return nil, errors.New("friend not found")
	}

	// 2. ตรวจสอบความเป็นเพื่อน (เพิ่มความเข้มงวด)
	isFriend, err := s.checkFriendship(userID, friendID)
	if err != nil {
		return nil, err
	}
	if !isFriend {
		return nil, errors.New("you must be friends to start a chat")
	}

	// 3. ตรวจสอบว่ามีการสนทนาอยู่แล้วหรือไม่
	existingConv, err := s.conversationRepo.FindDirectConversation(userID, friendID)
	if err == nil && existingConv != nil {
		// ถ้ามีการสนทนาอยู่แล้ว ดึงข้อมูลการสนทนาและส่งกลับ
		return s.convertToConversationDTO(existingConv, userID)
	}

	// 4. สร้างการสนทนาใหม่
	now := time.Now()
	conversation := &models.Conversation{
		ID:        uuid.New(),
		Type:      "direct",
		CreatedAt: now,
		UpdatedAt: now,
		CreatorID: &userID,
		IsActive:  true,
	}

	if err := s.conversationRepo.Create(conversation); err != nil {
		return nil, err
	}

	// เพิ่มผู้สร้างเป็นสมาชิก
	member1 := &models.ConversationMember{
		ID:             uuid.New(),
		ConversationID: conversation.ID,
		UserID:         userID,
		IsAdmin:        true,
		JoinedAt:       now,
	}
	if err := s.conversationRepo.AddMember(member1); err != nil {
		return nil, err
	}

	// เพิ่มผู้ใช้อีกคนเป็นสมาชิก
	member2 := &models.ConversationMember{
		ID:             uuid.New(),
		ConversationID: conversation.ID,
		UserID:         friendID,
		IsAdmin:        false,
		JoinedAt:       now,
	}
	if err := s.conversationRepo.AddMember(member2); err != nil {
		// ข้อผิดพลาดไม่ร้ายแรง แต่เราควรบันทึกลงในล็อก
	}

	// สร้างข้อความระบบแจ้งการสร้างการสนทนา
	welcomeMessageText := "Conversation created."
	err = s.createSystemMessage(conversation.ID, welcomeMessageText)
	if err != nil {
		// ไม่คืนค่าข้อผิดพลาด แต่ควรบันทึกลงในล็อก
	}

	// ดึงข้อมูลการสนทนาที่สร้างเสร็จแล้ว
	createdConv, err := s.conversationRepo.GetByID(conversation.ID)
	if err != nil {
		return nil, err
	}

	// แปลงเป็น DTO สำหรับผู้สร้าง
	creatorDTO, err := s.convertToConversationDTO(createdConv, userID)
	if err != nil {
		return nil, err
	}

	return creatorDTO, nil
}

// GetUserConversations ดึงรายการการสนทนาทั้งหมดของผู้ใช้ พร้อมตัวกรอง
func (s *conversationService) GetUserConversations(userID uuid.UUID, limit, offset int,
	convType string, pinned bool) ([]*dto.ConversationDTO, int, error) {

	// เรียกใช้ repository
	conversations, total, err := s.conversationRepo.GetUserConversationsWithFilter(
		userID, limit, offset, convType, pinned)
	if err != nil {
		return nil, 0, err
	}

	// แปลงเป็น DTOs และกรอง business conversations ที่ user เป็นแอดมิน
	dtos := make([]*dto.ConversationDTO, 0, len(conversations))
	filteredCount := 0 // นับจำนวนที่ถูกกรอง

	for _, conversation := range conversations {
		dto, err := s.convertToConversationDTO(conversation, userID)
		if err != nil {
			// ถ้าเป็น error จากการกรอง business conversation ให้ข้าม
			if err.Error() == "business conversation filtered for admin user" {
				filteredCount++
				continue
			}
			// ข้ามการสนทนาที่มีปัญหาอื่นๆ
			continue
		}
		dtos = append(dtos, dto)
	}

	// ปรับ total ให้ตรงกับจำนวนที่แสดงจริง
	adjustedTotal := total - filteredCount

	return dtos, adjustedTotal, nil
}

// ฟังก์ชันช่วยเหลือ
// application/serviceimpl/conversation_service.go
func (s *conversationService) convertToConversationDTO(conversation *models.Conversation, userID uuid.UUID) (*dto.ConversationDTO, error) {
	if conversation == nil {
		return nil, errors.New("conversation is nil")
	}

	// ⚠️ สำคัญ: กรอง business conversation ที่ user เป็นแอดมิน
	if conversation.Type == "business" && conversation.BusinessID != nil {
		// ตรวจสอบว่า user เป็นแอดมินของธุรกิจนี้หรือไม่
		isBusinessAdmin, err := s.businessAdminRepo.CheckAdminPermission(userID, *conversation.BusinessID, []string{})
		if err == nil && isBusinessAdmin {
			// ถ้าเป็นแอดมินของธุรกิจ = ไม่แสดงใน personal conversations
			return nil, errors.New("business conversation filtered for admin user")
		}
	}

	convDTO := &dto.ConversationDTO{
		ID:              conversation.ID,
		Type:            conversation.Type,
		Title:           conversation.Title,
		IconURL:         conversation.IconURL,
		CreatedAt:       conversation.CreatedAt,
		UpdatedAt:       conversation.UpdatedAt,
		LastMessageText: conversation.LastMessageText,
		LastMessageAt:   conversation.LastMessageAt,
		CreatorID:       conversation.CreatorID,
		BusinessID:      conversation.BusinessID,
		IsActive:        conversation.IsActive,
		Metadata:        conversation.Metadata,
	}

	// ดึงข้อมูลเพิ่มเติมตามประเภทการสนทนา
	if conversation.Type == "direct" {
		// ... โค้ดเดิมสำหรับ direct conversation
		members, err := s.conversationRepo.GetMembers(conversation.ID)
		if err == nil && len(members) > 0 {
			var otherMember *models.ConversationMember
			for _, member := range members {
				if member.UserID != userID {
					otherMember = member
					break
				}
			}

			if otherMember != nil {
				friend, err := s.userRepo.FindByID(otherMember.UserID)
				if err == nil && friend != nil {
					if convDTO.Title == "" {
						if friend.DisplayName != "" {
							convDTO.Title = friend.DisplayName
						} else {
							convDTO.Title = friend.Username
						}
					}

					if convDTO.IconURL == "" {
						convDTO.IconURL = friend.ProfileImageURL
					}

					contactInfo := types.JSONB{
						"user_id":           friend.ID.String(),
						"username":          friend.Username,
						"display_name":      friend.DisplayName,
						"profile_image_url": friend.ProfileImageURL,
					}
					convDTO.ContactInfo = contactInfo
				}
			}
		}
	} else if conversation.Type == "business" && conversation.BusinessID != nil {
		// เฉพาะกรณีที่ไม่ใช่แอดมิน (ผ่านการกรองข้างต้นแล้ว)
		business, err := s.businessRepo.GetByID(*conversation.BusinessID)
		if err == nil && business != nil {
			if convDTO.Title == "" {
				convDTO.Title = business.Name
			}

			if convDTO.IconURL == "" {
				convDTO.IconURL = business.ProfileImageURL
			}

			businessInfo := types.JSONB{
				"id":                business.ID.String(),
				"name":              business.Name,
				"profile_image_url": business.ProfileImageURL,
			}
			convDTO.BusinessInfo = businessInfo
		}
	}

	// ตรวจสอบสถานะ pin/mute
	member, err := s.conversationRepo.GetMember(conversation.ID, userID)
	if err == nil && member != nil {
		convDTO.IsPinned = member.IsPinned
		convDTO.IsMuted = member.IsMuted

		// คำนวณ unread_count
		var unreadCount int
		if member.LastReadAt != nil {
			messages, err := s.messageRepo.GetMessagesAfterTime(
				conversation.ID, *member.LastReadAt, userID)
			if err == nil {
				unreadCount = len(messages)
			}
		} else {
			messages, err := s.messageRepo.GetAllUnreadMessages(
				conversation.ID, userID)
			if err == nil {
				unreadCount = len(messages)
			}
		}

		convDTO.UnreadCount = unreadCount
	} else {
		convDTO.IsPinned = false
		convDTO.IsMuted = false
		convDTO.UnreadCount = 0
	}

	// จำนวนสมาชิก
	members, err := s.conversationRepo.GetMembers(conversation.ID)
	if err == nil {
		convDTO.MemberCount = len(members)
	} else {
		convDTO.MemberCount = 0
	}

	return convDTO, nil
}

func (s *conversationService) checkFriendship(userID, friendID uuid.UUID) (bool, error) {
	// ต้องมีฟังก์ชันในระบบของคุณสำหรับการตรวจสอบความเป็นเพื่อน
	// ในตัวอย่างนี้จะให้ค่าจริงเสมอ
	return true, nil
}

func (s *conversationService) createSystemMessage(conversationID uuid.UUID, content string) error {
	// ควรเรียกใช้ MessageRepository เพื่อสร้างข้อความระบบ
	// ในตัวอย่างนี้จะไม่ทำการสร้างข้อความจริง
	return nil
}

func (s *conversationService) getUserName(userID uuid.UUID) (string, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return "", err
	}

	if user.DisplayName != "" {
		return user.DisplayName, nil
	}
	return user.Username, nil
}

func (s *conversationService) userExists(userID uuid.UUID) (bool, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return false, err
	}
	return user != nil, nil
}

// application/serviceimpl/conversations_service.go

// CreateBusinessConversation สร้างการสนทนากับธุรกิจ

func (s *conversationService) CreateBusinessConversation(userID, businessID uuid.UUID) (*dto.ConversationDTO, error) {
	// ตรวจสอบว่าธุรกิจมีอยู่จริง
	business, err := s.businessRepo.GetByID(businessID)
	if err != nil {
		return nil, err
	}

	if business == nil {
		return nil, errors.New("business not found")
	}

	// 🆕 สร้าง Customer Profile อัตโนมัติ (ถ้ายังไม่มี)
	err = s.ensureCustomerProfile(businessID, userID)
	if err != nil {
		// Log error แต่ไม่ให้ fail การสร้าง conversation
		// เพราะ customer profile ไม่ใช่ส่วนที่จำเป็นต่อการทำงานของ conversation
		fmt.Printf("Warning: Failed to create customer profile for user %s in business %s: %v\n",
			userID.String(), businessID.String(), err)
	}

	// สร้างการสนทนาใหม่
	now := time.Now()
	conversation := &models.Conversation{
		ID:         uuid.New(),
		Type:       "business",
		CreatedAt:  now,
		UpdatedAt:  now,
		CreatorID:  &userID,
		BusinessID: &businessID,
		IsActive:   true,
	}

	if err := s.conversationRepo.Create(conversation); err != nil {
		return nil, err
	}

	// เพิ่มผู้สร้างเป็นสมาชิก
	creator := &models.ConversationMember{
		ID:             uuid.New(),
		ConversationID: conversation.ID,
		UserID:         userID,
		IsAdmin:        false,
		JoinedAt:       now,
	}
	if err := s.conversationRepo.AddMember(creator); err != nil {
		return nil, err
	}

	// รวบรวม member IDs ทั้งหมดสำหรับ WebSocket notification
	allMemberIDs := []uuid.UUID{userID}

	// ดึงข้อมูลเจ้าของธุรกิจและเพิ่มเป็นสมาชิก
	if business.OwnerID != nil {
		ownerID := *business.OwnerID
		if ownerID != userID {
			owner := &models.ConversationMember{
				ID:             uuid.New(),
				ConversationID: conversation.ID,
				UserID:         ownerID,
				IsAdmin:        true,
				JoinedAt:       now,
			}
			if err := s.conversationRepo.AddMember(owner); err == nil {
				allMemberIDs = append(allMemberIDs, ownerID)
			}
		}
	}

	// สร้างข้อความระบบ
	/*
		welcomeMessageText := "Welcome to our business chat! How can we help you today?"
		if business.WelcomeMessage != "" {
			welcomeMessageText = business.WelcomeMessage
		}
		s.createSystemMessage(conversation.ID, welcomeMessageText)
	*/

	// ดึงข้อมูลการสนทนาที่สร้างเสร็จแล้ว
	createdConv, err := s.conversationRepo.GetByID(conversation.ID)
	if err != nil {
		return nil, err
	}

	// แปลงเป็น DTO
	convDTO, err := s.convertToConversationDTO(createdConv, userID)
	if err != nil {
		return nil, err
	}

	// เพิ่มข้อมูลเพิ่มเติม
	convDTO.MemberCount = len(allMemberIDs)
	convDTO.IsPinned = false
	convDTO.IsMuted = false
	convDTO.UnreadCount = 0

	return convDTO, nil
}

// แก้ไขเมธอด ensureCustomerProfile ในไฟล์ application/serviceimpl/conversation_service.go

// 🆕 Helper method สำหรับสร้าง Customer Profile
func (s *conversationService) ensureCustomerProfile(businessID, userID uuid.UUID) error {
	// ตรวจสอบว่ามี customer profile อยู่แล้วหรือไม่
	_, err := s.customerProfileRepo.GetByBusinessAndUser(businessID, userID)
	if err == nil {
		// มีอยู่แล้ว ไม่ต้องทำอะไร
		return nil
	}

	// ดึงข้อมูล user เพื่อใช้สร้าง profile
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}

	// สร้าง customer profile ใหม่
	now := time.Now()
	profile := &models.CustomerProfile{
		ID:         uuid.New(),
		BusinessID: businessID,
		UserID:     userID,
		Nickname:   "", // เริ่มต้นว่าง admin จะตั้งชื่อเล่นทีหลัง
		//Notes:        "Auto-created when customer started conversation",
		Notes:        "",
		CustomerType: "New",    // ประเภทลูกค้าเริ่มต้น (ตาม model: VIP, Regular, New, etc.)
		Status:       "active", // สถานะเริ่มต้น
		Metadata: types.JSONB{
			"source":       "conversation",
			"display_name": user.DisplayName,
			"username":     user.Username,
			"auto_created": true,
			"created_via":  "business_conversation",
		},
		CreatedAt:     now,
		UpdatedAt:     now,
		LastContactAt: &now, // pointer to time.Time บันทึกเวลาติดต่อครั้งแรก
		CreatedByID:   nil,  // system created, no admin
	}

	// บันทึก customer profile
	err = s.customerProfileRepo.Create(profile)
	if err != nil {
		return fmt.Errorf("failed to create customer profile: %w", err)
	}

	return nil
}

// 🔧 ต้องเพิ่ม repositories ใน conversationService struct
// type conversationService struct {
// 	conversationRepo     repository.ConversationRepository
// 	businessRepo         repository.BusinessRepository
// 	userRepo             repository.UserRepository
// 	customerProfileRepo  repository.CustomerProfileRepository  // 🆕 เพิ่มบรรทัดนี้
// 	// ... repositories อื่นๆ
// }

// application/serviceimpl/conversations_service.go
// เพิ่มเมธอด CreateGroupConversation

// CreateGroupConversation สร้างการสนทนาแบบกลุ่ม
func (s *conversationService) CreateGroupConversation(userID uuid.UUID, title, iconURL string, memberIDs []uuid.UUID) (*dto.ConversationDTO, error) {
	// 1. ตรวจสอบข้อมูลที่จำเป็น
	if title == "" {
		return nil, errors.New("group conversation requires a title")
	}

	// 2. ตรวจสอบว่ามีสมาชิกอย่างน้อย 1 คน (นอกเหนือจากผู้สร้าง)
	if len(memberIDs) == 0 {
		return nil, errors.New("at least one member is required for group conversation")
	}

	// 3. ตรวจสอบว่าสมาชิกทุกคนมีอยู่จริงและเป็นเพื่อนกับผู้สร้าง
	validMemberIDs := []uuid.UUID{}
	for _, memberID := range memberIDs {
		// ข้ามถ้าเป็น ID ของผู้สร้าง
		if memberID == userID {
			continue
		}

		// ตรวจสอบว่าผู้ใช้มีอยู่จริง
		user, err := s.userRepo.FindByID(memberID)
		if err != nil || user == nil {
			// ข้ามผู้ใช้ที่ไม่มีอยู่จริง
			continue
		}

		// ตรวจสอบความเป็นเพื่อน (เพิ่มความเข้มงวด)
		isFriend, err := s.checkFriendship(userID, memberID)
		if err != nil || !isFriend {
			// ข้ามผู้ใช้ที่ไม่ใช่เพื่อน
			continue
		}

		validMemberIDs = append(validMemberIDs, memberID)
	}

	// 4. ตรวจสอบว่ามีสมาชิกที่ถูกต้องอย่างน้อย 1 คน
	if len(validMemberIDs) == 0 {
		return nil, errors.New("no valid members found for group conversation")
	}

	// 5. สร้างการสนทนาใหม่
	now := time.Now()
	conversation := &models.Conversation{
		ID:        uuid.New(),
		Type:      "group",
		Title:     title,
		IconURL:   iconURL,
		CreatedAt: now,
		UpdatedAt: now,
		CreatorID: &userID,
		IsActive:  true,
	}

	if err := s.conversationRepo.Create(conversation); err != nil {
		return nil, err
	}

	// 6. เพิ่มผู้สร้างเป็นสมาชิกและแอดมิน
	creator := &models.ConversationMember{
		ID:             uuid.New(),
		ConversationID: conversation.ID,
		UserID:         userID,
		IsAdmin:        true,
		JoinedAt:       now,
	}
	if err := s.conversationRepo.AddMember(creator); err != nil {
		return nil, err
	}

	// 7. เพิ่มสมาชิกอื่นๆ (ที่ผ่านการตรวจสอบแล้ว)
	allMemberIDs := []uuid.UUID{userID} // เริ่มด้วยผู้สร้าง
	for _, memberID := range validMemberIDs {
		// เพิ่มเป็นสมาชิก
		member := &models.ConversationMember{
			ID:             uuid.New(),
			ConversationID: conversation.ID,
			UserID:         memberID,
			IsAdmin:        false,
			JoinedAt:       now,
		}
		if err := s.conversationRepo.AddMember(member); err != nil {
			// ไม่คืนค่าข้อผิดพลาด แต่ควรบันทึกลงในล็อก
			continue
		}
		allMemberIDs = append(allMemberIDs, memberID)
	}

	// 8. สร้างข้อความระบบแจ้งการสร้างกลุ่ม
	welcomeMessageText := "Group created."
	creatorName, err := s.getUserName(userID)
	if err == nil && creatorName != "" {
		welcomeMessageText = creatorName + " created the group."
	}
	s.createSystemMessage(conversation.ID, welcomeMessageText)

	// 9. ดึงข้อมูลการสนทนาที่สร้างเสร็จแล้ว
	createdConv, err := s.conversationRepo.GetByID(conversation.ID)
	if err != nil {
		return nil, err
	}

	// 10. แปลงเป็น DTO
	convDTO, err := s.convertToConversationDTO(createdConv, userID)
	if err != nil {
		return nil, err
	}

	// 11. เพิ่มข้อมูลเพิ่มเติม
	convDTO.MemberCount = len(allMemberIDs)
	convDTO.IsPinned = false
	convDTO.IsMuted = false
	convDTO.UnreadCount = 0

	return convDTO, nil
}

// GetConversationMessages ดึงข้อความทั้งหมดในการสนทนา
func (s *conversationService) GetConversationMessages(conversationID, userID uuid.UUID, limit, offset int) ([]*dto.MessageDTO, int64, error) {
	// ตรวจสอบว่าผู้ใช้เป็นสมาชิกของการสนทนานี้
	isMember, err := s.conversationRepo.IsMember(conversationID, userID)
	if err != nil {
		return nil, 0, err
	}

	if !isMember {
		return nil, 0, errors.New("you are not a member of this conversation")
	}

	// ดึงข้อความทั้งหมดในการสนทนา
	messages, total, err := s.messageRepo.GetMessagesByConversationID(conversationID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// แปลงข้อความเป็น DTOs
	messageDTOs := make([]*dto.MessageDTO, 0, len(messages))
	for _, msg := range messages {
		messageDTO, err := s.ConvertToMessageDTO(msg, userID)
		if err != nil {
			// ข้ามข้อความที่มีปัญหา
			continue
		}
		messageDTOs = append(messageDTOs, messageDTO)
	}

	return messageDTOs, total, nil
}

// SetPinStatus กำหนดสถานะการปักหมุดของการสนทนา
func (s *conversationService) SetPinStatus(conversationID, userID uuid.UUID, isPinned bool) error {
	// ตรวจสอบว่าผู้ใช้เป็นสมาชิกของการสนทนานี้
	isMember, err := s.conversationRepo.IsMember(conversationID, userID)
	if err != nil {
		return err
	}

	if !isMember {
		return errors.New("you are not a member of this conversation")
	}

	// อัพเดตสถานะปักหมุด
	return s.conversationRepo.SetPinStatus(conversationID, userID, isPinned)
}

// SetMuteStatus กำหนดสถานะการปิดเสียงของการสนทนา
func (s *conversationService) SetMuteStatus(conversationID, userID uuid.UUID, isMuted bool) error {
	// ตรวจสอบว่าผู้ใช้เป็นสมาชิกของการสนทนานี้
	isMember, err := s.conversationRepo.IsMember(conversationID, userID)
	if err != nil {
		return err
	}

	if !isMember {
		return errors.New("you are not a member of this conversation")
	}

	// อัพเดตสถานะปิดเสียง
	return s.conversationRepo.SetMuteStatus(conversationID, userID, isMuted)
}

// CheckMembership ตรวจสอบว่าผู้ใช้เป็นสมาชิกของการสนทนาหรือไม่
func (s *conversationService) CheckMembership(userID, conversationID uuid.UUID) (bool, error) {
	return s.conversationRepo.IsMember(conversationID, userID)
}

// ConvertToMessageDTO แปลง Message model เป็น MessageDTO
func (s *conversationService) ConvertToMessageDTO(msg *models.Message, userID uuid.UUID) (*dto.MessageDTO, error) {
	if msg == nil {
		return nil, errors.New("message is nil")
	}

	messageDTO := &dto.MessageDTO{
		ID:                msg.ID,
		ConversationID:    msg.ConversationID,
		SenderID:          msg.SenderID,
		SenderType:        msg.SenderType,
		MessageType:       msg.MessageType,
		Content:           msg.Content,
		MediaURL:          msg.MediaURL,
		MediaThumbnailURL: msg.MediaThumbnailURL,
		Metadata:          msg.Metadata,
		CreatedAt:         msg.CreatedAt,
		UpdatedAt:         msg.UpdatedAt,
		IsDeleted:         msg.IsDeleted,
		IsEdited:          msg.IsEdited,
		EditCount:         msg.EditCount,
		ReplyToID:         msg.ReplyToID,
		BusinessID:        msg.BusinessID,
		ReadCount:         0,     // ค่าเริ่มต้น จะอัปเดตทีหลัง
		IsRead:            false, // ค่าเริ่มต้น จะอัปเดตทีหลัง
	}

	// 1. เพิ่มข้อมูลผู้ส่ง
	s.addSenderInfoToDTO(messageDTO)

	// 2. เพิ่มข้อมูลสถานะการอ่าน
	s.addReadStatusToDTO(messageDTO, userID)

	// 3. เพิ่มข้อมูลข้อความที่ตอบกลับ (ถ้ามี)
	if msg.ReplyToID != nil {
		s.addReplyToInfoToDTO(messageDTO)
	}

	return messageDTO, nil
}

// addSenderInfoToDTO เพิ่มข้อมูลผู้ส่งใน DTO
func (s *conversationService) addSenderInfoToDTO(msgDTO *dto.MessageDTO) {
	if msgDTO.SenderID == nil {
		return
	}

	if msgDTO.SenderType == "business" && msgDTO.BusinessID != nil {
		// ดึงข้อมูลธุรกิจ
		business, err := s.businessRepo.GetByID(*msgDTO.BusinessID)
		if err == nil && business != nil {
			msgDTO.SenderName = business.Name
			msgDTO.SenderAvatar = business.ProfileImageURL
		}
	} else {
		// ดึงข้อมูลผู้ใช้
		user, err := s.userRepo.FindByID(*msgDTO.SenderID)
		if err == nil && user != nil {
			if user.DisplayName != "" {
				msgDTO.SenderName = user.DisplayName
			} else {
				msgDTO.SenderName = user.Username
			}
			msgDTO.SenderAvatar = user.ProfileImageURL
		}
	}
}

// addReadStatusToDTO เพิ่มข้อมูลสถานะการอ่านใน DTO
func (s *conversationService) addReadStatusToDTO(msgDTO *dto.MessageDTO, userID uuid.UUID) {
	// ดึงข้อมูลการอ่านทั้งหมดของข้อความนี้
	reads, err := s.messageRepo.GetReads(msgDTO.ID)
	if err != nil {
		return
	}

	// คำนวณ ReadCount
	msgDTO.ReadCount = len(reads)

	// ตรวจสอบว่าผู้ใช้อ่านข้อความแล้วหรือยัง
	for _, read := range reads {
		if read.UserID == userID {
			msgDTO.IsRead = true
			break
		}
	}
}

// addReplyToInfoToDTO เพิ่มข้อมูลข้อความที่ตอบกลับใน DTO
func (s *conversationService) addReplyToInfoToDTO(msgDTO *dto.MessageDTO) {
	if msgDTO.ReplyToID == nil {
		return
	}

	// ดึงข้อความที่ตอบกลับ
	replyMsg, err := s.messageRepo.GetByID(*msgDTO.ReplyToID)
	if err != nil || replyMsg == nil {
		return
	}

	// สร้างข้อมูลย่อของข้อความที่ตอบกลับ
	replyInfo := &dto.ReplyInfoDTO{
		ID:          replyMsg.ID.String(),
		MessageType: replyMsg.MessageType,
		Content:     replyMsg.Content,
		SenderID:    replyMsg.SenderID,
	}

	// เพิ่มข้อมูลผู้ส่งของข้อความที่ตอบกลับ
	if replyMsg.SenderID != nil {
		if replyMsg.SenderType == "business" && replyMsg.BusinessID != nil {
			business, err := s.businessRepo.GetByID(*replyMsg.BusinessID)
			if err == nil && business != nil {
				replyInfo.SenderName = business.Name
			}
		} else {
			user, err := s.userRepo.FindByID(*replyMsg.SenderID) // แก้ไขตรงนี้: เพิ่ม * เพื่อดึงค่าจาก pointer
			if err == nil && user != nil {
				if user.DisplayName != "" {
					replyInfo.SenderName = user.DisplayName
				} else {
					replyInfo.SenderName = user.Username
				}
			}
		}
	}

	msgDTO.ReplyToMessage = replyInfo
}

// GetMessageContext ดึงข้อความเป้าหมายพร้อมข้อความก่อนหน้าและถัดไป
func (s *conversationService) GetMessageContext(conversationID, userID uuid.UUID, targetID string,
	beforeCount, afterCount int) ([]*dto.MessageDTO, bool, bool, error) {

	// แปลง targetID เป็น uuid
	targetUUID, err := uuid.Parse(targetID)
	if err != nil {
		return nil, false, false, fmt.Errorf("invalid target message ID: %w", err)
	}

	// ตรวจสอบสิทธิ์การเข้าถึง
	isMember, err := s.conversationRepo.IsMember(conversationID, userID)
	if err != nil {
		return nil, false, false, err
	}

	if !isMember {
		return nil, false, false, errors.New("you are not a member of this conversation")
	}

	// ตรวจสอบว่ามีข้อความเป้าหมายจริงหรือไม่
	targetMsg, err := s.messageRepo.GetByID(targetUUID)
	if err != nil {
		return nil, false, false, fmt.Errorf("error fetching target message: %w", err)
	}

	if targetMsg == nil {
		return nil, false, false, errors.New("target message not found")
	}

	// ตรวจสอบว่าข้อความเป้าหมายอยู่ในการสนทนานี้หรือไม่
	if targetMsg.ConversationID != conversationID {
		return nil, false, false, errors.New("target message does not belong to this conversation")
	}

	// ดึงข้อความก่อนหน้าเป้าหมาย
	beforeMessages, err := s.messageRepo.GetMessagesBefore(conversationID, targetUUID, beforeCount+1) // +1 เพื่อตรวจสอบ hasMore
	if err != nil {
		return nil, false, false, fmt.Errorf("error fetching messages before target: %w", err)
	}

	// ตรวจสอบว่ามีข้อความเพิ่มเติมก่อนหน้าหรือไม่
	hasMoreBefore := len(beforeMessages) > beforeCount
	if hasMoreBefore {
		// ตัดข้อความส่วนเกิน
		beforeMessages = beforeMessages[:beforeCount]
	}

	// ดึงข้อความหลังเป้าหมาย
	afterMessages, err := s.messageRepo.GetMessagesAfter(conversationID, targetUUID, afterCount+1) // +1 เพื่อตรวจสอบ hasMore
	if err != nil {
		return nil, false, false, fmt.Errorf("error fetching messages after target: %w", err)
	}

	// ตรวจสอบว่ามีข้อความเพิ่มเติมหลังหรือไม่
	hasMoreAfter := len(afterMessages) > afterCount
	if hasMoreAfter {
		// ตัดข้อความส่วนเกิน
		afterMessages = afterMessages[:afterCount]
	}

	// รวมข้อความทั้งหมดและจัดเรียงตามเวลา
	allMessages := make([]*models.Message, 0, len(beforeMessages)+1+len(afterMessages))
	allMessages = append(allMessages, beforeMessages...)
	allMessages = append(allMessages, targetMsg)
	allMessages = append(allMessages, afterMessages...)

	// จัดเรียงข้อความตามเวลา (จากเก่าไปใหม่)
	sort.Slice(allMessages, func(i, j int) bool {
		return allMessages[i].CreatedAt.Before(allMessages[j].CreatedAt)
	})

	// แปลงเป็น DTOs โดยใช้ฟังก์ชันที่มีอยู่แล้ว
	messageDTOs := make([]*dto.MessageDTO, 0, len(allMessages))
	for _, msg := range allMessages {
		messageDTO, err := s.ConvertToMessageDTO(msg, userID)
		if err != nil {
			// ข้ามข้อความที่มีปัญหา
			continue
		}

		// เพิ่มการเน้นสำหรับข้อความเป้าหมาย (ถ้าต้องการ)
		if msg.ID == targetUUID {
			// ถ้ามีฟิลด์ IsHighlighted ใน MessageDTO ให้กำหนดค่าเป็น true
			// messageDTO.IsHighlighted = true
		}

		messageDTOs = append(messageDTOs, messageDTO)
	}

	return messageDTOs, hasMoreBefore, hasMoreAfter, nil
}

// GetMessagesBeforeID ดึงข้อความที่เก่ากว่า ID ที่ระบุ
func (s *conversationService) GetMessagesBeforeID(conversationID, userID uuid.UUID, beforeID string,
	limit int) ([]*dto.MessageDTO, int64, error) {

	// แปลง beforeID เป็น uuid
	beforeUUID, err := uuid.Parse(beforeID)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid before message ID: %w", err)
	}

	// ตรวจสอบสิทธิ์การเข้าถึง
	isMember, err := s.conversationRepo.IsMember(conversationID, userID)
	if err != nil {
		return nil, 0, err
	}

	if !isMember {
		return nil, 0, errors.New("you are not a member of this conversation")
	}

	// ดึงข้อความที่เก่ากว่า ID ที่ระบุ
	messages, err := s.messageRepo.GetMessagesBefore(conversationID, beforeUUID, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("error fetching messages before ID: %w", err)
	}

	// ดึงจำนวนข้อความทั้งหมดในการสนทนา (หรือจะใช้จำนวนที่น้อยกว่า ID นี้ก็ได้)
	total, err := s.messageRepo.CountAllMessages(conversationID)
	if err != nil {
		// ถ้ามีข้อผิดพลาดในการนับข้อความ ใช้ค่าประมาณ
		total = int64(len(messages))
	}

	// แปลงเป็น DTOs โดยใช้ฟังก์ชันที่มีอยู่แล้ว
	messageDTOs := make([]*dto.MessageDTO, 0, len(messages))
	for _, msg := range messages {
		messageDTO, err := s.ConvertToMessageDTO(msg, userID)
		if err != nil {
			// ข้ามข้อความที่มีปัญหา
			continue
		}
		messageDTOs = append(messageDTOs, messageDTO)
	}

	return messageDTOs, total, nil
}

// GetMessagesAfterID ดึงข้อความที่ใหม่กว่า ID ที่ระบุ
func (s *conversationService) GetMessagesAfterID(conversationID, userID uuid.UUID, afterID string,
	limit int) ([]*dto.MessageDTO, int64, error) {

	// แปลง afterID เป็น uuid
	afterUUID, err := uuid.Parse(afterID)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid after message ID: %w", err)
	}

	// ตรวจสอบสิทธิ์การเข้าถึง
	isMember, err := s.conversationRepo.IsMember(conversationID, userID)
	if err != nil {
		return nil, 0, err
	}

	if !isMember {
		return nil, 0, errors.New("you are not a member of this conversation")
	}

	// ดึงข้อความที่ใหม่กว่า ID ที่ระบุ
	messages, err := s.messageRepo.GetMessagesAfter(conversationID, afterUUID, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("error fetching messages after ID: %w", err)
	}

	// ดึงจำนวนข้อความทั้งหมดในการสนทนา
	total, err := s.messageRepo.CountAllMessages(conversationID)
	if err != nil {
		// ถ้ามีข้อผิดพลาดในการนับข้อความ ใช้ค่าประมาณ
		total = int64(len(messages))
	}

	// แปลงเป็น DTOs โดยใช้ฟังก์ชันที่มีอยู่แล้ว
	messageDTOs := make([]*dto.MessageDTO, 0, len(messages))
	for _, msg := range messages {
		messageDTO, err := s.ConvertToMessageDTO(msg, userID)
		if err != nil {
			// ข้ามข้อความที่มีปัญหา
			continue
		}
		messageDTOs = append(messageDTOs, messageDTO)
	}

	return messageDTOs, total, nil
}

// GetConversationsBeforeTime ดึงการสนทนาที่เก่ากว่าเวลาที่ระบุ
func (s *conversationService) GetConversationsBeforeTime(userID uuid.UUID, beforeTime string, limit int,
	convType string, pinned bool) ([]*dto.ConversationDTO, int, error) {

	// แปลง string เป็น time.Time
	parsedTime, err := time.Parse(time.RFC3339, beforeTime)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid time format: %w", err)
	}

	// เรียกใช้ repository
	conversations, total, err := s.conversationRepo.GetConversationsBeforeTime(
		userID, parsedTime, limit, convType, pinned)
	if err != nil {
		return nil, 0, err
	}

	// แปลงเป็น DTOs
	dtos := make([]*dto.ConversationDTO, 0, len(conversations))
	for _, conversation := range conversations {
		dto, err := s.convertToConversationDTO(conversation, userID)
		if err != nil {
			// ข้ามการสนทนาที่มีปัญหา
			continue
		}
		dtos = append(dtos, dto)
	}

	return dtos, total, nil
}

// GetConversationsAfterTime ดึงการสนทนาที่ใหม่กว่าเวลาที่ระบุ
func (s *conversationService) GetConversationsAfterTime(userID uuid.UUID, afterTime string, limit int,
	convType string, pinned bool) ([]*dto.ConversationDTO, int, error) {

	// แปลง string เป็น time.Time
	parsedTime, err := time.Parse(time.RFC3339, afterTime)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid time format: %w", err)
	}

	// เรียกใช้ repository
	conversations, total, err := s.conversationRepo.GetConversationsAfterTime(
		userID, parsedTime, limit, convType, pinned)
	if err != nil {
		return nil, 0, err
	}

	// แปลงเป็น DTOs
	dtos := make([]*dto.ConversationDTO, 0, len(conversations))
	for _, conversation := range conversations {
		dto, err := s.convertToConversationDTO(conversation, userID)
		if err != nil {
			// ข้ามการสนทนาที่มีปัญหา
			continue
		}
		dtos = append(dtos, dto)
	}

	return dtos, total, nil
}

// GetConversationsBeforeID ดึงการสนทนาที่เก่ากว่า ID ที่ระบุ
func (s *conversationService) GetConversationsBeforeID(userID, beforeID uuid.UUID, limit int,
	convType string, pinned bool) ([]*dto.ConversationDTO, int, error) {

	// เรียกใช้ repository
	conversations, total, err := s.conversationRepo.GetConversationsBeforeID(
		userID, beforeID, limit, convType, pinned)
	if err != nil {
		return nil, 0, err
	}

	// แปลงเป็น DTOs
	dtos := make([]*dto.ConversationDTO, 0, len(conversations))
	for _, conversation := range conversations {
		dto, err := s.convertToConversationDTO(conversation, userID)
		if err != nil {
			// ข้ามการสนทนาที่มีปัญหา
			continue
		}
		dtos = append(dtos, dto)
	}

	return dtos, total, nil
}

// GetConversationsAfterID ดึงการสนทนาที่ใหม่กว่า ID ที่ระบุ
func (s *conversationService) GetConversationsAfterID(userID, afterID uuid.UUID, limit int,
	convType string, pinned bool) ([]*dto.ConversationDTO, int, error) {

	// เรียกใช้ repository
	conversations, total, err := s.conversationRepo.GetConversationsAfterID(
		userID, afterID, limit, convType, pinned)
	if err != nil {
		return nil, 0, err
	}

	// แปลงเป็น DTOs
	dtos := make([]*dto.ConversationDTO, 0, len(conversations))
	for _, conversation := range conversations {
		dto, err := s.convertToConversationDTO(conversation, userID)
		if err != nil {
			// ข้ามการสนทนาที่มีปัญหา
			continue
		}
		dtos = append(dtos, dto)
	}

	return dtos, total, nil
}

// UpdateConversation อัปเดตข้อมูลการสนทนา
func (s *conversationService) UpdateConversation(id uuid.UUID, updateData types.JSONB) error {
	return s.conversationRepo.UpdateConversation(id, updateData)
}

// application/serviceimpl/conversation_service.go
// เพิ่มเมธอดเหล่านี้ใน conversationService struct ที่มีอยู่แล้ว

// ========================================
// 🏢 BUSINESS CONVERSATION SERVICE IMPLEMENTATIONS
// ========================================

// GetBusinessConversations ดึงการสนทนาทั้งหมดของธุรกิจ
func (s *conversationService) GetBusinessConversations(businessID uuid.UUID, adminID uuid.UUID, limit, offset int) ([]*dto.ConversationDTO, int, error) {
	// ดึงการสนทนาทั้งหมดของธุรกิจจาก repository
	conversations, total, err := s.conversationRepo.GetBusinessConversations(businessID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// แปลงเป็น DTOs
	dtos := make([]*dto.ConversationDTO, 0, len(conversations))
	for _, conversation := range conversations {
		// ใช้ฟังก์ชันเดิมในการแปลง แต่ส่ง businessID เพื่อ context
		dto, err := s.convertToBusinessConversationDTO(conversation, businessID, adminID)
		if err != nil {
			// ข้ามการสนทนาที่มีปัญหา
			continue
		}
		dtos = append(dtos, dto)
	}

	return dtos, total, nil
}

// GetBusinessConversationsBeforeTime ดึงการสนทนาธุรกิจที่เก่ากว่าเวลาที่ระบุ
func (s *conversationService) GetBusinessConversationsBeforeTime(businessID uuid.UUID, adminID uuid.UUID, beforeTime string, limit int) ([]*dto.ConversationDTO, int, error) {
	// แปลง string เป็น time.Time
	parsedTime, err := time.Parse(time.RFC3339, beforeTime)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid time format: %w", err)
	}

	// เรียกใช้ repository
	conversations, total, err := s.conversationRepo.GetBusinessConversationsBeforeTime(businessID, parsedTime, limit)
	if err != nil {
		return nil, 0, err
	}

	// แปลงเป็น DTOs
	dtos := make([]*dto.ConversationDTO, 0, len(conversations))
	for _, conversation := range conversations {
		dto, err := s.convertToBusinessConversationDTO(conversation, businessID, adminID)
		if err != nil {
			continue
		}
		dtos = append(dtos, dto)
	}

	return dtos, total, nil
}

// GetBusinessConversationsAfterTime ดึงการสนทนาธุรกิจที่ใหม่กว่าเวลาที่ระบุ
func (s *conversationService) GetBusinessConversationsAfterTime(businessID uuid.UUID, adminID uuid.UUID, afterTime string, limit int) ([]*dto.ConversationDTO, int, error) {
	// แปลง string เป็น time.Time
	parsedTime, err := time.Parse(time.RFC3339, afterTime)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid time format: %w", err)
	}

	// เรียกใช้ repository
	conversations, total, err := s.conversationRepo.GetBusinessConversationsAfterTime(businessID, parsedTime, limit)
	if err != nil {
		return nil, 0, err
	}

	// แปลงเป็น DTOs
	dtos := make([]*dto.ConversationDTO, 0, len(conversations))
	for _, conversation := range conversations {
		dto, err := s.convertToBusinessConversationDTO(conversation, businessID, adminID)
		if err != nil {
			continue
		}
		dtos = append(dtos, dto)
	}

	return dtos, total, nil
}

// GetBusinessConversationsBeforeID ดึงการสนทนาธุรกิจที่เก่ากว่า ID ที่ระบุ
func (s *conversationService) GetBusinessConversationsBeforeID(businessID uuid.UUID, adminID uuid.UUID, beforeID uuid.UUID, limit int) ([]*dto.ConversationDTO, int, error) {
	// เรียกใช้ repository
	conversations, total, err := s.conversationRepo.GetBusinessConversationsBeforeID(businessID, beforeID, limit)
	if err != nil {
		return nil, 0, err
	}

	// แปลงเป็น DTOs
	dtos := make([]*dto.ConversationDTO, 0, len(conversations))
	for _, conversation := range conversations {
		dto, err := s.convertToBusinessConversationDTO(conversation, businessID, adminID)
		if err != nil {
			continue
		}
		dtos = append(dtos, dto)
	}

	return dtos, total, nil
}

// GetBusinessConversationsAfterID ดึงการสนทนาธุรกิจที่ใหม่กว่า ID ที่ระบุ
func (s *conversationService) GetBusinessConversationsAfterID(businessID uuid.UUID, adminID uuid.UUID, afterID uuid.UUID, limit int) ([]*dto.ConversationDTO, int, error) {
	// เรียกใช้ repository
	conversations, total, err := s.conversationRepo.GetBusinessConversationsAfterID(businessID, afterID, limit)
	if err != nil {
		return nil, 0, err
	}

	// แปลงเป็น DTOs
	dtos := make([]*dto.ConversationDTO, 0, len(conversations))
	for _, conversation := range conversations {
		dto, err := s.convertToBusinessConversationDTO(conversation, businessID, adminID)
		if err != nil {
			continue
		}
		dtos = append(dtos, dto)
	}

	return dtos, total, nil
}

// GetBusinessConversationMessages ดึงข้อความในการสนทนาธุรกิจ
func (s *conversationService) GetBusinessConversationMessages(conversationID, businessID uuid.UUID, limit, offset int) ([]*dto.MessageDTO, int64, error) {
	// ตรวจสอบว่าการสนทนาเป็นของธุรกิจ
	belongsToBusiness, err := s.conversationRepo.CheckConversationBelongsToBusiness(conversationID, businessID)
	if err != nil {
		return nil, 0, err
	}
	if !belongsToBusiness {
		return nil, 0, errors.New("this conversation does not belong to your business")
	}

	// ดึงข้อความทั้งหมดในการสนทนา (ใช้ฟังก์ชันเดิม)
	messages, total, err := s.messageRepo.GetMessagesByConversationID(conversationID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// แปลงข้อความเป็น DTOs
	messageDTOs := make([]*dto.MessageDTO, 0, len(messages))
	for _, msg := range messages {
		// ใช้ฟังก์ชันเดิมในการแปลง แต่ส่ง businessID เพื่อ context
		messageDTO, err := s.ConvertToBusinessMessageDTO(msg, businessID)
		if err != nil {
			// ข้ามข้อความที่มีปัญหา
			continue
		}
		messageDTOs = append(messageDTOs, messageDTO)
	}

	return messageDTOs, total, nil
}

// GetBusinessMessageContext ดึงข้อความเป้าหมายพร้อมบริบทสำหรับธุรกิจ
func (s *conversationService) GetBusinessMessageContext(conversationID, businessID uuid.UUID, targetID string, beforeCount, afterCount int) ([]*dto.MessageDTO, bool, bool, error) {
	// แปลง targetID เป็น uuid
	targetUUID, err := uuid.Parse(targetID)
	if err != nil {
		return nil, false, false, fmt.Errorf("invalid target message ID: %w", err)
	}

	// ตรวจสอบว่าการสนทนาเป็นของธุรกิจ
	belongsToBusiness, err := s.conversationRepo.CheckConversationBelongsToBusiness(conversationID, businessID)
	if err != nil {
		return nil, false, false, err
	}
	if !belongsToBusiness {
		return nil, false, false, errors.New("this conversation does not belong to your business")
	}

	// ตรวจสอบว่ามีข้อความเป้าหมายจริงหรือไม่
	targetMsg, err := s.messageRepo.GetByID(targetUUID)
	if err != nil {
		return nil, false, false, fmt.Errorf("error fetching target message: %w", err)
	}
	if targetMsg == nil {
		return nil, false, false, errors.New("target message not found")
	}
	if targetMsg.ConversationID != conversationID {
		return nil, false, false, errors.New("target message does not belong to this conversation")
	}

	// ดึงข้อความก่อนหน้าเป้าหมาย
	beforeMessages, err := s.messageRepo.GetMessagesBefore(conversationID, targetUUID, beforeCount+1)
	if err != nil {
		return nil, false, false, fmt.Errorf("error fetching messages before target: %w", err)
	}
	hasMoreBefore := len(beforeMessages) > beforeCount
	if hasMoreBefore {
		beforeMessages = beforeMessages[:beforeCount]
	}

	// ดึงข้อความหลังเป้าหมาย
	afterMessages, err := s.messageRepo.GetMessagesAfter(conversationID, targetUUID, afterCount+1)
	if err != nil {
		return nil, false, false, fmt.Errorf("error fetching messages after target: %w", err)
	}
	hasMoreAfter := len(afterMessages) > afterCount
	if hasMoreAfter {
		afterMessages = afterMessages[:afterCount]
	}

	// รวมข้อความทั้งหมดและจัดเรียง
	allMessages := make([]*models.Message, 0, len(beforeMessages)+1+len(afterMessages))
	allMessages = append(allMessages, beforeMessages...)
	allMessages = append(allMessages, targetMsg)
	allMessages = append(allMessages, afterMessages...)

	// จัดเรียงข้อความตามเวลา
	sort.Slice(allMessages, func(i, j int) bool {
		return allMessages[i].CreatedAt.Before(allMessages[j].CreatedAt)
	})

	// แปลงเป็น DTOs
	messageDTOs := make([]*dto.MessageDTO, 0, len(allMessages))
	for _, msg := range allMessages {
		messageDTO, err := s.ConvertToBusinessMessageDTO(msg, businessID)
		if err != nil {
			continue
		}
		messageDTOs = append(messageDTOs, messageDTO)
	}

	return messageDTOs, hasMoreBefore, hasMoreAfter, nil
}

// GetBusinessMessagesBeforeID ดึงข้อความธุรกิจที่เก่ากว่า ID ที่ระบุ
func (s *conversationService) GetBusinessMessagesBeforeID(conversationID, businessID uuid.UUID, beforeID string, limit int) ([]*dto.MessageDTO, int64, error) {
	// แปลง beforeID เป็น uuid
	beforeUUID, err := uuid.Parse(beforeID)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid before message ID: %w", err)
	}

	// ตรวจสอบว่าการสนทนาเป็นของธุรกิจ
	belongsToBusiness, err := s.conversationRepo.CheckConversationBelongsToBusiness(conversationID, businessID)
	if err != nil {
		return nil, 0, err
	}
	if !belongsToBusiness {
		return nil, 0, errors.New("this conversation does not belong to your business")
	}

	// ดึงข้อความที่เก่ากว่า ID ที่ระบุ
	messages, err := s.messageRepo.GetMessagesBefore(conversationID, beforeUUID, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("error fetching messages before ID: %w", err)
	}

	// ดึงจำนวนข้อความทั้งหมด
	total, err := s.messageRepo.CountAllMessages(conversationID)
	if err != nil {
		total = int64(len(messages))
	}

	// แปลงเป็น DTOs
	messageDTOs := make([]*dto.MessageDTO, 0, len(messages))
	for _, msg := range messages {
		messageDTO, err := s.ConvertToBusinessMessageDTO(msg, businessID)
		if err != nil {
			continue
		}
		messageDTOs = append(messageDTOs, messageDTO)
	}

	return messageDTOs, total, nil
}

// GetBusinessMessagesAfterID ดึงข้อความธุรกิจที่ใหม่กว่า ID ที่ระบุ
func (s *conversationService) GetBusinessMessagesAfterID(conversationID, businessID uuid.UUID, afterID string, limit int) ([]*dto.MessageDTO, int64, error) {
	// แปลง afterID เป็น uuid
	afterUUID, err := uuid.Parse(afterID)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid after message ID: %w", err)
	}

	// ตรวจสอบว่าการสนทนาเป็นของธุรกิจ
	belongsToBusiness, err := s.conversationRepo.CheckConversationBelongsToBusiness(conversationID, businessID)
	if err != nil {
		return nil, 0, err
	}
	if !belongsToBusiness {
		return nil, 0, errors.New("this conversation does not belong to your business")
	}

	// ดึงข้อความที่ใหม่กว่า ID ที่ระบุ
	messages, err := s.messageRepo.GetMessagesAfter(conversationID, afterUUID, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("error fetching messages after ID: %w", err)
	}

	// ดึงจำนวนข้อความทั้งหมด
	total, err := s.messageRepo.CountAllMessages(conversationID)
	if err != nil {
		total = int64(len(messages))
	}

	// แปลงเป็น DTOs
	messageDTOs := make([]*dto.MessageDTO, 0, len(messages))
	for _, msg := range messages {
		messageDTO, err := s.ConvertToBusinessMessageDTO(msg, businessID)
		if err != nil {
			continue
		}
		messageDTOs = append(messageDTOs, messageDTO)
	}

	return messageDTOs, total, nil
}

// CheckConversationBelongsToBusiness ตรวจสอบว่าการสนทนาเป็นของธุรกิจ
func (s *conversationService) CheckConversationBelongsToBusiness(conversationID, businessID uuid.UUID) (bool, error) {
	return s.conversationRepo.CheckConversationBelongsToBusiness(conversationID, businessID)
}

// ========================================
// 🔧 HELPER FUNCTIONS สำหรับ Business Context
// ========================================

// convertToBusinessConversationDTO แปลง Conversation model เป็น DTO สำหรับ business context
func (s *conversationService) convertToBusinessConversationDTO(conversation *models.Conversation, businessID uuid.UUID, adminID uuid.UUID) (*dto.ConversationDTO, error) {
	if conversation == nil {
		return nil, errors.New("conversation is nil")
	}

	convDTO := &dto.ConversationDTO{
		ID:              conversation.ID,
		Type:            conversation.Type,
		Title:           conversation.Title,
		IconURL:         conversation.IconURL,
		CreatedAt:       conversation.CreatedAt,
		UpdatedAt:       conversation.UpdatedAt,
		LastMessageText: conversation.LastMessageText,
		LastMessageAt:   conversation.LastMessageAt,
		CreatorID:       conversation.CreatorID,
		BusinessID:      conversation.BusinessID,
		IsActive:        conversation.IsActive,
		Metadata:        conversation.Metadata,
	}

	// เพิ่มข้อมูลสำหรับ business conversation
	if conversation.Type == "business" {
		// ดึงข้อมูลลูกค้า (user ที่ไม่ใช่ business admin)
		members, err := s.conversationRepo.GetMembers(conversation.ID)
		if err == nil && len(members) > 0 {
			for _, member := range members {
				// หาสมาชิกที่ไม่ใช่ business admin
				user, err := s.userRepo.FindByID(member.UserID)
				if err == nil && user != nil {
					// ตรวจสอบว่าเป็น customer (ไม่ใช่ admin ของธุรกิจ)
					isBusinessAdmin, err := s.businessAdminRepo.CheckAdminPermission(member.UserID, businessID, []string{})
					if err != nil || !isBusinessAdmin {
						// นี่คือลูกค้า
						// 1. ลองหา customer profile ก่อน
						customerProfile, profileErr := s.customerProfileRepo.GetByBusinessAndUser(businessID, member.UserID)

						// 2. ตั้งค่า title ตาม nickname ก่อน ถ้ามี
						if profileErr == nil && customerProfile != nil && customerProfile.Nickname != "" {
							convDTO.Title = customerProfile.Nickname
						} else {
							// 3. ถ้าไม่มี nickname หรือมีปัญหา ใช้ display_name ตามเดิม
							convDTO.Title = user.DisplayName
							if convDTO.Title == "" {
								convDTO.Title = user.Username
							}
						}

						// ใช้ profile image ตามเดิม
						convDTO.IconURL = user.ProfileImageURL

						// เพิ่มข้อมูลลูกค้า - เพิ่มข้อมูล nickname เข้าไปด้วย
						contactInfo := types.JSONB{
							"user_id":           user.ID.String(),
							"username":          user.Username,
							"display_name":      user.DisplayName,
							"profile_image_url": user.ProfileImageURL,
						}

						// เพิ่ม nickname เข้าไปใน contactInfo ถ้ามี
						if profileErr == nil && customerProfile != nil && customerProfile.Nickname != "" {
							contactInfo["nickname"] = customerProfile.Nickname
						}

						convDTO.ContactInfo = contactInfo
						break
					}
				}
			}
		}

		// ❌ ไม่ส่ง business_info เพราะไม่จำเป็นในมุมมองของธุรกิจ
		// Admin รู้อยู่แล้วว่าเป็นธุรกิจไหน
	}

	// จำนวนสมาชิก
	members, err := s.conversationRepo.GetMembers(conversation.ID)
	if err == nil {
		convDTO.MemberCount = len(members)
	}

	// สำหรับธุรกิจ คำนวณ unread_count แบบพิเศษ
	// นับข้อความที่ลูกค้าส่งมาแต่ธุรกิจยังไม่ได้อ่าน
	var unreadCount int

	// หาแอดมินคนสุดท้ายที่อ่านข้อความ
	var lastBusinessReadTime *time.Time
	for _, member := range members {
		// ตรวจสอบว่าเป็นแอดมินของธุรกิจหรือไม่
		isBusinessAdmin, err := s.businessAdminRepo.CheckAdminPermission(member.UserID, businessID, []string{})
		if err == nil && isBusinessAdmin && member.LastReadAt != nil {
			// หาเวลาอ่านล่าสุดของแอดมิน
			if lastBusinessReadTime == nil || member.LastReadAt.After(*lastBusinessReadTime) {
				lastBusinessReadTime = member.LastReadAt
			}
		}
	}

	// นับข้อความที่ส่งหลังจากเวลาอ่านล่าสุดของแอดมิน
	if lastBusinessReadTime != nil {
		// นับข้อความจากลูกค้าที่ส่งหลังจากเวลาอ่านล่าสุด
		customerMessages, err := s.messageRepo.GetCustomerMessagesAfterTime(
			conversation.ID, *lastBusinessReadTime, businessID)
		if err == nil {
			unreadCount = len(customerMessages)
		}
	} else {
		// ถ้าไม่มีการอ่านเลย นับข้อความจากลูกค้าทั้งหมด
		customerMessages, err := s.messageRepo.GetAllCustomerMessages(
			conversation.ID, businessID)
		if err == nil {
			unreadCount = len(customerMessages)
		}
	}

	convDTO.UnreadCount = unreadCount

	// ตรวจสอบสถานะ pin/mute ของแอดมิน
	member, err := s.conversationRepo.GetMember(conversation.ID, adminID)
	if err == nil && member != nil {
		convDTO.IsPinned = member.IsPinned
		convDTO.IsMuted = member.IsMuted
	} else {
		// กรณีไม่พบข้อมูล ใช้ค่าเริ่มต้น
		convDTO.IsPinned = false
		convDTO.IsMuted = false
	}

	return convDTO, nil
}

// ConvertToBusinessMessageDTO แปลง Message model เป็น DTO สำหรับ business context
func (s *conversationService) ConvertToBusinessMessageDTO(msg *models.Message, businessID uuid.UUID) (*dto.MessageDTO, error) {
	if msg == nil {
		return nil, errors.New("message is nil")
	}

	messageDTO := &dto.MessageDTO{
		ID:                msg.ID,
		ConversationID:    msg.ConversationID,
		SenderID:          msg.SenderID,
		SenderType:        msg.SenderType,
		MessageType:       msg.MessageType,
		Content:           msg.Content,
		MediaURL:          msg.MediaURL,
		MediaThumbnailURL: msg.MediaThumbnailURL,
		Metadata:          msg.Metadata,
		CreatedAt:         msg.CreatedAt,
		UpdatedAt:         msg.UpdatedAt,
		IsDeleted:         msg.IsDeleted,
		IsEdited:          msg.IsEdited,
		EditCount:         msg.EditCount,
		ReplyToID:         msg.ReplyToID,
		BusinessID:        msg.BusinessID,
		ReadCount:         0,     // ค่าเริ่มต้น จะอัปเดตทีหลัง
		IsRead:            false, // ค่าเริ่มต้น จะอัปเดตทีหลัง
	}

	// 1. เพิ่มข้อมูลผู้ส่ง (อาจต้องปรับในอนาคตสำหรับ business context)
	s.addSenderInfoToDTO(messageDTO)

	// 2. เพิ่มข้อมูลสถานะการอ่านสำหรับ business context
	s.addBusinessReadStatusToDTO(messageDTO, businessID)

	// 3. เพิ่มข้อมูลข้อความที่ตอบกลับ
	if msg.ReplyToID != nil {
		s.addReplyToInfoToDTO(messageDTO)
	}

	// 🚀 ในอนาคตอาจเพิ่มเติม:
	// s.addBusinessAdminInfoToDTO(messageDTO, businessID)
	// s.addBusinessAnalyticsToDTO(messageDTO, businessID)

	return messageDTO, nil
}

// addBusinessReadStatusToDTO เพิ่มข้อมูลสถานะการอ่านสำหรับ business context
func (s *conversationService) addBusinessReadStatusToDTO(msgDTO *dto.MessageDTO, businessID uuid.UUID) {
	// ดึงข้อมูลการอ่านทั้งหมดของข้อความนี้
	reads, err := s.messageRepo.GetReads(msgDTO.ID)
	if err != nil {
		return
	}

	// คำนวณ ReadCount (จำนวนการอ่านทั้งหมด)
	msgDTO.ReadCount = len(reads)

	// ตรวจสอบว่ามีแอดมินของธุรกิจอ่านข้อความแล้วหรือยัง
	// Logic นี้แตกต่างจาก addReadStatusToDTO แบบปกติ
	for _, read := range reads {
		// ตรวจสอบว่า user ที่อ่านเป็นแอดมินของธุรกิจหรือไม่
		isBusinessAdmin, err := s.businessAdminRepo.CheckAdminPermission(read.UserID, businessID, []string{})
		if err == nil && isBusinessAdmin {
			msgDTO.IsRead = true
			break
		}
	}

	// 💡 ในอนาคตอาจเพิ่มข้อมูลเพิ่มเติม:
	// - ดูว่าแอดมินคนไหนอ่านบ้าง
	// - เวลาที่อ่าน
	// - การแจ้งเตือนไปยังลูกค้า
}
