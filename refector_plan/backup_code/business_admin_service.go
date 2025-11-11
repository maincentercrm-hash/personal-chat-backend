// application/serviceimpl/business_admin_service.go
package serviceimpl

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"github.com/thizplus/gofiber-chat-api/domain/service"
)

type businessAdminService struct {
	businessAdminRepo   repository.BusinessAdminRepository
	businessAccountRepo repository.BusinessAccountRepository
	userRepo            repository.UserRepository
}

// NewBusinessAdminService สร้าง instance ใหม่ของ BusinessAdminService
func NewBusinessAdminService(
	businessAdminRepo repository.BusinessAdminRepository,
	businessAccountRepo repository.BusinessAccountRepository,
	userRepo repository.UserRepository,
) service.BusinessAdminService {
	return &businessAdminService{
		businessAdminRepo:   businessAdminRepo,
		businessAccountRepo: businessAccountRepo,
		userRepo:            userRepo,
	}
}

// GetAdmins ดึงรายชื่อแอดมินของธุรกิจ
func (s *businessAdminService) GetAdmins(businessID uuid.UUID, userID uuid.UUID) ([]*models.BusinessAdmin, error) {
	// ตรวจสอบว่าธุรกิจมีอยู่จริง
	exists, err := s.businessAccountRepo.ExistsById(businessID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("business not found")
	}

	// ตรวจสอบสิทธิ์
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(userID, businessID, []string{})
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errors.New("you don't have permission to access this business x2")
	}

	// ดึงรายชื่อแอดมิน
	admins, err := s.businessAdminRepo.GetAdminsByBusinessID(businessID)
	if err != nil {
		return nil, err
	}

	// โหลดข้อมูลผู้ใช้สำหรับแต่ละแอดมิน
	for _, admin := range admins {
		user, err := s.userRepo.FindByID(admin.UserID)
		if err == nil && user != nil {
			admin.User = user
		}
	}

	// เรียงลำดับตามบทบาท (owner ก่อน, ตามด้วย admin, operator)
	var owners, adminUsers, operators, others []*models.BusinessAdmin
	for _, admin := range admins {
		switch admin.Role {
		case "owner":
			owners = append(owners, admin)
		case "admin":
			adminUsers = append(adminUsers, admin)
		case "operator":
			operators = append(operators, admin)
		default:
			others = append(others, admin)
		}
	}

	// รวมผลลัพธ์ตามลำดับ
	result := make([]*models.BusinessAdmin, 0, len(admins))
	result = append(result, owners...)
	result = append(result, adminUsers...)
	result = append(result, operators...)
	result = append(result, others...)

	return result, nil
}

// AddAdmin เพิ่มแอดมินให้ธุรกิจ
func (s *businessAdminService) AddAdmin(businessID uuid.UUID, requestedBy uuid.UUID, newAdminUserID uuid.UUID, role string) (*models.BusinessAdmin, error) {
	// ตรวจสอบว่าธุรกิจมีอยู่จริง
	exists, err := s.businessAccountRepo.ExistsById(businessID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("business not found")
	}

	// ตรวจสอบสิทธิ์ - ต้องเป็น owner เท่านั้น
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(requestedBy, businessID, []string{"owner"})
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errors.New("only the owner can add admins")
	}

	// ตรวจสอบบทบาท
	if role == "" {
		role = "admin" // ค่าเริ่มต้น
	} else if role == "owner" {
		return nil, errors.New("cannot add another owner")
	}

	// ตรวจสอบว่าผู้ใช้มีอยู่จริง
	user, err := s.userRepo.FindByID(newAdminUserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// เช็คว่าเป็นแอดมินอยู่แล้วหรือไม่
	existingAdmin, err := s.businessAdminRepo.GetByUserAndBusinessID(newAdminUserID, businessID)
	if err == nil && existingAdmin != nil {
		// อัพเดทบทบาทเท่านั้น
		existingAdmin.Role = role
		existingAdmin.AddedAt = time.Now()
		existingAdmin.AddedBy = &requestedBy

		err = s.businessAdminRepo.Update(existingAdmin)
		if err != nil {
			return nil, err
		}

		// โหลดข้อมูลผู้ใช้สำหรับส่งกลับ
		existingAdmin.User = user
		return existingAdmin, nil
	}

	// สร้างแอดมินใหม่
	now := time.Now()
	admin := &models.BusinessAdmin{
		ID:         uuid.New(),
		BusinessID: businessID,
		UserID:     newAdminUserID,
		Role:       role,
		AddedAt:    now,
		AddedBy:    &requestedBy,
		User:       user,
	}

	// บันทึกแอดมินใหม่
	err = s.businessAdminRepo.Create(admin)
	if err != nil {
		return nil, err
	}

	return admin, nil
}

