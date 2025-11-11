// interfaces/api/handler/business_welcome_message_handler.go
package handler

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/domain/types"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
	"github.com/thizplus/gofiber-chat-api/pkg/utils"
)

// BusinessWelcomeMessageHandler จัดการ HTTP requests เกี่ยวกับข้อความต้อนรับของธุรกิจ
type BusinessWelcomeMessageHandler struct {
	welcomeMessageService service.BusinessWelcomeMessageService
}

// NewBusinessWelcomeMessageHandler สร้าง handler ใหม่
func NewBusinessWelcomeMessageHandler(welcomeMessageService service.BusinessWelcomeMessageService) *BusinessWelcomeMessageHandler {
	return &BusinessWelcomeMessageHandler{
		welcomeMessageService: welcomeMessageService,
	}
}

// CreateWelcomeMessage สร้างข้อความต้อนรับใหม่
func (h *BusinessWelcomeMessageHandler) CreateWelcomeMessage(c *fiber.Ctx) error {
	// ดึง User ID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง Business ID จาก URL params
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid business ID: " + err.Error(),
		})
	}

	// รับข้อมูลจาก request body
	var input struct {
		MessageType   string          `json:"message_type"`
		Title         string          `json:"title"`
		Content       string          `json:"content"`
		ImageURL      string          `json:"image_url"`
		ThumbnailURL  string          `json:"thumbnail_url"`
		ActionButtons json.RawMessage `json:"action_buttons"`
		Components    json.RawMessage `json:"components"`
		TriggerType   string          `json:"trigger_type"`
		TriggerParams json.RawMessage `json:"trigger_params"`
		SortOrder     int             `json:"sort_order"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
	}

	// แปลง JSON เป็น types.JSONB
	var actionButtons types.JSONB
	if len(input.ActionButtons) > 0 {
		if err := json.Unmarshal(input.ActionButtons, &actionButtons); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid action buttons format: " + err.Error(),
			})
		}
	}

	var components types.JSONB
	if len(input.Components) > 0 {
		if err := json.Unmarshal(input.Components, &components); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid components format: " + err.Error(),
			})
		}
	}

	var triggerParams types.JSONB
	if len(input.TriggerParams) > 0 {
		if err := json.Unmarshal(input.TriggerParams, &triggerParams); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid trigger params format: " + err.Error(),
			})
		}
	}

	// เรียกใช้ service
	welcomeMessage, err := h.welcomeMessageService.CreateWelcomeMessage(
		businessID,
		userID,
		input.MessageType,
		input.Title,
		input.Content,
		input.ImageURL,
		input.ThumbnailURL,
		actionButtons,
		components,
		input.TriggerType,
		triggerParams,
		input.SortOrder,
	)

	if err != nil {
		statusCode := fiber.StatusInternalServerError
		if err.Error() == "you don't have permission to create welcome message for this business" {
			statusCode = fiber.StatusForbidden
		} else if err.Error() == "invalid message type" || err.Error() == "invalid trigger type" {
			statusCode = fiber.StatusBadRequest
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success":         true,
		"message":         "Welcome message created successfully",
		"welcome_message": welcomeMessage,
	})
}

// GetWelcomeMessages ดึงรายการข้อความต้อนรับของธุรกิจ
func (h *BusinessWelcomeMessageHandler) GetWelcomeMessages(c *fiber.Ctx) error {
	// ดึง User ID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง Business ID จาก URL params
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid business ID: " + err.Error(),
		})
	}

	// ตรวจสอบว่าต้องการดึงข้อความที่ไม่ได้ใช้งานด้วยหรือไม่
	includeInactive := c.QueryBool("include_inactive", false)

	// เรียกใช้ service
	welcomeMessages, err := h.welcomeMessageService.GetBusinessWelcomeMessages(businessID, userID, includeInactive)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		if err.Error() == "you don't have permission to access welcome messages of this business" {
			statusCode = fiber.StatusForbidden
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":          true,
		"message":          "Welcome messages retrieved successfully",
		"welcome_messages": welcomeMessages,
	})
}

// GetWelcomeMessageByID ดึงข้อมูลข้อความต้อนรับตาม ID
func (h *BusinessWelcomeMessageHandler) GetWelcomeMessageByID(c *fiber.Ctx) error {
	// ดึง User ID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง Business ID จาก URL params
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid business ID: " + err.Error(),
		})
	}

	// ดึง Welcome Message ID จาก URL params
	welcomeMessageID, err := utils.ParseUUIDParam(c, "welcomeMessageId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid welcome message ID: " + err.Error(),
		})
	}

	// เรียกใช้ service
	welcomeMessage, err := h.welcomeMessageService.GetWelcomeMessageByID(welcomeMessageID, userID)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		if err.Error() == "you don't have permission to access this welcome message" {
			statusCode = fiber.StatusForbidden
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	// ตรวจสอบว่า welcome message นี้เป็นของธุรกิจที่ระบุหรือไม่
	if welcomeMessage.BusinessID != businessID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Welcome message does not belong to the specified business",
		})
	}

	return c.JSON(fiber.Map{
		"success":         true,
		"message":         "Welcome message retrieved successfully",
		"welcome_message": welcomeMessage,
	})
}

// UpdateWelcomeMessage อัพเดทข้อมูลข้อความต้อนรับ
func (h *BusinessWelcomeMessageHandler) UpdateWelcomeMessage(c *fiber.Ctx) error {
	// ดึง User ID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง Business ID จาก URL params
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid business ID: " + err.Error(),
		})
	}

	// ดึง Welcome Message ID จาก URL params
	welcomeMessageID, err := utils.ParseUUIDParam(c, "welcomeMessageId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid welcome message ID: " + err.Error(),
		})
	}

	// รับข้อมูลจาก request body
	var input map[string]interface{}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
	}

	// แปลงข้อมูลเป็น types.JSONB
	updateData := types.JSONB(input)

	// เรียกใช้ service
	updatedMessage, err := h.welcomeMessageService.UpdateWelcomeMessage(welcomeMessageID, businessID, userID, updateData)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		if err.Error() == "you don't have permission to update welcome message of this business" {
			statusCode = fiber.StatusForbidden
		} else if err.Error() == "welcome message does not belong to the specified business" {
			statusCode = fiber.StatusForbidden
		} else if err.Error() == "no valid data to update" {
			statusCode = fiber.StatusBadRequest
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":         true,
		"message":         "Welcome message updated successfully",
		"welcome_message": updatedMessage,
	})
}

// DeleteWelcomeMessage ลบข้อความต้อนรับ
func (h *BusinessWelcomeMessageHandler) DeleteWelcomeMessage(c *fiber.Ctx) error {
	// ดึง User ID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง Business ID จาก URL params
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid business ID: " + err.Error(),
		})
	}

	// ดึง Welcome Message ID จาก URL params
	welcomeMessageID, err := utils.ParseUUIDParam(c, "welcomeMessageId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid welcome message ID: " + err.Error(),
		})
	}

	// เรียกใช้ service
	err = h.welcomeMessageService.DeleteWelcomeMessage(welcomeMessageID, businessID, userID)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		if err.Error() == "you don't have permission to delete welcome message of this business" {
			statusCode = fiber.StatusForbidden
		} else if err.Error() == "welcome message does not belong to the specified business" {
			statusCode = fiber.StatusForbidden
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Welcome message deleted successfully",
	})
}

// SetWelcomeMessageActive กำหนดสถานะการใช้งานของข้อความต้อนรับ
func (h *BusinessWelcomeMessageHandler) SetWelcomeMessageActive(c *fiber.Ctx) error {
	// ดึง User ID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง Business ID จาก URL params
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid business ID: " + err.Error(),
		})
	}

	// ดึง Welcome Message ID จาก URL params
	welcomeMessageID, err := utils.ParseUUIDParam(c, "welcomeMessageId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid welcome message ID: " + err.Error(),
		})
	}

	// รับข้อมูลจาก request body
	var input struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
	}

	// เรียกใช้ service
	err = h.welcomeMessageService.SetWelcomeMessageActive(welcomeMessageID, businessID, userID, input.IsActive)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		if err.Error() == "you don't have permission to update welcome message of this business" {
			statusCode = fiber.StatusForbidden
		} else if err.Error() == "welcome message does not belong to the specified business" {
			statusCode = fiber.StatusForbidden
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"message":   "Welcome message status updated successfully",
		"is_active": input.IsActive,
	})
}

// UpdateWelcomeMessageSortOrder อัพเดทลำดับการแสดงผลของข้อความต้อนรับ
func (h *BusinessWelcomeMessageHandler) UpdateWelcomeMessageSortOrder(c *fiber.Ctx) error {
	// ดึง User ID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง Business ID จาก URL params
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid business ID: " + err.Error(),
		})
	}

	// ดึง Welcome Message ID จาก URL params
	welcomeMessageID, err := utils.ParseUUIDParam(c, "welcomeMessageId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid welcome message ID: " + err.Error(),
		})
	}

	// รับข้อมูลจาก request body
	var input struct {
		SortOrder int `json:"sort_order"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
	}

	// เรียกใช้ service
	err = h.welcomeMessageService.UpdateWelcomeMessageSortOrder(welcomeMessageID, businessID, userID, input.SortOrder)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		if err.Error() == "you don't have permission to update welcome message of this business" {
			statusCode = fiber.StatusForbidden
		} else if err.Error() == "welcome message does not belong to the specified business" {
			statusCode = fiber.StatusForbidden
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":    true,
		"message":    "Welcome message sort order updated successfully",
		"sort_order": input.SortOrder,
	})
}

