// interfaces/api/handler/user_tag_handler.go
package handler

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"

	"github.com/thizplus/gofiber-chat-api/pkg/utils"
)

// UserTagHandler จัดการ HTTP request/response สำหรับ UserTag (Advanced Tag Management)
type UserTagHandler struct {
	userTagService service.UserTagService
}

// NewUserTagHandler สร้าง instance ใหม่ของ UserTagHandler
func NewUserTagHandler(userTagService service.UserTagService) *UserTagHandler {
	return &UserTagHandler{
		userTagService: userTagService,
	}
}

// AddTagToUser เพิ่มแท็กให้กับผู้ใช้
func (h *UserTagHandler) AddTagToUser(c *fiber.Ctx) error {
	// ดึง userID จาก middleware (admin ที่เพิ่มแท็ก)
	adminID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก parameter
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err
	}

	// ดึง tagID จาก parameter
	tagID, err := utils.ParseUUIDParam(c, "tagId")
	if err != nil {
		return err
	}

	// ดึง customerUserID จาก parameter
	customerUserID, err := utils.ParseUUIDParam(c, "userId")
	if err != nil {
		return err
	}

	// เพิ่มแท็กให้ผู้ใช้
	userTag, err := h.userTagService.AddTagToUser(businessID, customerUserID, tagID, adminID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error adding tag to user: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Tag added to user successfully",
		"data":    userTag,
	})
}

// RemoveTagFromUser ลบแท็กออกจากผู้ใช้
func (h *UserTagHandler) RemoveTagFromUser(c *fiber.Ctx) error {
	// ดึง userID จาก middleware (admin ที่ลบแท็ก)
	adminID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก parameter
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err
	}

	// ดึง tagID จาก parameter
	tagID, err := utils.ParseUUIDParam(c, "tagId")
	if err != nil {
		return err
	}

	// ดึง customerUserID จาก parameter
	customerUserID, err := utils.ParseUUIDParam(c, "userId")
	if err != nil {
		return err
	}

	// ลบแท็กจากผู้ใช้
	err = h.userTagService.RemoveTagFromUser(businessID, customerUserID, tagID, adminID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error removing tag from user: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Tag removed from user successfully",
	})
}

// GetUserTags ดึงแท็กทั้งหมดของผู้ใช้ในธุรกิจ
func (h *UserTagHandler) GetUserTags(c *fiber.Ctx) error {
	// ดึง userID จาก middleware (admin ที่ดูแท็ก)
	adminID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก parameter
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err
	}

	// ดึง customerUserID จาก parameter
	customerUserID, err := utils.ParseUUIDParam(c, "userId")
	if err != nil {
		return err
	}

	// ดึงแท็กของผู้ใช้
	userTags, err := h.userTagService.GetUserTags(businessID, customerUserID, adminID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error fetching user tags: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "User tags retrieved successfully",
		"data": fiber.Map{
			"user_tags": userTags,
			"count":     len(userTags),
		},
	})
}

// GetUsersByTag ดึงรายชื่อผู้ใช้ที่มีแท็กนี้
func (h *UserTagHandler) GetUsersByTag(c *fiber.Ctx) error {
	// ดึง userID จาก middleware (admin ที่ดูข้อมูล)
	adminID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก parameter
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err
	}

	// ดึง tagID จาก parameter
	tagID, err := utils.ParseUUIDParam(c, "tagId")
	if err != nil {
		return err
	}

	// ดึงพารามิเตอร์การแบ่งหน้า
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	// ตรวจสอบค่าที่ถูกต้อง
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	// ดึงผู้ใช้ที่มีแท็กนี้
	userTags, total, err := h.userTagService.GetUsersByTag(businessID, tagID, adminID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error fetching users by tag: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Users retrieved successfully",
		"data": fiber.Map{
			"user_tags": userTags,
			"total":     total,
			"limit":     limit,
			"offset":    offset,
		},
	})
}

