// infrastructure/persistence/postgres/user_tag_repository.go
package postgres

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"gorm.io/gorm"
)

type userTagRepository struct {
	db *gorm.DB
}

// NewUserTagRepository สร้าง instance ใหม่ของ UserTagRepository
func NewUserTagRepository(db *gorm.DB) repository.UserTagRepository {
	return &userTagRepository{db: db}
}

// Create สร้าง UserTag ใหม่
func (r *userTagRepository) Create(userTag *models.UserTag) error {
	return r.db.Create(userTag).Error
}

// Delete ลบ UserTag
func (r *userTagRepository) Delete(businessID, userID, tagID uuid.UUID) error {
	return r.db.Where("business_id = ? AND user_id = ? AND tag_id = ?", businessID, userID, tagID).
		Delete(&models.UserTag{}).Error
}

// GetUserTags ดึงแท็กทั้งหมดของผู้ใช้ในธุรกิจ
func (r *userTagRepository) GetUserTags(businessID, userID uuid.UUID) ([]*models.UserTag, error) {
	var userTags []*models.UserTag

	err := r.db.
		Preload("Tag").
		Preload("User").
		Preload("Business").
		Preload("AddedBy").
		Where("business_id = ? AND user_id = ?", businessID, userID).
		Order("added_at DESC").
		Find(&userTags).Error

	if err != nil {
		return nil, err
	}

	return userTags, nil
}

// GetUsersByTag ดึงรายชื่อผู้ใช้ที่มีแท็กนี้
func (r *userTagRepository) GetUsersByTag(businessID, tagID uuid.UUID) ([]*models.UserTag, error) {
	var userTags []*models.UserTag

	err := r.db.
		Preload("Tag").
		Preload("User").
		Preload("Business").
		Preload("AddedBy").
		Where("business_id = ? AND tag_id = ?", businessID, tagID).
		Order("added_at DESC").
		Find(&userTags).Error

	if err != nil {
		return nil, err
	}

	return userTags, nil
}

// GetByID ดึง UserTag ตาม ID
func (r *userTagRepository) GetByID(id uuid.UUID) (*models.UserTag, error) {
	var userTag models.UserTag

	err := r.db.
		Preload("Tag").
		Preload("User").
		Preload("Business").
		Preload("AddedBy").
		Where("id = ?", id).
		First(&userTag).Error

	if err != nil {
		return nil, err
	}

	return &userTag, nil
}

