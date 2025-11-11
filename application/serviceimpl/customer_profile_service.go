// application/serviceimpl/customer_profile_service.go
package serviceimpl

import (
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/domain/types"
)

type customerProfileService struct {
	customerProfileRepo repository.CustomerProfileRepository
	businessAdminRepo   repository.BusinessAdminRepository
	businessAccountRepo repository.BusinessAccountRepository
	userRepo            repository.UserRepository
	userTagRepo         repository.UserTagRepository
}

// NewCustomerProfileService สร้าง instance ใหม่ของ CustomerProfileService
func NewCustomerProfileService(
	customerProfileRepo repository.CustomerProfileRepository,
	businessAdminRepo repository.BusinessAdminRepository,
	businessAccountRepo repository.BusinessAccountRepository,
	userRepo repository.UserRepository,
	userTagRepo repository.UserTagRepository,
) service.CustomerProfileService {
	return &customerProfileService{
		customerProfileRepo: customerProfileRepo,
		businessAdminRepo:   businessAdminRepo,
		businessAccountRepo: businessAccountRepo,
		userRepo:            userRepo,
		userTagRepo:         userTagRepo,
	}
}

// CreateCustomerProfile สร้างโปรไฟล์ลูกค้าใหม่
func (s *customerProfileService) CreateCustomerProfile(businessID, userID uuid.UUID, nickname, notes, customerType string) (*models.CustomerProfile, error) {
	// ตรวจสอบว่าธุรกิจมีอยู่จริง
	exists, err := s.businessAccountRepo.ExistsById(businessID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("business not found")
	}

	// ตรวจสอบว่าผู้ใช้มีอยู่จริง
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// ตรวจสอบว่ามี CustomerProfile อยู่แล้วหรือไม่
	existingProfile, err := s.customerProfileRepo.GetByBusinessAndUser(businessID, userID)
	if err == nil && existingProfile != nil {
		return existingProfile, nil // ส่งคืนโปรไฟล์ที่มีอยู่แล้ว
	}

	// ตั้งค่าเริ่มต้น
	if customerType == "" {
		customerType = "New" // ค่าเริ่มต้น
	}

	// สร้าง CustomerProfile ใหม่
	now := time.Now()
	profile := &models.CustomerProfile{
		ID:            uuid.New(),
		BusinessID:    businessID,
		UserID:        userID,
		Nickname:      nickname,
		Notes:         notes,
		CustomerType:  customerType,
		Status:        "active",
		LastContactAt: &now, // ตั้งเป็นเวลาปัจจุบันเมื่อสร้าง
		CreatedAt:     now,
		UpdatedAt:     now,
		User:          user,
	}

	// บันทึกลงฐานข้อมูล
	err = s.customerProfileRepo.Create(profile)
	if err != nil {
		return nil, err
	}

	return profile, nil
}

// GetCustomerProfile ดึงโปรไฟล์ลูกค้า
func (s *customerProfileService) GetCustomerProfile(businessID, userID uuid.UUID) (*models.CustomerProfile, error) {
	// ตรวจสอบว่าธุรกิจมีอยู่จริง
	exists, err := s.businessAccountRepo.ExistsById(businessID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("business not found")
	}

	// ดึงโปรไฟล์ลูกค้า
	profile, err := s.customerProfileRepo.GetByBusinessAndUser(businessID, userID)
	if err != nil {
		return nil, err
	}

	// ดึง tags ของลูกค้า
	if profile.UserID != uuid.Nil {
		userTags, err := s.userTagRepo.GetUserTags(businessID, profile.UserID)
		if err != nil {
			// Log error แต่ไม่หยุดการทำงาน
			// log.Printf("Error fetching tags for user %s: %v", profile.UserID, err)
		} else {
			profile.Tags = userTags
		}
	}

	return profile, nil
}

