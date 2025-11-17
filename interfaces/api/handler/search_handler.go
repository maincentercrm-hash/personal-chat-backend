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
}

// NewSearchHandler สร้าง instance ใหม่ของ SearchHandler
func NewSearchHandler(
	userService service.UserService,
	userFriendshipService service.UserFriendshipService,
) *SearchHandler {
	return &SearchHandler{
		userService:            userService,
		userFriendshipService:  userFriendshipService,
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


	return c.JSON(response)
}
