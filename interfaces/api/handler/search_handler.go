// interfaces/api/handler/search_handler.go
package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/domain/types"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
	"github.com/thizplus/gofiber-chat-api/pkg/utils"
)

// SearchHandler จัดการค้นหาทั้งผู้ใช้และธุรกิจ
type SearchHandler struct {
	userService            service.UserService
	userFriendshipService  service.UserFriendshipService
	businessAccountService service.BusinessAccountService
}

// NewSearchHandler สร้าง instance ใหม่ของ SearchHandler
func NewSearchHandler(
	userService service.UserService,
	userFriendshipService service.UserFriendshipService,
	businessAccountService service.BusinessAccountService,
) *SearchHandler {
	return &SearchHandler{
		userService:            userService,
		userFriendshipService:  userFriendshipService,
		businessAccountService: businessAccountService,
	}
}

// SearchAll ค้นหาทั้งผู้ใช้และธุรกิจ
func (h *SearchHandler) SearchAll(c *fiber.Ctx) error {
	// ดึง userID จาก token
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึงพารามิเตอร์การค้นหา
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Search query is required",
		})
	}

	// ดึงตัวกรองประเภท (ถ้ามี)
	searchType := c.Query("type", "all") // ค่าเริ่มต้นคือ "all"

	// ดึง limit และ offset
	limit := utils.ParseIntWithLimit(c.Query("limit"), 20, 1, 50)
	offset := utils.ParseInt(c.Query("offset"), 0)

	// สร้าง response
	response := fiber.Map{
		"success": true,
	}

	// ค้นหาตามประเภท
	if searchType == "all" || searchType == "user" {
		// ค้นหาผู้ใช้
		users, _, err := h.userService.SearchUsers(query, limit, offset)
		if err == nil {
			// กรองตัวเองออก
			var filteredUsers []types.JSONB
			for _, user := range users {
				if user.ID != userID {
					// ตรวจสอบความสัมพันธ์
					status, friendshipID, _ := h.userFriendshipService.GetFriendshipStatus(userID, user.ID)

					userData := types.JSONB{
						"id":                user.ID.String(),
						"type":              "user",
						"username":          user.Username,
						"display_name":      user.DisplayName,
						"profile_image_url": user.ProfileImageURL,
						"bio":               user.Bio,
						"friendship_status": status,
						"friendship_id":     friendshipID,
					}

					filteredUsers = append(filteredUsers, userData)
				}
			}

			response["users"] = filteredUsers
		}
	}

	if searchType == "all" || searchType == "business" {
		// ค้นหาธุรกิจ
		businesses, _, err := h.businessAccountService.SearchBusinesses(query, limit, offset, userID)
		if err == nil {
			// สร้างข้อมูลสำหรับส่งกลับ
			var businessesData []types.JSONB
			for _, business := range businesses {
				// ตรวจสอบการติดตาม
				isFollowing := false
				if business.Settings != nil {
					if value, ok := business.Settings["is_following"].(bool); ok {
						isFollowing = value
					}
				}

				// ดึงจำนวนผู้ติดตาม
				followerCount := int64(0)
				if business.Settings != nil {
					if value, ok := business.Settings["follower_count"].(int64); ok {
						followerCount = value
					}
				}

				businessData := types.JSONB{
					"id":                business.ID.String(),
					"type":              "business",
					"username":          business.Username,
					"name":              business.Name,
					"display_name":      business.Name, // ใช้ name เป็น display_name เพื่อความสอดคล้อง
					"description":       business.Description,
					"profile_image_url": business.ProfileImageURL,
					"cover_image_url":   business.CoverImageURL,
					"status":            business.Status,
					"is_following":      isFollowing,
					"follower_count":    followerCount,
				}

				businessesData = append(businessesData, businessData)
			}

			response["businesses"] = businessesData
		}
	}

	return c.JSON(response)
}
