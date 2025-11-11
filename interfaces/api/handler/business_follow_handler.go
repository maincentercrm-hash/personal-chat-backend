// interfaces/api/handler/business_follow_handler.go
package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
	"github.com/thizplus/gofiber-chat-api/pkg/utils"
)

// BusinessFollowHandler จัดการ HTTP request/response สำหรับการติดตามธุรกิจ
type BusinessFollowHandler struct {
	businessFollowService service.BusinessFollowService
}

// NewBusinessFollowHandler สร้าง instance ใหม่ของ BusinessFollowHandler
func NewBusinessFollowHandler(businessFollowService service.BusinessFollowService) *BusinessFollowHandler {
	return &BusinessFollowHandler{
		businessFollowService: businessFollowService,
	}
}

// FollowBusiness - ผู้ใช้ติดตามธุรกิจ
func (h *BusinessFollowHandler) FollowBusiness(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จากพารามิเตอร์
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// รับข้อมูลเพิ่มเติม
	var input struct {
		Source string `json:"source"`
	}
	c.BodyParser(&input) // ไม่จำเป็นต้องตรวจสอบข้อผิดพลาด

	// ติดตามธุรกิจ
	err = h.businessFollowService.FollowBusiness(userID, businessID, input.Source)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error following business: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Business followed successfully",
	})
}

// UnfollowBusiness - ผู้ใช้เลิกติดตามธุรกิจ
func (h *BusinessFollowHandler) UnfollowBusiness(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จากพารามิเตอร์
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// เลิกติดตามธุรกิจ
	err = h.businessFollowService.UnfollowBusiness(userID, businessID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error unfollowing business: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Business unfollowed successfully",
	})
}

// GetBusinessFollowers - ดึงรายชื่อผู้ติดตามของธุรกิจ
func (h *BusinessFollowHandler) GetBusinessFollowers(c *fiber.Ctx) error {
	// ดึง userID จาก middleware เพื่อตรวจสอบการเข้าถึง
	_, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จากพารามิเตอร์
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// ดึง limit และ offset จาก query params
	limit := utils.ParseIntWithLimit(c.Query("limit"), 20, 1, 100)
	offset := utils.ParseInt(c.Query("offset"), 0)

	// ดึงข้อมูลผู้ติดตาม
	followers, total, err := h.businessFollowService.GetBusinessFollowers(businessID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error getting followers: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"followers": followers,
			"total":     total,
			"limit":     limit,
			"offset":    offset,
		},
	})
}

// GetUserFollowedBusinesses - ดึงรายชื่อธุรกิจที่ผู้ใช้ติดตาม
func (h *BusinessFollowHandler) GetUserFollowedBusinesses(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	fmt.Printf("DEBUG: userID=%s, err=%v\n", userID, err) // เพิ่มบรรทัดนี้เพื่อ debug

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ตรวจสอบว่าเป็นการดูข้อมูลของตัวเองหรือของผู้อื่น
	var targetUserID uuid.UUID

	if targetUserIDStr := c.Params("userId"); targetUserIDStr != "" {
		targetUserID, err = utils.ParseUUIDOrError(c, targetUserIDStr, "Invalid target user ID format")
		if err != nil {
			return err
		}
	} else {
		// ถ้าไม่ระบุ userId ให้ใช้ของผู้ใช้ปัจจุบัน
		targetUserID = userID
	}

	// ดึง limit และ offset จาก query params
	limit := utils.ParseIntWithLimit(c.Query("limit"), 20, 1, 100)
	offset := utils.ParseInt(c.Query("offset"), 0)

	// ดึงข้อมูลธุรกิจที่ติดตาม
	businesses, total, err := h.businessFollowService.GetUserFollowedBusinesses(targetUserID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error getting followed businesses: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"businesses": businesses,
			"total":      total,
			"limit":      limit,
			"offset":     offset,
		},
	})
}

// CheckFollowStatus - ตรวจสอบสถานะการติดตาม
func (h *BusinessFollowHandler) CheckFollowStatus(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จากพารามิเตอร์
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err
	}

	// ตรวจสอบสถานะการติดตาม
	isFollowing, err := h.businessFollowService.IsFollowing(userID, businessID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error checking follow status: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"is_following": isFollowing,
		},
	})
}