// Exists ตรวจสอบว่าผู้ใช้มีแท็กนี้หรือไม่
func (r *userTagRepository) Exists(businessID, userID, tagID uuid.UUID) (bool, error) {
	var count int64

	err := r.db.Model(&models.UserTag{}).
		Where("business_id = ? AND user_id = ? AND tag_id = ?", businessID, userID, tagID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetByBusinessID ดึง UserTags ทั้งหมดของธุรกิจ
func (r *userTagRepository) GetByBusinessID(businessID uuid.UUID, limit, offset int) ([]*models.UserTag, int64, error) {
	var userTags []*models.UserTag
	var total int64

	// นับจำนวนทั้งหมด
	err := r.db.Model(&models.UserTag{}).
		Where("business_id = ?", businessID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลแบบ pagination
	err = r.db.
		Preload("Tag").
		Preload("User").
		Preload("AddedBy").
		Where("business_id = ?", businessID).
		Order("added_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&userTags).Error

	if err != nil {
		return nil, 0, err
	}

	return userTags, total, nil
}

// DeleteByUserID ลบแท็กทั้งหมดของผู้ใช้ในธุรกิจ
func (r *userTagRepository) DeleteByUserID(businessID, userID uuid.UUID) error {
	return r.db.Where("business_id = ? AND user_id = ?", businessID, userID).
		Delete(&models.UserTag{}).Error
}

// DeleteByTagID ลบ UserTags ทั้งหมดที่ใช้แท็กนี้
func (r *userTagRepository) DeleteByTagID(tagID uuid.UUID) error {
	return r.db.Where("tag_id = ?", tagID).Delete(&models.UserTag{}).Error
}

// BulkCreate สร้าง UserTags หลายตัวพร้อมกัน
func (r *userTagRepository) BulkCreate(userTags []*models.UserTag) error {
	if len(userTags) == 0 {
		return nil
	}

	return r.db.Create(&userTags).Error
}

// BulkDelete ลบ UserTags หลายตัวพร้อมกัน
func (r *userTagRepository) BulkDelete(businessID uuid.UUID, userIDs []uuid.UUID, tagID uuid.UUID) error {
	if len(userIDs) == 0 {
		return nil
	}

	return r.db.Where("business_id = ? AND user_id IN ? AND tag_id = ?", businessID, userIDs, tagID).
		Delete(&models.UserTag{}).Error
}

// BulkDeleteByTagIDs ลบ UserTags ตามแท็กหลายตัว
func (r *userTagRepository) BulkDeleteByTagIDs(businessID uuid.UUID, tagIDs []uuid.UUID) error {
	if len(tagIDs) == 0 {
		return nil
	}

	return r.db.Where("business_id = ? AND tag_id IN ?", businessID, tagIDs).
		Delete(&models.UserTag{}).Error
}

// GetUsersByMultipleTags ดึงผู้ใช้ที่มีแท็กหลายตัว (AND logic)
func (r *userTagRepository) GetUsersByMultipleTags(businessID uuid.UUID, tagIDs []uuid.UUID) ([]*models.UserTag, error) {
	var userTags []*models.UserTag

	if len(tagIDs) == 0 {
		return userTags, nil
	}

	// ใช้ subquery เพื่อหาผู้ใช้ที่มีทุกแท็กที่กำหนด
	subQuery := r.db.Table("user_tags").
		Select("user_id").
		Where("business_id = ? AND tag_id IN ?", businessID, tagIDs).
		Group("user_id").
		Having("COUNT(DISTINCT tag_id) = ?", len(tagIDs))

	err := r.db.
		Preload("Tag").
		Preload("User").
		Preload("AddedBy").
		Where("business_id = ? AND user_id IN (?)", businessID, subQuery).
		Find(&userTags).Error

	if err != nil {
		return nil, err
	}

	return userTags, nil
}

// GetUsersByAnyTags ดึงผู้ใช้ที่มีแท็กใดแท็กหนึ่ง (OR logic)
func (r *userTagRepository) GetUsersByAnyTags(businessID uuid.UUID, tagIDs []uuid.UUID) ([]*models.UserTag, error) {
	var userTags []*models.UserTag

	if len(tagIDs) == 0 {
		return userTags, nil
	}

	err := r.db.
		Preload("Tag").
		Preload("User").
		Preload("AddedBy").
		Where("business_id = ? AND tag_id IN ?", businessID, tagIDs).
		Find(&userTags).Error

	if err != nil {
		return nil, err
	}

	return userTags, nil
}

// GetUsersExcludingTags ดึงผู้ใช้ที่ไม่มีแท็กที่กำหนด
func (r *userTagRepository) GetUsersExcludingTags(businessID uuid.UUID, excludeTagIDs []uuid.UUID, limit, offset int) ([]*models.UserTag, int64, error) {
	var userTags []*models.UserTag
	var total int64

	if len(excludeTagIDs) == 0 {
		// ถ้าไม่มีแท็กที่ต้องการ exclude ให้ดึงทั้งหมด
		return r.GetByBusinessID(businessID, limit, offset)
	}

	// ดึงผู้ใช้ที่ไม่มีแท็กที่ต้องการ exclude
	excludeUserIDs := r.db.Table("user_tags").
		Select("DISTINCT user_id").
		Where("business_id = ? AND tag_id IN ?", businessID, excludeTagIDs)

	baseQuery := r.db.Model(&models.UserTag{}).
		Where("business_id = ?", businessID).
		Where("user_id NOT IN (?)", excludeUserIDs)

	// นับจำนวนทั้งหมด
	err := baseQuery.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลแบบ pagination
	err = r.db.
		Preload("Tag").
		Preload("User").
		Preload("AddedBy").
		Where("business_id = ?", businessID).
		Where("user_id NOT IN (?)", excludeUserIDs).
		Order("added_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&userTags).Error

	if err != nil {
		return nil, 0, err
	}

	return userTags, total, nil
}

// SearchUserTags ค้นหา UserTags ตามเงื่อนไขต่างๆ
func (r *userTagRepository) SearchUserTags(businessID uuid.UUID, query string, limit, offset int) ([]*models.UserTag, int64, error) {
	var userTags []*models.UserTag
	var total int64

	// Base query
	baseQuery := r.db.Model(&models.UserTag{}).
		Joins("LEFT JOIN tags ON user_tags.tag_id = tags.id").
		Joins("LEFT JOIN users ON user_tags.user_id = users.id").
		Where("user_tags.business_id = ?", businessID)

	// เพิ่มเงื่อนไขการค้นหา
	if query != "" {
		searchCondition := "tags.name ILIKE ? OR users.display_name ILIKE ? OR users.username ILIKE ?"
		baseQuery = baseQuery.Where(searchCondition, "%"+query+"%", "%"+query+"%", "%"+query+"%")
	}

	// นับจำนวนทั้งหมด
	err := baseQuery.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลแบบ pagination
	searchQuery := r.db.
		Preload("Tag").
		Preload("User").
		Preload("AddedBy").
		Joins("LEFT JOIN tags ON user_tags.tag_id = tags.id").
		Joins("LEFT JOIN users ON user_tags.user_id = users.id").
		Where("user_tags.business_id = ?", businessID)

	if query != "" {
		searchCondition := "tags.name ILIKE ? OR users.display_name ILIKE ? OR users.username ILIKE ?"
		searchQuery = searchQuery.Where(searchCondition, "%"+query+"%", "%"+query+"%", "%"+query+"%")
	}

	err = searchQuery.
		Order("user_tags.added_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&userTags).Error

	if err != nil {
		return nil, 0, err
	}

	return userTags, total, nil
}

// GetUserTagsByDateRange ดึง UserTags ตามช่วงวันที่
func (r *userTagRepository) GetUserTagsByDateRange(businessID uuid.UUID, startDate, endDate string, limit, offset int) ([]*models.UserTag, int64, error) {
	var userTags []*models.UserTag
	var total int64

	baseQuery := r.db.Model(&models.UserTag{}).
		Where("business_id = ?", businessID).
		Where("added_at >= ? AND added_at <= ?", startDate, endDate)

	// นับจำนวนทั้งหมด
	err := baseQuery.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลแบบ pagination
	err = r.db.
		Preload("Tag").
		Preload("User").
		Preload("AddedBy").
		Where("business_id = ?", businessID).
		Where("added_at >= ? AND added_at <= ?", startDate, endDate).
		Order("added_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&userTags).Error

	if err != nil {
		return nil, 0, err
	}

	return userTags, total, nil
}

// GetRecentUserTags ดึง UserTags ล่าสุด
func (r *userTagRepository) GetRecentUserTags(businessID uuid.UUID, days int, limit int) ([]*models.UserTag, error) {
	var userTags []*models.UserTag

	if limit <= 0 {
		limit = 50 // ค่าเริ่มต้น
	}

	err := r.db.
		Preload("Tag").
		Preload("User").
		Preload("AddedBy").
		Where("business_id = ?", businessID).
		Where("added_at >= NOW() - INTERVAL '? days'", days).
		Order("added_at DESC").
		Limit(limit).
		Find(&userTags).Error

	if err != nil {
		return nil, err
	}

	return userTags, nil
}

// CountUsersByTag นับจำนวนผู้ใช้ที่มีแท็กแต่ละตัว
func (r *userTagRepository) CountUsersByTag(businessID uuid.UUID, tagIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	type TagCount struct {
		TagID uuid.UUID `json:"tag_id"`
		Count int64     `json:"count"`
	}

	var results []TagCount
	result := make(map[uuid.UUID]int64)

	if len(tagIDs) == 0 {
		return result, nil
	}

	err := r.db.Table("user_tags").
		Select("tag_id, COUNT(*) as count").
		Where("business_id = ? AND tag_id IN ?", businessID, tagIDs).
		Group("tag_id").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// แปลงเป็น map
	for _, tagCount := range results {
		result[tagCount.TagID] = tagCount.Count
	}

	return result, nil
}

// infrastructure/persistence/postgres/user_tag_repository_part3.go
// ต่อจาก Part 2

// GetUserTagStatistics ดึงสถิติ UserTags ของธุรกิจ
func (r *userTagRepository) GetUserTagStatistics(businessID uuid.UUID) (*UserTagStatistics, error) {
	var stats UserTagStatistics

	// นับจำนวน UserTags ทั้งหมด
	err := r.db.Model(&models.UserTag{}).
		Where("business_id = ?", businessID).
		Count(&stats.TotalUserTags).Error
	if err != nil {
		return nil, err
	}

	// นับจำนวนผู้ใช้ที่มีแท็ก (ไม่ซ้ำ)
	err = r.db.Model(&models.UserTag{}).
		Where("business_id = ?", businessID).
		Distinct("user_id").
		Count(&stats.UsersWithTags).Error
	if err != nil {
		return nil, err
	}

	// นับจำนวนแท็กที่ถูกใช้ (ไม่ซ้ำ)
	err = r.db.Model(&models.UserTag{}).
		Where("business_id = ?", businessID).
		Distinct("tag_id").
		Count(&stats.ActiveTags).Error
	if err != nil {
		return nil, err
	}

	// หาค่าเฉลี่ยแท็กต่อผู้ใช้
	if stats.UsersWithTags > 0 {
		stats.AvgTagsPerUser = float64(stats.TotalUserTags) / float64(stats.UsersWithTags)
	}

	return &stats, nil
}

// GetMostPopularTags ดึงแท็กที่ได้รับความนิยมมากที่สุด
func (r *userTagRepository) GetMostPopularTags(businessID uuid.UUID, limit int) ([]TagPopularity, error) {
	var results []TagPopularity

	if limit <= 0 {
		limit = 10
	}

	err := r.db.Table("user_tags").
		Select("user_tags.tag_id, tags.name as tag_name, tags.color as tag_color, COUNT(*) as user_count").
		Joins("LEFT JOIN tags ON user_tags.tag_id = tags.id").
		Where("user_tags.business_id = ?", businessID).
		Group("user_tags.tag_id, tags.name, tags.color").
		Order("user_count DESC").
		Limit(limit).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetUserTagTrends ดึงแนวโน้มการใช้แท็กตามเวลา
func (r *userTagRepository) GetUserTagTrends(businessID uuid.UUID, days int) ([]TagTrend, error) {
	var results []TagTrend

	if days <= 0 {
		days = 30 // ค่าเริ่มต้น 30 วัน
	}

	err := r.db.Table("user_tags").
		Select("DATE(added_at) as date, COUNT(*) as count").
		Where("business_id = ? AND added_at >= NOW() - INTERVAL '? days'", businessID, days).
		Group("DATE(added_at)").
		Order("date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetUserTagDistribution ดึงการกระจายแท็กของผู้ใช้
func (r *userTagRepository) GetUserTagDistribution(businessID uuid.UUID) ([]UserTagDistribution, error) {
	var results []UserTagDistribution

	err := r.db.Raw(`
		SELECT 
			tag_count,
			COUNT(*) as user_count
		FROM (
			SELECT 
				user_id,
				COUNT(*) as tag_count
			FROM user_tags 
			WHERE business_id = ?
			GROUP BY user_id
		) as user_tag_counts
		GROUP BY tag_count
		ORDER BY tag_count ASC
	`, businessID).Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetTopTaggedUsers ดึงผู้ใช้ที่มีแท็กมากที่สุด
func (r *userTagRepository) GetTopTaggedUsers(businessID uuid.UUID, limit int) ([]UserTagCount, error) {
	var results []UserTagCount

	if limit <= 0 {
		limit = 10
	}

	err := r.db.Table("user_tags").
		Select("user_tags.user_id, users.display_name, users.username, COUNT(*) as tag_count").
		Joins("LEFT JOIN users ON user_tags.user_id = users.id").
		Where("user_tags.business_id = ?", businessID).
		Group("user_tags.user_id, users.display_name, users.username").
		Order("tag_count DESC").
		Limit(limit).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetTagUsageByAdmin ดึงสถิติการใช้แท็กตามผู้เพิ่ม
func (r *userTagRepository) GetTagUsageByAdmin(businessID uuid.UUID) ([]AdminTagUsage, error) {
	var results []AdminTagUsage

	err := r.db.Table("user_tags").
		Select("user_tags.added_by_id, users.display_name as admin_name, COUNT(*) as tags_added").
		Joins("LEFT JOIN users ON user_tags.added_by_id = users.id").
		Where("user_tags.business_id = ? AND user_tags.added_by_id IS NOT NULL", businessID).
		Group("user_tags.added_by_id, users.display_name").
		Order("tags_added DESC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetUntaggedUsers ดึงผู้ใช้ที่ยังไม่มีแท็ก
func (r *userTagRepository) GetUntaggedUsers(businessID uuid.UUID, limit, offset int) ([]UntaggedUser, int64, error) {
	var results []UntaggedUser
	var total int64

	// Base query สำหรับนับ
	baseQuery := r.db.Table("customer_profiles").
		Select("customer_profiles.user_id").
		Joins("LEFT JOIN user_tags ON customer_profiles.user_id = user_tags.user_id AND customer_profiles.business_id = user_tags.business_id").
		Where("customer_profiles.business_id = ? AND user_tags.user_id IS NULL", businessID)

	// นับจำนวนทั้งหมด
	err := baseQuery.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// ดึงข้อมูลแบบ pagination
	err = r.db.Table("customer_profiles").
		Select("customer_profiles.user_id, users.display_name, users.username, customer_profiles.created_at as joined_at").
		Joins("LEFT JOIN users ON customer_profiles.user_id = users.id").
		Joins("LEFT JOIN user_tags ON customer_profiles.user_id = user_tags.user_id AND customer_profiles.business_id = user_tags.business_id").
		Where("customer_profiles.business_id = ? AND user_tags.user_id IS NULL", businessID).
		Order("customer_profiles.created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(&results).Error

	if err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// GetTagCombinations ดึงการจับคู่แท็กที่เกิดขึ้นบ่อย
func (r *userTagRepository) GetTagCombinations(businessID uuid.UUID, limit int) ([]TagCombination, error) {
	var results []TagCombination

	if limit <= 0 {
		limit = 20
	}

	err := r.db.Raw(`
		SELECT 
			t1.name as tag1_name,
			t2.name as tag2_name,
			COUNT(*) as combination_count
		FROM user_tags ut1
		JOIN user_tags ut2 ON ut1.user_id = ut2.user_id 
			AND ut1.business_id = ut2.business_id 
			AND ut1.tag_id < ut2.tag_id
		JOIN tags t1 ON ut1.tag_id = t1.id
		JOIN tags t2 ON ut2.tag_id = t2.id
		WHERE ut1.business_id = ?
		GROUP BY t1.name, t2.name
		HAVING COUNT(*) > 1
		ORDER BY combination_count DESC
		LIMIT ?
	`, businessID, limit).Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetRecentTagActivity ดึงกิจกรรมแท็กล่าสุด
func (r *userTagRepository) GetRecentTagActivity(businessID uuid.UUID, limit int) ([]*models.UserTag, error) {
	var userTags []*models.UserTag

	if limit <= 0 {
		limit = 50
	}

	err := r.db.
		Preload("Tag").
		Preload("User").
		Preload("AddedBy").
		Where("business_id = ?", businessID).
		Order("added_at DESC").
		Limit(limit).
		Find(&userTags).Error

	if err != nil {
		return nil, err
	}

	return userTags, nil
}

// Custom types สำหรับ Analytics

// UserTagStatistics สถิติโดยรวมของ UserTags
type UserTagStatistics struct {
	TotalUserTags  int64   `json:"total_user_tags"`
	UsersWithTags  int64   `json:"users_with_tags"`
	ActiveTags     int64   `json:"active_tags"`
	AvgTagsPerUser float64 `json:"avg_tags_per_user"`
}

// TagPopularity ความนิยมของแท็ก
type TagPopularity struct {
	TagID     uuid.UUID `json:"tag_id"`
	TagName   string    `json:"tag_name"`
	TagColor  string    `json:"tag_color"`
	UserCount int64     `json:"user_count"`
}

// TagTrend แนวโน้มการใช้แท็กตามวันที่
type TagTrend struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// UserTagDistribution การกระจายจำนวนแท็กของผู้ใช้
type UserTagDistribution struct {
	TagCount  int64 `json:"tag_count"`  // จำนวนแท็กที่ผู้ใช้มี
	UserCount int64 `json:"user_count"` // จำนวนผู้ใช้ที่มีแท็กเท่านี้
}

// UserTagCount จำนวนแท็กของผู้ใช้
type UserTagCount struct {
	UserID      uuid.UUID `json:"user_id"`
	DisplayName string    `json:"display_name"`
	Username    string    `json:"username"`
	TagCount    int64     `json:"tag_count"`
}

// AdminTagUsage การใช้แท็กตาม Admin
type AdminTagUsage struct {
	AddedByID *uuid.UUID `json:"added_by_id"`
	AdminName string     `json:"admin_name"`
	TagsAdded int64      `json:"tags_added"`
}

// UntaggedUser ผู้ใช้ที่ยังไม่มีแท็ก
type UntaggedUser struct {
	UserID      uuid.UUID `json:"user_id"`
	DisplayName string    `json:"display_name"`
	Username    string    `json:"username"`
	JoinedAt    string    `json:"joined_at"`
}

// TagCombination การจับคู่แท็กที่พบบ่อย
type TagCombination struct {
	Tag1Name         string `json:"tag1_name"`
	Tag2Name         string `json:"tag2_name"`
	CombinationCount int64  `json:"combination_count"`
}
