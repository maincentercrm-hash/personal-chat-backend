// infrastructure/persistence/postgres/customer_profile_repository.go
package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"gorm.io/gorm"
)

type customerProfileRepository struct {
	db *gorm.DB
}

// NewCustomerProfileRepository สร้าง instance ใหม่ของ CustomerProfileRepository
func NewCustomerProfileRepository(db *gorm.DB) repository.CustomerProfileRepository {
	return &customerProfileRepository{db: db}
}

// Create สร้าง CustomerProfile ใหม่
func (r *customerProfileRepository) Create(profile *models.CustomerProfile) error {
	return r.db.Create(profile).Error
}

// GetByBusinessAndUser ดึง CustomerProfile ตาม BusinessID และ UserID
func (r *customerProfileRepository) GetByBusinessAndUser(businessID, userID uuid.UUID) (*models.CustomerProfile, error) {
	var profile models.CustomerProfile

	err := r.db.
		// ลบ Preload("Business") ออก
		Preload("User").
		Preload("CreatedBy").
		Preload("UpdatedBy").
		Where("business_id = ? AND user_id = ?", businessID, userID).
		First(&profile).Error

	if err != nil {
		return nil, err
	}

	return &profile, nil
}

// Update อัปเดต CustomerProfile
func (r *customerProfileRepository) Update(profile *models.CustomerProfile) error {
	return r.db.Save(profile).Error
}

// GetByBusinessID ดึงรายชื่อ CustomerProfiles ทั้งหมดของธุรกิจ
func (r *customerProfileRepository) GetByBusinessID(businessID uuid.UUID, limit, offset int) ([]*models.CustomerProfile, int64, error) {
	var profiles []*models.CustomerProfile
	var total int64

	// นับจำนวนทั้งหมด
	err := r.db.Model(&models.CustomerProfile{}).
		Where("business_id = ?", businessID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลแบบ pagination
	err = r.db.
		Preload("User").
		Preload("CreatedBy").
		Preload("UpdatedBy").
		Where("business_id = ?", businessID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&profiles).Error

	if err != nil {
		return nil, 0, err
	}

	return profiles, total, nil
}

// SearchByBusinessID ค้นหา CustomerProfiles ในธุรกิจ
func (r *customerProfileRepository) SearchByBusinessID(businessID uuid.UUID, query string, limit, offset int) ([]*models.CustomerProfile, int64, error) {
	var profiles []*models.CustomerProfile
	var total int64

	// สร้าง base query
	baseQuery := r.db.Model(&models.CustomerProfile{}).
		Joins("LEFT JOIN users ON customer_profiles.user_id = users.id").
		Where("customer_profiles.business_id = ?", businessID)

	// เพิ่มเงื่อนไขการค้นหา
	searchQuery := baseQuery.Where(
		"customer_profiles.nickname ILIKE ? OR "+
			"customer_profiles.notes ILIKE ? OR "+
			"customer_profiles.customer_type ILIKE ? OR "+
			"users.display_name ILIKE ? OR "+
			"users.username ILIKE ?",
		"%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%",
	)

	// นับจำนวนทั้งหมด
	err := searchQuery.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลแบบ pagination
	err = r.db.
		Preload("User").
		Preload("CreatedBy").
		Preload("UpdatedBy").
		Joins("LEFT JOIN users ON customer_profiles.user_id = users.id").
		Where("customer_profiles.business_id = ?", businessID).
		Where(
			"customer_profiles.nickname ILIKE ? OR "+
				"customer_profiles.notes ILIKE ? OR "+
				"customer_profiles.customer_type ILIKE ? OR "+
				"users.display_name ILIKE ? OR "+
				"users.username ILIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%", "%"+query+"%",
		).
		Order("customer_profiles.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&profiles).Error

	if err != nil {
		return nil, 0, err
	}

	return profiles, total, nil
}

// GetByID ดึง CustomerProfile ตาม ID
func (r *customerProfileRepository) GetByID(id uuid.UUID) (*models.CustomerProfile, error) {
	var profile models.CustomerProfile

	err := r.db.
		Preload("Business").
		Preload("User").
		Preload("CreatedBy").
		Preload("UpdatedBy").
		Where("id = ?", id).
		First(&profile).Error

	if err != nil {
		return nil, err
	}

	return &profile, nil
}

// Delete ลบ CustomerProfile (soft delete)
func (r *customerProfileRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.CustomerProfile{}).Error
}

// GetByUserID ดึง CustomerProfiles ทั้งหมดของผู้ใช้ (ในธุรกิจต่างๆ)
func (r *customerProfileRepository) GetByUserID(userID uuid.UUID) ([]*models.CustomerProfile, error) {
	var profiles []*models.CustomerProfile

	err := r.db.
		Preload("Business").
		Preload("User").
		Where("user_id = ?", userID).
		Find(&profiles).Error

	if err != nil {
		return nil, err
	}

	return profiles, nil
}

// GetByCustomerType ดึง CustomerProfiles ตามประเภทลูกค้า
func (r *customerProfileRepository) GetByCustomerType(businessID uuid.UUID, customerType string, limit, offset int) ([]*models.CustomerProfile, int64, error) {
	var profiles []*models.CustomerProfile
	var total int64

	// นับจำนวนทั้งหมด
	err := r.db.Model(&models.CustomerProfile{}).
		Where("business_id = ? AND customer_type = ?", businessID, customerType).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลแบบ pagination
	err = r.db.
		Preload("User").
		Where("business_id = ? AND customer_type = ?", businessID, customerType).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&profiles).Error

	if err != nil {
		return nil, 0, err
	}

	return profiles, total, nil
}

