// application/serviceimpl/business_account_service.go
package serviceimpl

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/dto"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/domain/types"
)

type businessAccountService struct {
	businessAccountRepo repository.BusinessAccountRepository
	adminRepo           repository.BusinessAdminRepository
	userRepo            repository.UserRepository
}

// NewBusinessAccountService สร้าง instance ใหม่ของ BusinessAccountService
func NewBusinessAccountService(
	businessAccountRepo repository.BusinessAccountRepository,
	adminRepo repository.BusinessAdminRepository,
	userRepo repository.UserRepository,
) service.BusinessAccountService {
	return &businessAccountService{
		businessAccountRepo: businessAccountRepo,
		adminRepo:           adminRepo,
		userRepo:            userRepo,
	}
}

// GetBusinessByID ดึงข้อมูลธุรกิจตาม ID
func (s *businessAccountService) GetBusinessByID(id uuid.UUID, userID uuid.UUID) (*models.BusinessAccount, error) {
	// ตรวจสอบสิทธิ์
	hasPermission, err := s.adminRepo.CheckAdminPermission(userID, id, []string{})
	if err != nil {
		return nil, err
	}

	if !hasPermission {
		return nil, errors.New("you don't have permission to access this business x1")
	}

	// ดึงข้อมูลธุรกิจ
	business, err := s.businessAccountRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// ตรวจสอบสถานะ
	if business.Status != "active" {
		return nil, errors.New("business not found or inactive")
	}

	// ดึงจำนวนผู้ติดตาม
	followerCount, err := s.businessAccountRepo.GetFollowerCount(id)
	if err == nil {
		if business.Settings == nil {
			business.Settings = types.JSONB{}
		}
		business.Settings["follower_count"] = followerCount
	}

	return business, nil
}

// GetBusinessByUsername ดึงข้อมูลธุรกิจตาม username
func (s *businessAccountService) GetBusinessByUsername(username string, userID uuid.UUID) (*models.BusinessAccount, error) {
	// ดึงข้อมูลธุรกิจ
	business, err := s.businessAccountRepo.GetByUsername(username)
	if err != nil {
		return nil, err
	}

	// ตรวจสอบสถานะ
	if business.Status != "active" {
		return nil, errors.New("business not found or inactive")
	}

	// ตรวจสอบการติดตาม
	isFollowing, err := s.businessAccountRepo.IsFollowing(userID, business.ID)
	if err == nil && isFollowing {
		// ในโมเดลไม่มี IsFollowing จึงเก็บใน Settings สำหรับส่งกลับ
		if business.Settings == nil {
			business.Settings = types.JSONB{}
		}
		business.Settings["is_following"] = true
	}

	// ดึงจำนวนผู้ติดตาม
	followerCount, err := s.businessAccountRepo.GetFollowerCount(business.ID)
	if err == nil {
		if business.Settings == nil {
			business.Settings = types.JSONB{}
		}
		business.Settings["follower_count"] = followerCount
	}

	return business, nil
}

// GetUserBusinesses ดึงรายการธุรกิจที่ผู้ใช้เป็นแอดมิน
func (s *businessAccountService) GetUserBusinesses(userID uuid.UUID) ([]*models.BusinessAccount, error) {
	businesses, err := s.businessAccountRepo.GetBusinessesByUserID(userID)
	if err != nil {
		return nil, err
	}

	// ดึงบทบาทของผู้ใช้สำหรับแต่ละธุรกิจ
	for _, business := range businesses {
		// ตัวอย่างการเพิ่มข้อมูลบทบาทลงใน Settings
		admins, err := s.adminRepo.GetAdminsByBusinessID(business.ID)
		if err != nil {
			continue
		}

		for _, admin := range admins {
			if admin.UserID == userID {
				if business.Settings == nil {
					business.Settings = make(map[string]interface{})
				}
				business.Settings["user_role"] = admin.Role
				break
			}
		}
	}

	return businesses, nil
}

// CreateBusiness สร้างธุรกิจใหม่
func (s *businessAccountService) CreateBusiness(userID uuid.UUID, name string, username string, description string, welcomeMessage string) (*models.BusinessAccount, error) {
	// ตรวจสอบรูปแบบ username
	if !isValidUsername(username) {
		return nil, errors.New("invalid username format")
	}

	// สร้างข้อมูลธุรกิจ
	business := &models.BusinessAccount{
		ID:          uuid.New(),
		Name:        name,
		Username:    username,
		Description: description,
		//WelcomeMessage: welcomeMessage,
		OwnerID:   &userID,
		CreatedAt: time.Now(),
		Status:    "active",
	}

	// บันทึกข้อมูลธุรกิจและสร้างแอดมิน
	err := s.businessAccountRepo.Create(business)
	if err != nil {
		return nil, err
	}

	return business, nil
}