// CheckUserHasTag ตรวจสอบว่าผู้ใช้มีแท็กนี้หรือไม่
func (h *UserTagHandler) CheckUserHasTag(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	_, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก parameter
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err
	}

	// ดึง tagID จาก parameter
	tagID, err := utils.ParseUUIDParam(c, "tagId")
	if err != nil {
		return err
	}

	// ดึง customerUserID จาก parameter
	customerUserID, err := utils.ParseUUIDParam(c, "userId")
	if err != nil {
		return err
	}

	// ตรวจสอบว่าผู้ใช้มีแท็กนี้หรือไม่
	hasTag, err := h.userTagService.CheckUserHasTag(businessID, customerUserID, tagID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error checking user tag: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "User tag check completed",
		"data": fiber.Map{
			"has_tag":     hasTag,
			"user_id":     customerUserID,
			"tag_id":      tagID,
			"business_id": businessID,
		},
	})
}

// ReplaceUserTags แทนที่แท็กทั้งหมดของผู้ใช้ด้วยแท็กใหม่
func (h *UserTagHandler) ReplaceUserTags(c *fiber.Ctx) error {
	// ดึง userID จาก middleware (admin ที่แทนที่แท็ก)
	adminID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก parameter
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err
	}

	// ดึง customerUserID จาก parameter
	customerUserID, err := utils.ParseUUIDParam(c, "userId")
	if err != nil {
		return err
	}

	// รับข้อมูลแท็กใหม่
	var input struct {
		TagIDs []string `json:"tag_ids" validate:"required"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
	}

	// แปลง string UUIDs เป็น uuid.UUID
	var tagIDs []uuid.UUID
	for _, tagIDStr := range input.TagIDs {
		tagID, err := uuid.Parse(tagIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid tag ID format: " + tagIDStr,
			})
		}
		tagIDs = append(tagIDs, tagID)
	}

	// แทนที่แท็กทั้งหมด
	newUserTags, err := h.userTagService.ReplaceUserTags(businessID, customerUserID, tagIDs, adminID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error replacing user tags: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "User tags replaced successfully",
		"data": fiber.Map{
			"new_user_tags": newUserTags,
			"count":         len(newUserTags),
		},
	})
}

// interfaces/api/handler/user_tag_handler_part2.go
// ต่อจาก Part 1

// BulkAddTagToUsers เพิ่มแท็กให้กับผู้ใช้หลายคน
func (h *UserTagHandler) BulkAddTagToUsers(c *fiber.Ctx) error {
	// ดึง userID จาก middleware (admin ที่เพิ่มแท็ก)
	adminID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก parameter
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err
	}

	// ดึง tagID จาก parameter
	tagID, err := utils.ParseUUIDParam(c, "tagId")
	if err != nil {
		return err
	}

	// รับข้อมูลรายชื่อผู้ใช้
	var input struct {
		UserIDs []string `json:"user_ids" validate:"required,min=1"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
	}

	if len(input.UserIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "User IDs are required",
		})
	}

	// ตรวจสอบจำนวนผู้ใช้ที่อนุญาต
	if len(input.UserIDs) > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Maximum 100 users allowed per request",
		})
	}

	// แปลง string UUIDs เป็น uuid.UUID
	var userIDs []uuid.UUID
	for _, userIDStr := range input.UserIDs {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid user ID format: " + userIDStr,
			})
		}
		userIDs = append(userIDs, userID)
	}

	// เพิ่มแท็กให้ผู้ใช้หลายคน
	addedUserTags, err := h.userTagService.BulkAddTagToUsers(businessID, tagID, userIDs, adminID)
	if err != nil {
		// อาจมี partial success
		return c.Status(fiber.StatusMultiStatus).JSON(fiber.Map{
			"success": false,
			"message": "Some operations failed: " + err.Error(),
			"data": fiber.Map{
				"successful_count": len(addedUserTags),
				"requested_count":  len(userIDs),
				"added_user_tags":  addedUserTags,
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Tag added to users successfully",
		"data": fiber.Map{
			"successful_count": len(addedUserTags),
			"requested_count":  len(userIDs),
			"added_user_tags":  addedUserTags,
		},
	})
}

// BulkRemoveTagFromUsers ลบแท็กออกจากผู้ใช้หลายคน
func (h *UserTagHandler) BulkRemoveTagFromUsers(c *fiber.Ctx) error {
	// ดึง userID จาก middleware (admin ที่ลบแท็ก)
	adminID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก parameter
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err
	}

	// ดึง tagID จาก parameter
	tagID, err := utils.ParseUUIDParam(c, "tagId")
	if err != nil {
		return err
	}

	// รับข้อมูลรายชื่อผู้ใช้
	var input struct {
		UserIDs []string `json:"user_ids" validate:"required,min=1"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
	}

	if len(input.UserIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "User IDs are required",
		})
	}

	// แปลง string UUIDs เป็น uuid.UUID
	var userIDs []uuid.UUID
	for _, userIDStr := range input.UserIDs {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid user ID format: " + userIDStr,
			})
		}
		userIDs = append(userIDs, userID)
	}

	// ลบแท็กจากผู้ใช้หลายคน
	err = h.userTagService.BulkRemoveTagFromUsers(businessID, tagID, userIDs, adminID)
	if err != nil {
		return c.Status(fiber.StatusMultiStatus).JSON(fiber.Map{
			"success": false,
			"message": "Some operations failed: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Tag removed from users successfully",
		"data": fiber.Map{
			"processed_count": len(userIDs),
		},
	})
}

// SearchUsersByTags ค้นหาผู้ใช้ที่มีแท็กตามเงื่อนไข
func (h *UserTagHandler) SearchUsersByTags(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	adminID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก parameter
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err
	}

	// รับเงื่อนไขการค้นหา
	var input struct {
		IncludeTags []string `json:"include_tags"` // แท็กที่ต้องมี
		ExcludeTags []string `json:"exclude_tags"` // แท็กที่ไม่ต้องมี
		MatchType   string   `json:"match_type"`   // "all" หรือ "any"
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
	}

	// ตรวจสอบ match_type
	if input.MatchType == "" {
		input.MatchType = "any" // ค่าเริ่มต้น
	}
	if input.MatchType != "all" && input.MatchType != "any" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "match_type must be 'all' or 'any'",
		})
	}

	// แปลง string UUIDs เป็น uuid.UUID
	var includeTagIDs []uuid.UUID
	for _, tagIDStr := range input.IncludeTags {
		tagID, err := uuid.Parse(tagIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid include tag ID format: " + tagIDStr,
			})
		}
		includeTagIDs = append(includeTagIDs, tagID)
	}

	var excludeTagIDs []uuid.UUID
	for _, tagIDStr := range input.ExcludeTags {
		tagID, err := uuid.Parse(tagIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid exclude tag ID format: " + tagIDStr,
			})
		}
		excludeTagIDs = append(excludeTagIDs, tagID)
	}

	// ดึงพารามิเตอร์การแบ่งหน้า
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	// ตรวจสอบค่าที่ถูกต้อง
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	// สร้างเงื่อนไขการค้นหา
	var matchType service.TagMatchType
	if input.MatchType == "all" {
		matchType = service.TagMatchAll
	} else {
		matchType = service.TagMatchAny
	}

	criteria := service.TagSearchCriteria{
		IncludeTags: includeTagIDs,
		ExcludeTags: excludeTagIDs,
		MatchType:   matchType,
	}

	// ค้นหา
	userTags, total, err := h.userTagService.SearchUsersByTags(businessID, criteria, adminID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error searching users by tags: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Search completed successfully",
		"data": fiber.Map{
			"user_tags": userTags,
			"total":     total,
			"limit":     limit,
			"offset":    offset,
			"criteria": fiber.Map{
				"include_tags": input.IncludeTags,
				"exclude_tags": input.ExcludeTags,
				"match_type":   input.MatchType,
			},
		},
	})
}

// GetUsersWithMultipleTags ดึงผู้ใช้ที่มีแท็กตามที่กำหนด (AND/OR logic)
func (h *UserTagHandler) GetUsersWithMultipleTags(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	adminID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก parameter
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err
	}

	// ดึง tag IDs จาก query parameter
	tagIDsStr := c.Query("tag_ids", "")
	if tagIDsStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "tag_ids parameter is required (comma-separated UUIDs)",
		})
	}

	// แยก tag IDs
	tagIDStrings := strings.Split(tagIDsStr, ",")
	var tagIDs []uuid.UUID
	for _, tagIDStr := range tagIDStrings {
		tagIDStr = strings.TrimSpace(tagIDStr)
		tagID, err := uuid.Parse(tagIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid tag ID format: " + tagIDStr,
			})
		}
		tagIDs = append(tagIDs, tagID)
	}

	// ดึง match type จาก query parameter
	matchTypeStr := c.Query("match_type", "any")
	var matchType service.TagMatchType
	if matchTypeStr == "all" {
		matchType = service.TagMatchAll
	} else if matchTypeStr == "any" {
		matchType = service.TagMatchAny
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "match_type must be 'all' or 'any'",
		})
	}

	// ดึงพารามิเตอร์การแบ่งหน้า
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	// ตรวจสอบค่าที่ถูกต้อง
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	// ค้นหาผู้ใช้
	userTags, total, err := h.userTagService.GetUsersWithMultipleTags(businessID, tagIDs, matchType, adminID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error fetching users with multiple tags: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Users with multiple tags retrieved successfully",
		"data": fiber.Map{
			"user_tags":  userTags,
			"total":      total,
			"limit":      limit,
			"offset":     offset,
			"tag_ids":    tagIDs,
			"match_type": matchTypeStr,
		},
	})
}

// GetTagStatistics ดึงสถิติการใช้แท็ก
func (h *UserTagHandler) GetTagStatistics(c *fiber.Ctx) error {
	// ดึง userID จาก middleware (ต้องเป็น admin)
	adminID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก parameter
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err
	}

	// ดึงสถิติการใช้แท็ก
	statistics, err := h.userTagService.GetTagStatistics(businessID, adminID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error fetching tag statistics: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Tag statistics retrieved successfully",
		"data": fiber.Map{
			"statistics": statistics,
			"count":      len(statistics),
		},
	})
}

// GetUserTagAnalytics ดึงข้อมูลวิเคราะห์ UserTag ขั้นสูง
func (h *UserTagHandler) GetUserTagAnalytics(c *fiber.Ctx) error {
	// ดึง userID จาก middleware (ต้องเป็น owner/admin)
	adminID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก parameter
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err
	}

	// ดึงประเภทการวิเคราะห์จาก query parameter
	analysisType := c.Query("type", "overview")

	switch analysisType {
	case "overview":
		// สถิติภาพรวม
		statistics, err := h.userTagService.GetTagStatistics(businessID, adminID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Error fetching overview analytics: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Overview analytics retrieved successfully",
			"data": fiber.Map{
				"type":       "overview",
				"statistics": statistics,
			},
		})

	case "trends":
		// แนวโน้มการใช้แท็ก (ต้องเพิ่ม method ใน repository)
		days, _ := strconv.Atoi(c.Query("days", "30"))
		if days <= 0 || days > 365 {
			days = 30
		}

		// trends, err := h.userTagService.GetTagTrends(businessID, adminID, days)
		// if err != nil {
		// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		// 		"success": false,
		// 		"message": "Error fetching trend analytics: " + err.Error(),
		// 	})
		// }

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Trend analytics retrieved successfully",
			"data": fiber.Map{
				"type": "trends",
				"days": days,
				// "trends": trends,
			},
		})

	case "distribution":
		// การกระจายแท็กของผู้ใช้
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Distribution analytics retrieved successfully",
			"data": fiber.Map{
				"type": "distribution",
				"note": "Distribution analytics coming soon",
			},
		})

	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid analysis type. Available types: overview, trends, distribution",
		})
	}
}

// ExportUserTags ส่งออกข้อมูล UserTags
func (h *UserTagHandler) ExportUserTags(c *fiber.Ctx) error {
	// ดึง userID จาก middleware (ต้องเป็น admin)
	adminID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก parameter
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err
	}

	// ดึงรูปแบบการส่งออกจาก query parameter
	format := c.Query("format", "json")
	if format != "json" && format != "csv" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid format. Supported formats: json, csv",
		})
	}

	// ดึงฟิลเตอร์เพิ่มเติม
	tagIDStr := c.Query("tag_id", "")
	var tagID *uuid.UUID
	if tagIDStr != "" {
		parsedTagID, err := uuid.Parse(tagIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid tag ID format",
			})
		}
		tagID = &parsedTagID
	}

	// ดึงข้อมูล UserTags ทั้งหมด (หรือตาม tag ที่กำหนด)
	var userTags []*models.UserTag

	if tagID != nil {
		// ดึงผู้ใช้ที่มีแท็กเฉพาะ
		var getUsersErr error
		userTags, _, getUsersErr = h.userTagService.GetUsersByTag(businessID, *tagID, adminID, 1000, 0)
		if getUsersErr != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Error fetching data for export: " + getUsersErr.Error(),
			})
		}
	} else {
		// ดึงสถิติทั้งหมด
		statistics, statErr := h.userTagService.GetTagStatistics(businessID, adminID)
		if statErr != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Error fetching data for export: " + statErr.Error(),
			})
		}

		// สำหรับตัวอย่าง ส่งกลับข้อมูลสถิติ
		if format == "json" {
			return c.JSON(fiber.Map{
				"success": true,
				"message": "Data exported successfully",
				"data": fiber.Map{
					"format":      "json",
					"statistics":  statistics,
					"exported_at": time.Now().Format("2006-01-02 15:04:05"),
				},
			})
		}

		// สำหรับ CSV (simplified)
		c.Set("Content-Type", "text/csv")
		c.Set("Content-Disposition", "attachment; filename=user_tags_export.csv")

		csvData := "tag_name,user_count,created_at\n"
		for _, stat := range statistics {
			csvData += fmt.Sprintf("%s,%d,%s\n", stat.TagName, stat.UserCount, stat.CreatedAt)
		}

		return c.SendString(csvData)
	}

	// ส่งข้อมูลตามรูปแบบที่ต้องการ
	if format == "json" {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Data exported successfully",
			"data": fiber.Map{
				"format":      "json",
				"user_tags":   userTags,
				"count":       len(userTags),
				"exported_at": time.Now().Format("2006-01-02 15:04:05"),
			},
		})
	}

	// CSV format (simplified)
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=user_tags_export.csv")

	csvData := "user_id,tag_name,added_at,added_by\n"
	for _, userTag := range userTags {
		addedBy := ""
		if userTag.AddedBy != nil {
			addedBy = userTag.AddedBy.DisplayName
		}
		csvData += fmt.Sprintf("%s,%s,%s,%s\n",
			userTag.UserID.String(),
			userTag.Tag.Name,
			userTag.AddedAt.Format("2006-01-02 15:04:05"),
			addedBy)
	}

	return c.SendString(csvData)
}

// GetRecentTagActivity ดึงกิจกรรมแท็กล่าสุด
func (h *UserTagHandler) GetRecentTagActivity(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	_, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก parameter
	_, err = utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err
	}

	// ดึงจำนวนกิจกรรมจาก query parameter
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	// ดึงกิจกรรมล่าสุด (ต้องเพิ่ม method ใน service)
	// activities, err := h.userTagService.GetRecentTagActivity(businessID, adminID, limit)
	// if err != nil {
	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
	// 		"success": false,
	// 		"message": "Error fetching recent tag activity: " + err.Error(),
	// 	})
	// }

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Recent tag activity retrieved successfully",
		"data": fiber.Map{
			"limit": limit,
			// "activities": activities,
			"note": "Recent activity feature coming soon",
		},
	})
}

// ValidateTagging ตรวจสอบความถูกต้องของการแท็ก
func (h *UserTagHandler) ValidateTagging(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	_, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก parameter
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err
	}

	// รับข้อมูลการตรวจสอบ
	var input struct {
		UserID string   `json:"user_id" validate:"required"`
		TagIDs []string `json:"tag_ids" validate:"required,min=1"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
	}

	// แปลง UUIDs
	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid user ID format",
		})
	}

	var tagIDs []uuid.UUID
	for _, tagIDStr := range input.TagIDs {
		tagID, err := uuid.Parse(tagIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid tag ID format: " + tagIDStr,
			})
		}
		tagIDs = append(tagIDs, tagID)
	}

	// ตรวจสอบแต่ละแท็ก
	var validationResults []fiber.Map
	for _, tagID := range tagIDs {
		hasTag, err := h.userTagService.CheckUserHasTag(businessID, userID, tagID)
		if err != nil {
			validationResults = append(validationResults, fiber.Map{
				"tag_id": tagID,
				"valid":  false,
				"error":  err.Error(),
			})
		} else {
			validationResults = append(validationResults, fiber.Map{
				"tag_id":  tagID,
				"valid":   true,
				"has_tag": hasTag,
			})
		}
	}

	// นับผลลัพธ์
	validCount := 0
	hasTagCount := 0
	for _, result := range validationResults {
		if result["valid"].(bool) {
			validCount++
			if hasTag, exists := result["has_tag"].(bool); exists && hasTag {
				hasTagCount++
			}
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Tag validation completed",
		"data": fiber.Map{
			"user_id":       userID,
			"business_id":   businessID,
			"total_tags":    len(tagIDs),
			"valid_tags":    validCount,
			"existing_tags": hasTagCount,
			"results":       validationResults,
		},
	})
}