// RemoveAdmin ลบแอดมินออกจากธุรกิจ
func (s *businessAdminService) RemoveAdmin(businessID uuid.UUID, requestedBy uuid.UUID, targetUserID uuid.UUID) error {
	// ตรวจสอบว่าธุรกิจมีอยู่จริง
	exists, err := s.businessAccountRepo.ExistsById(businessID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("business not found")
	}

	// ตรวจสอบสิทธิ์ - ต้องเป็น owner เท่านั้น
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(requestedBy, businessID, []string{"owner"})
	if err != nil {
		return err
	}
	if !hasPermission {
		return errors.New("only the owner can remove admins")
	}

	// ตรวจสอบว่าไม่ได้ลบตัวเอง (owner)
	targetAdmin, err := s.businessAdminRepo.GetByUserAndBusinessID(targetUserID, businessID)
	if err != nil {
		return errors.New("admin not found")
	}

	if targetAdmin.Role == "owner" {
		return errors.New("cannot remove the owner")
	}

	// ลบแอดมิน
	return s.businessAdminRepo.DeleteByUserAndBusinessID(targetUserID, businessID)
}

// ChangeAdminRole เปลี่ยนบทบาทของแอดมิน
func (s *businessAdminService) ChangeAdminRole(businessID uuid.UUID, requestedBy uuid.UUID, targetUserID uuid.UUID, newRole string) (*models.BusinessAdmin, error) {
	// ตรวจสอบว่าธุรกิจมีอยู่จริง
	exists, err := s.businessAccountRepo.ExistsById(businessID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("business not found")
	}

	// ตรวจสอบสิทธิ์ - ต้องเป็น owner เท่านั้น
	hasPermission, err := s.businessAdminRepo.CheckAdminPermission(requestedBy, businessID, []string{"owner"})
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errors.New("only the owner can change admin roles")
	}

	// ตรวจสอบบทบาท
	if newRole == "" {
		return nil, errors.New("role is required")
	}

	// ไม่อนุญาตให้เปลี่ยนเป็น owner
	if newRole == "owner" {
		return nil, errors.New("cannot change role to owner")
	}

	// ตรวจสอบว่าไม่ได้เปลี่ยนบทบาทของ owner
	targetAdmin, err := s.businessAdminRepo.GetByUserAndBusinessID(targetUserID, businessID)
	if err != nil {
		return nil, errors.New("admin not found")
	}

	if targetAdmin.Role == "owner" {
		return nil, errors.New("cannot change the role of the owner")
	}

	// เปลี่ยนบทบาท
	targetAdmin.Role = newRole
	err = s.businessAdminRepo.Update(targetAdmin)
	if err != nil {
		return nil, err
	}

	// โหลดข้อมูลผู้ใช้สำหรับส่งกลับ
	user, err := s.userRepo.FindByID(targetUserID)
	if err == nil && user != nil {
		targetAdmin.User = user
	}

	return targetAdmin, nil
}

// CheckAdminPermission ตรวจสอบว่าผู้ใช้เป็นแอดมินของธุรกิจหรือไม่และมีบทบาทตามที่กำหนดหรือไม่
func (s *businessAdminService) CheckAdminPermission(userID uuid.UUID, businessID uuid.UUID, allowedRoles []string) (bool, error) {
	return s.businessAdminRepo.CheckAdminPermission(userID, businessID, allowedRoles)
}

func (s *businessAdminService) GetAdminByUserAndBusinessID(userID, businessID uuid.UUID) (*models.BusinessAdmin, error) {
	return s.businessAdminRepo.GetByUserAndBusinessID(userID, businessID)
}