// UpdateBusiness อัพเดทข้อมูลธุรกิจ
func (s *businessAccountService) UpdateBusiness(id uuid.UUID, userID uuid.UUID, updateData types.JSONB) (*models.BusinessAccount, error) {
	// ตรวจสอบสิทธิ์
	hasPermission, err := s.adminRepo.CheckAdminPermission(userID, id, []string{"owner", "admin"})
	if err != nil {
		return nil, err
	}

	if !hasPermission {
		return nil, errors.New("you don't have permission to update this business")
	}

	// ดึงข้อมูลธุรกิจปัจจุบัน
	business, err := s.businessAccountRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// อัพเดทข้อมูล
	if name, ok := updateData["name"].(string); ok && name != "" {
		business.Name = name
	}

	if description, ok := updateData["description"].(string); ok {
		business.Description = description
	}

	if profileImageURL, ok := updateData["profile_image_url"].(string); ok {
		business.ProfileImageURL = profileImageURL
	}

	if coverImageURL, ok := updateData["cover_image_url"].(string); ok {
		business.CoverImageURL = coverImageURL
	}

	// เพิ่มการรองรับการอัพเดทสถานะ
	if status, ok := updateData["status"].(string); ok && (status == "active" || status == "deleted") {
		business.Status = status
	}

	/*
		if welcomeMessage, ok := updateData["welcome_message"].(string); ok {
			business.WelcomeMessage = welcomeMessage
		}
	*/

	// บันทึกการอัพเดท
	err = s.businessAccountRepo.Update(business)
	if err != nil {
		return nil, err
	}

	return business, nil
}

// DeleteBusiness ลบธุรกิจ
func (s *businessAccountService) DeleteBusiness(id uuid.UUID, userID uuid.UUID) error {
	// ตรวจสอบสิทธิ์ (ต้องเป็น owner เท่านั้น)
	hasPermission, err := s.adminRepo.CheckAdminPermission(userID, id, []string{"owner"})
	if err != nil {
		return err
	}

	if !hasPermission {
		return errors.New("only the owner can delete this business")
	}

	// ลบธุรกิจ (เปลี่ยนสถานะเป็น deleted)
	return s.businessAccountRepo.Delete(id)
}

// isValidUsername ตรวจสอบความถูกต้องของ username
func isValidUsername(username string) bool {
	// ตรวจสอบความยาว (3-30 ตัวอักษร)
	if len(username) < 3 || len(username) > 30 {
		return false
	}

	// ตรวจสอบว่ามีเฉพาะตัวอักษร ตัวเลข และ underscore
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}

	return true
}

// UploadBusinessProfileImage อัปโหลดรูปโปรไฟล์ธุรกิจ
func (s *businessAccountService) UploadBusinessProfileImage(id uuid.UUID, userID uuid.UUID, imageURL string) error {
	// ตรวจสอบสิทธิ์
	hasPermission, err := s.adminRepo.CheckAdminPermission(userID, id, []string{"owner", "admin"})
	if err != nil {
		return err
	}

	if !hasPermission {
		return errors.New("you don't have permission to update this business")
	}

	// ดึงข้อมูลธุรกิจปัจจุบัน
	business, err := s.businessAccountRepo.GetByID(id)
	if err != nil {
		return err
	}

	// อัปเดตรูปโปรไฟล์
	business.ProfileImageURL = imageURL

	// บันทึกการอัปเดท
	return s.businessAccountRepo.Update(business)
}

// UploadBusinessCoverImage อัปโหลดรูปปกธุรกิจ
func (s *businessAccountService) UploadBusinessCoverImage(id uuid.UUID, userID uuid.UUID, imageURL string) error {
	// ตรวจสอบสิทธิ์
	hasPermission, err := s.adminRepo.CheckAdminPermission(userID, id, []string{"owner", "admin"})
	if err != nil {
		return err
	}

	if !hasPermission {
		return errors.New("you don't have permission to update this business")
	}

	// ดึงข้อมูลธุรกิจปัจจุบัน
	business, err := s.businessAccountRepo.GetByID(id)
	if err != nil {
		return err
	}

	// อัปเดตรูปปก
	business.CoverImageURL = imageURL

	// บันทึกการอัปเดท
	return s.businessAccountRepo.Update(business)
}