// UpdateCustomerProfile อัปเดตโปรไฟล์ลูกค้า
func (s *customerProfileService) UpdateCustomerProfile(businessID, userID, adminID uuid.UUID, updateData types.JSONB) (*models.CustomerProfile, error) {
	// ตรวจสอบสิทธิ์แอดมิน
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(adminID, businessID, []string{"owner", "admin", "operator"})
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errors.New("you don't have permission to update customer profiles")
	}

	// ดึงโปรไฟล์ปัจจุบัน
	profile, err := s.customerProfileRepo.GetByBusinessAndUser(businessID, userID)
	if err != nil {
		return nil, err
	}

	// อัปเดตข้อมูล
	if nickname, ok := updateData["nickname"].(string); ok {
		profile.Nickname = nickname
	}

	if notes, ok := updateData["notes"].(string); ok {
		profile.Notes = notes
	}

	if customerType, ok := updateData["customer_type"].(string); ok && customerType != "" {
		profile.CustomerType = customerType
	}

	if status, ok := updateData["status"].(string); ok && isValidStatus(status) {
		profile.Status = status
	}

	// อัปเดต metadata ถ้ามี
	if metadata, ok := updateData["metadata"].(types.JSONB); ok {
		if profile.Metadata == nil {
			profile.Metadata = make(types.JSONB)
		}
		for key, value := range metadata {
			profile.Metadata[key] = value
		}
	}

	// อัปเดตผู้แก้ไขและเวลา
	profile.UpdatedByID = &adminID
	profile.UpdatedAt = time.Now()

	// บันทึกการเปลี่ยนแปลง
	err = s.customerProfileRepo.Update(profile)
	if err != nil {
		return nil, err
	}

	// ดึง tags ของลูกค้า เพื่อให้ข้อมูลที่ส่งกลับมีความครบถ้วน
	if profile.UserID != uuid.Nil {
		userTags, err := s.userTagRepo.GetUserTags(businessID, profile.UserID)
		if err != nil {
			// Log error แต่ไม่หยุดการทำงาน
			log.Printf("Error fetching tags for user %s: %v", profile.UserID, err)
		} else {
			profile.Tags = userTags
		}
	}

	return profile, nil
}

// GetBusinessCustomers ดึงรายชื่อลูกค้าทั้งหมดของธุรกิจพร้อม tags
func (s *customerProfileService) GetBusinessCustomers(businessID uuid.UUID, limit, offset int) ([]*models.CustomerProfile, int64, error) {
	// ตรวจสอบว่าธุรกิจมีอยู่จริง
	exists, err := s.businessAccountRepo.ExistsById(businessID)
	if err != nil {
		return nil, 0, err
	}
	if !exists {
		return nil, 0, errors.New("business not found")
	}

	// ตั้งค่าเริ่มต้นสำหรับ pagination
	if limit <= 0 || limit > 100 {
		limit = 20 // ค่าเริ่มต้น
	}
	if offset < 0 {
		offset = 0
	}

	// ดึงรายชื่อลูกค้า
	profiles, total, err := s.customerProfileRepo.GetByBusinessID(businessID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูล tags สำหรับลูกค้าแต่ละคน
	for _, profile := range profiles {
		if profile.UserID != uuid.Nil {
			// ดึง tags ของลูกค้า
			userTags, err := s.userTagRepo.GetUserTags(businessID, profile.UserID)
			if err != nil {
				// ทำเครื่องหมายว่ามีข้อผิดพลาดแต่ไม่หยุดการทำงานทั้งหมด
				// อาจจะ log ข้อผิดพลาดไว้
				// log.Printf("Error fetching tags for user %s: %v", profile.UserID, err)
				continue
			}

			// เพิ่ม tags เข้าไปใน profile
			// ต้องแน่ใจว่า CustomerProfile มี field Tags หรือต้องสร้าง struct ใหม่ที่มี field นี้
			profile.Tags = userTags
		}
	}

	return profiles, total, nil
}

// SearchCustomers ค้นหาลูกค้า
func (s *customerProfileService) SearchCustomers(businessID uuid.UUID, query string, limit, offset int) ([]*models.CustomerProfile, int64, error) {
	// ตรวจสอบว่าธุรกิจมีอยู่จริง
	exists, err := s.businessAccountRepo.ExistsById(businessID)
	if err != nil {
		return nil, 0, err
	}
	if !exists {
		return nil, 0, errors.New("business not found")
	}

	// ตรวจสอบคำค้นหา
	if query == "" {
		return s.GetBusinessCustomers(businessID, limit, offset)
	}

	// ตั้งค่าเริ่มต้นสำหรับ pagination
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	// ค้นหาลูกค้า
	profiles, total, err := s.customerProfileRepo.SearchByBusinessID(businessID, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return profiles, total, nil
}

// UpdateLastContact อัปเดตเวลาติดต่อล่าสุด (สำหรับเรียกจาก message service)
func (s *customerProfileService) UpdateLastContact(businessID, userID uuid.UUID) error {
	// ดึงโปรไฟล์ปัจจุบัน (หรือสร้างใหม่ถ้าไม่มี)
	profile, err := s.customerProfileRepo.GetByBusinessAndUser(businessID, userID)
	if err != nil {
		// ถ้าไม่มีโปรไฟล์ ให้สร้างใหม่
		_, err = s.CreateCustomerProfile(businessID, userID, "", "", "New")
		return err
	}

	// อัปเดตเวลาติดต่อล่าสุด
	now := time.Now()
	profile.LastContactAt = &now
	profile.UpdatedAt = now

	return s.customerProfileRepo.Update(profile)
}

// isValidStatus ตรวจสอบความถูกต้องของสถานะ
func isValidStatus(status string) bool {
	validStatuses := []string{"active", "inactive", "blocked", "archived"}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}
