// interfaces/api/handler/business_admin_handler.go
package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/service"
)

// BusinessAdminHandler จัดการ HTTP request/response สำหรับแอดมินของธุรกิจ
type BusinessAdminHandler struct {
	businessAdminService service.BusinessAdminService
}

// NewBusinessAdminHandler สร้าง instance ใหม่ของ BusinessAdminHandler
func NewBusinessAdminHandler(businessAdminService service.BusinessAdminService) *BusinessAdminHandler {
	return &BusinessAdminHandler{
		businessAdminService: businessAdminService,
	}
}

// GetAdmins ดึงรายชื่อแอดมินของธุรกิจ
func (h *BusinessAdminHandler) GetAdmins(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userIDStr := c.Locals("userID").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid user ID format",
		})
	}

	businessIDStr := c.Params("businessId")
	if businessIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Business ID is required",
		})
	}

	businessID, err := uuid.Parse(businessIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid business ID format",
		})
	}

	// ดึงข้อมูลแอดมิน
	admins, err := h.businessAdminService.GetAdmins(businessID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error fetching admins: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Admins retrieved successfully",
		"admins":  admins,
	})
}

// AddAdmin เพิ่มแอดมินให้ธุรกิจ
func (h *BusinessAdminHandler) AddAdmin(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userIDStr := c.Locals("userID").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid user ID format",
		})
	}

	businessIDStr := c.Params("businessId")
	if businessIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Business ID is required",
		})
	}

	businessID, err := uuid.Parse(businessIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid business ID format",
		})
	}

	// รับข้อมูลแอดมินใหม่
	var input struct {
		Username string `json:"username"`
		UserID   string `json:"user_id"`
		Role     string `json:"role"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
	}

	var newAdminUserID uuid.UUID

	// ตรวจสอบข้อมูลที่จำเป็น - รับได้ทั้ง username หรือ user_id
	if input.UserID != "" {
		newAdminUserID, err = uuid.Parse(input.UserID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid user ID format",
			})
		}
	} else if input.Username != "" {
		// ในกรณีนี้ควรเพิ่ม method ใน UserRepository สำหรับค้นหาด้วย username
		// แต่เนื่องจากเราไม่มี method นี้ในตัวอย่าง ให้ return error ไปก่อน
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "User ID is required",
		})
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "User ID or username is required",
		})
	}

	// เพิ่มแอดมินใหม่
	admin, err := h.businessAdminService.AddAdmin(businessID, userID, newAdminUserID, input.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error adding admin: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Admin added successfully",
		"admin":   admin,
	})
}

// RemoveAdmin ลบแอดมินออกจากธุรกิจ
func (h *BusinessAdminHandler) RemoveAdmin(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userIDStr := c.Locals("userID").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid user ID format",
		})
	}

	businessIDStr := c.Params("businessId")
	if businessIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Business ID is required",
		})
	}

	businessID, err := uuid.Parse(businessIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid business ID format",
		})
	}

	targetUserIDStr := c.Params("userId")
	if targetUserIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "User ID is required",
		})
	}

	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid target user ID format",
		})
	}

	// ลบแอดมิน
	err = h.businessAdminService.RemoveAdmin(businessID, userID, targetUserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error removing admin: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Admin removed successfully",
	})
}

// ChangeAdminRole เปลี่ยนบทบาทของแอดมิน
func (h *BusinessAdminHandler) ChangeAdminRole(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userIDStr := c.Locals("userID").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid user ID format",
		})
	}

	businessIDStr := c.Params("businessId")
	if businessIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Business ID is required",
		})
	}

	businessID, err := uuid.Parse(businessIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid business ID format",
		})
	}

	targetUserIDStr := c.Params("userId")
	if targetUserIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "User ID is required",
		})
	}

	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid target user ID format",
		})
	}

	// รับข้อมูลบทบาทใหม่
	var input struct {
		Role string `json:"role"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
	}

	if input.Role == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Role is required",
		})
	}

	// เปลี่ยนบทบาท
	admin, err := h.businessAdminService.ChangeAdminRole(businessID, userID, targetUserID, input.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error updating admin role: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Admin role updated successfully",
		"admin":   admin,
	})
}