// GetWelcomeMessagesByTriggerType ดึงข้อมูลข้อความต้อนรับตามประเภททริกเกอร์
func (h *BusinessWelcomeMessageHandler) GetWelcomeMessagesByTriggerType(c *fiber.Ctx) error {
	// ดึง User ID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง Business ID จาก URL params
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid business ID: " + err.Error(),
		})
	}

	// ดึงประเภททริกเกอร์จาก query params
	triggerType := c.Query("trigger_type")
	if triggerType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Trigger type is required",
		})
	}

	// เรียกใช้ service
	welcomeMessages, err := h.welcomeMessageService.GetWelcomeMessagesByTriggerType(businessID, userID, triggerType)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		if err.Error() == "you don't have permission to access welcome messages of this business" {
			statusCode = fiber.StatusForbidden
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":          true,
		"message":          "Welcome messages retrieved successfully",
		"welcome_messages": welcomeMessages,
		"trigger_type":     triggerType,
	})
}

// TrackMessageClick บันทึกการคลิกในแต่ละแอคชั่นของ welcome message
func (h *BusinessWelcomeMessageHandler) TrackMessageClick(c *fiber.Ctx) error {
	// ดึง Welcome Message ID จาก URL params
	welcomeMessageID, err := utils.ParseUUIDParam(c, "welcomeMessageId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid welcome message ID: " + err.Error(),
		})
	}

	// รับข้อมูลจาก request body
	var input struct {
		ActionType string          `json:"action_type"`
		ActionData json.RawMessage `json:"action_data"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
	}

	// แปลง JSON เป็น types.JSONB
	var actionData types.JSONB
	if len(input.ActionData) > 0 {
		if err := json.Unmarshal(input.ActionData, &actionData); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid action data format: " + err.Error(),
			})
		}
	}

	// เรียกใช้ service
	err = h.welcomeMessageService.TrackMessageClick(welcomeMessageID, input.ActionType, actionData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Click tracked successfully",
	})
}