// GetByStatus ดึง CustomerProfiles ตามสถานะ
func (r *customerProfileRepository) GetByStatus(businessID uuid.UUID, status string, limit, offset int) ([]*models.CustomerProfile, int64, error) {
	var profiles []*models.CustomerProfile
	var total int64

	// นับจำนวนทั้งหมด
	err := r.db.Model(&models.CustomerProfile{}).
		Where("business_id = ? AND status = ?", businessID, status).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลแบบ pagination
	err = r.db.
		Preload("User").
		Where("business_id = ? AND status = ?", businessID, status).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&profiles).Error

	if err != nil {
		return nil, 0, err
	}

	return profiles, total, nil
}

// GetInactiveCustomers ดึงลูกค้าที่ไม่ได้ติดต่อมานาน
func (r *customerProfileRepository) GetInactiveCustomers(businessID uuid.UUID, days int, limit, offset int) ([]*models.CustomerProfile, int64, error) {
	var profiles []*models.CustomerProfile
	var total int64

	// สร้าง query สำหรับลูกค้าที่ไม่ได้ติดต่อมานาน
	inactiveQuery := r.db.Model(&models.CustomerProfile{}).
		Where("business_id = ?", businessID).
		Where("last_contact_at IS NULL OR last_contact_at < NOW() - INTERVAL '? days'", days)

	// นับจำนวนทั้งหมด
	err := inactiveQuery.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลแบบ pagination
	err = r.db.
		Preload("User").
		Where("business_id = ?", businessID).
		Where("last_contact_at IS NULL OR last_contact_at < NOW() - INTERVAL '? days'", days).
		Order("last_contact_at ASC NULLS FIRST").
		Limit(limit).
		Offset(offset).
		Find(&profiles).Error

	if err != nil {
		return nil, 0, err
	}

	return profiles, total, nil
}

// UpdateLastContact อัปเดตเวลาติดต่อล่าสุด
func (r *customerProfileRepository) UpdateLastContact(businessID, userID uuid.UUID) error {
	return r.db.Model(&models.CustomerProfile{}).
		Where("business_id = ? AND user_id = ?", businessID, userID).
		Update("last_contact_at", "NOW()").
		Update("updated_at", "NOW()").Error
}

// FindByConditions ค้นหา customer profiles ตามเงื่อนไข
func (r *customerProfileRepository) FindByConditions(
	businessID uuid.UUID,
	customerTypes []string,
	lastContactFrom, lastContactTo *time.Time,
	statuses []string,
	customQuery map[string]interface{},
) ([]*models.CustomerProfile, error) {
	var profiles []*models.CustomerProfile

	query := r.db.Where("business_id = ?", businessID)

	// กรองตามประเภทลูกค้า
	if len(customerTypes) > 0 {
		query = query.Where("customer_type IN ?", customerTypes)
	}

	// กรองตามเวลาติดต่อล่าสุด
	if lastContactFrom != nil {
		query = query.Where("last_contact_at >= ?", lastContactFrom)
	}
	if lastContactTo != nil {
		query = query.Where("last_contact_at <= ?", lastContactTo)
	}

	// กรองตามสถานะ
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}

	// กรองตาม custom query
	if customQuery != nil {
		// ตรวจสอบ metadata field
		if metadata, ok := customQuery["metadata"]; ok {
			if metadataMap, ok := metadata.(map[string]interface{}); ok {
				for key, value := range metadataMap {
					// ใช้ JSONB query ในการค้นหาข้อมูลใน metadata field
					query = query.Where("metadata->>'"+key+"' = ?", value)
				}
			}
		}
	}

	if err := query.Find(&profiles).Error; err != nil {
		return nil, err
	}

	return profiles, nil
}