// SearchBusinesses ค้นหาธุรกิจ
func (s *businessAccountService) SearchBusinesses(query string, limit, offset int, userID uuid.UUID) ([]*models.BusinessAccount, int64, error) {
	// ค้นหาธุรกิจ
	businesses, total, err := s.businessAccountRepo.SearchBusinesses(query, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// เพิ่มข้อมูลเพิ่มเติมสำหรับแต่ละธุรกิจ
	for _, business := range businesses {
		// ตรวจสอบการติดตาม
		isFollowing, err := s.businessAccountRepo.IsFollowing(userID, business.ID)
		if err == nil && isFollowing {
			// ในโมเดลไม่มี IsFollowing จึงเก็บใน Settings สำหรับส่งกลับ
			if business.Settings == nil {
				business.Settings = types.JSONB{}
			}
			business.Settings["is_following"] = true
		}

		// ดึงจำนวนผู้ติดตาม
		followerCount, err := s.businessAccountRepo.GetFollowerCount(business.ID)
		if err == nil {
			if business.Settings == nil {
				business.Settings = types.JSONB{}
			}
			business.Settings["follower_count"] = followerCount
		}
	}

	return businesses, total, nil
}

// GetUserBusinessDTOs ดึงรายการธุรกิจที่ผู้ใช้เป็นแอดมินและแปลงเป็น DTO
// GetUserBusinessDTOs ดึงรายการธุรกิจที่ผู้ใช้เป็นแอดมินและแปลงเป็น DTO
func (s *businessAccountService) GetUserBusinessDTOs(userID uuid.UUID) ([]dto.BusinessItem, error) {
	businesses, err := s.businessAccountRepo.GetBusinessesByUserID(userID)
	if err != nil {
		return nil, err
	}

	// แปลงจาก model เป็น DTO โดยตรง
	businessDTOs := make([]dto.BusinessItem, len(businesses))
	for i, business := range businesses {
		// กำหนดค่าแต่ละ field โดยตรง
		var profileImageURL, coverImageURL *string
		if business.ProfileImageURL != "" {
			profileImageURL = &business.ProfileImageURL
		}
		if business.CoverImageURL != "" {
			coverImageURL = &business.CoverImageURL
		}

		// สำหรับข้อมูลพื้นฐาน
		isAdmin := false
		adminRole := ""

		// ดึงข้อมูลจาก settings
		if business.Settings != nil {
			if role, ok := business.Settings["user_role"].(string); ok {
				isAdmin = true
				adminRole = role
			}
		}

		// ดึงจำนวนผู้ติดตามจากฐานข้อมูลโดยตรง
		followerCount := 0
		count, err := s.businessAccountRepo.GetFollowerCount(business.ID)
		if err == nil {
			followerCount = int(count)
		}

		// ตรวจสอบการติดตามจากฐานข้อมูลโดยตรง
		isFollowing := false
		following, err := s.businessAccountRepo.IsFollowing(userID, business.ID)
		if err == nil {
			isFollowing = following
		}

		businessDTOs[i] = dto.BusinessItem{
			ID:              business.ID,
			Name:            business.Name,
			Username:        business.Username,
			Description:     business.Description,
			ProfileImageURL: profileImageURL,
			CoverImageURL:   coverImageURL,
			CreatedAt:       business.CreatedAt,
			OwnerID:         business.OwnerID,
			Status:          business.Status,
			Settings:        business.Settings,
			FollowerCount:   followerCount,
			IsFollowing:     isFollowing,
			IsAdmin:         isAdmin,
			AdminRole:       adminRole,
		}
	}

	return businessDTOs, nil
}

// GetBusinessByUsernameExact ดึงข้อมูลธุรกิจตาม username แบบตรงกับทั้งหมด
func (s *businessAccountService) GetBusinessByUsernameExact(username string, userID uuid.UUID) (*models.BusinessAccount, error) {
	// ดึงข้อมูลธุรกิจ
	business, err := s.businessAccountRepo.GetByUsernameExact(username)
	if err != nil {
		return nil, err
	}

	// ตรวจสอบสถานะ
	if business.Status != "active" {
		return nil, errors.New("business not found or inactive")
	}

	// ตรวจสอบการติดตาม
	isFollowing, err := s.businessAccountRepo.IsFollowing(userID, business.ID)
	if err == nil && isFollowing {
		// ในโมเดลไม่มี IsFollowing จึงเก็บใน Settings สำหรับส่งกลับ
		if business.Settings == nil {
			business.Settings = types.JSONB{}
		}
		business.Settings["is_following"] = true
	}

	// ดึงจำนวนผู้ติดตาม
	followerCount, err := s.businessAccountRepo.GetFollowerCount(business.ID)
	if err == nil {
		if business.Settings == nil {
			business.Settings = types.JSONB{}
		}
		business.Settings["follower_count"] = followerCount
	}

	return business, nil
}

// SearchBusinessesExact ค้นหาธุรกิจแบบตรงกับทั้งหมด
func (s *businessAccountService) SearchBusinessesExact(query string, limit, offset int, userID uuid.UUID) ([]*models.BusinessAccount, int64, error) {
	// ค้นหาธุรกิจ
	businesses, total, err := s.businessAccountRepo.SearchBusinessesExact(query, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// เพิ่มข้อมูลเพิ่มเติมสำหรับแต่ละธุรกิจ
	for _, business := range businesses {
		// ตรวจสอบการติดตาม
		isFollowing, err := s.businessAccountRepo.IsFollowing(userID, business.ID)
		if err == nil && isFollowing {
			// ในโมเดลไม่มี IsFollowing จึงเก็บใน Settings สำหรับส่งกลับ
			if business.Settings == nil {
				business.Settings = types.JSONB{}
			}
			business.Settings["is_following"] = true
		}

		// ดึงจำนวนผู้ติดตาม
		followerCount, err := s.businessAccountRepo.GetFollowerCount(business.ID)
		if err == nil {
			if business.Settings == nil {
				business.Settings = types.JSONB{}
			}
			business.Settings["follower_count"] = followerCount
		}
	}

	return businesses, total, nil
}
