// interfaces/api/handler/conversation_member_handler.go
package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/dto"
	"github.com/thizplus/gofiber-chat-api/domain/service"
)

// ConversationMemberHandler จัดการคำขอสำหรับการจัดการสมาชิกในการสนทนา
type ConversationMemberHandler struct {
	memberService service.ConversationMemberService
}

// NewConversationMemberHandler สร้าง handler ใหม่
func NewConversationMemberHandler(memberService service.ConversationMemberService) *ConversationMemberHandler {
	return &ConversationMemberHandler{
		memberService: memberService,
	}
}

// AddConversationMember เพิ่มสมาชิกในการสนทนา
func (h *ConversationMemberHandler) AddConversationMember(c *fiber.Ctx) error {
	// 1. ดึงข้อมูลผู้ใช้จาก context
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// 2. ดึง conversation ID จาก parameter
	conversationIDStr := c.Params("conversationId")
	if conversationIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Conversation ID is required",
		})
	}

	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid conversation ID",
		})
	}

	// 3. รับข้อมูลผู้ใช้ที่ต้องการเพิ่ม
	var input dto.AddMemberRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
	}

	// 4. ตรวจสอบ user ID
	if input.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "User ID is required",
		})
	}

	newMemberID, err := uuid.Parse(input.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid user ID",
		})
	}

	// 5. เรียกใช้ service
	memberDTO, err := h.memberService.AddMember(userID, conversationID, newMemberID)
	if err != nil {
		// จัดการรหัสสถานะตามข้อผิดพลาด
		statusCode := fiber.StatusInternalServerError
		switch err.Error() {
		case "user is already a member of this conversation":
			statusCode = fiber.StatusConflict
		case "user to add not found":
			statusCode = fiber.StatusNotFound
		case "only admins can add members", "you are not a member of this conversation":
			statusCode = fiber.StatusForbidden
		case "cannot add members to direct conversation":
			statusCode = fiber.StatusBadRequest
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	// 6. ส่งผลลัพธ์กลับ
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Member added successfully",
		"data":    memberDTO,
	})
}

// GetConversationMembers ดึงรายการสมาชิกในการสนทนา
func (h *ConversationMemberHandler) GetConversationMembers(c *fiber.Ctx) error {
	// 1. ดึงข้อมูลผู้ใช้จาก context
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// 2. ดึง conversation ID จาก parameter
	conversationIDStr := c.Params("conversationId")
	if conversationIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Conversation ID is required",
		})
	}

	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid conversation ID",
		})
	}

	// 3. ดึงค่า query parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	// 4. เรียกใช้ service
	members, total, err := h.memberService.GetMembers(userID, conversationID, page, limit)
	if err != nil {
		// จัดการรหัสสถานะตามข้อผิดพลาด
		statusCode := fiber.StatusInternalServerError
		if err.Error() == "you are not a member of this conversation" {
			statusCode = fiber.StatusForbidden
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	// 5. ส่งผลลัพธ์กลับ
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Members retrieved successfully",
		"data": fiber.Map{
			"members":     members,
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + limit - 1) / limit,
		},
	})
}

// RemoveConversationMember ลบสมาชิกออกจากการสนทนา
func (h *ConversationMemberHandler) RemoveConversationMember(c *fiber.Ctx) error {
	// 1. ดึงข้อมูลผู้ใช้จาก context
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// 2. ดึง conversation ID จาก parameter
	conversationIDStr := c.Params("conversationId")
	if conversationIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Conversation ID is required",
		})
	}

	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid conversation ID",
		})
	}

	// 3. ดึง user ID ที่ต้องการลบ
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
			"message": "Invalid user ID",
		})
	}

	// 4. เรียกใช้ service
	err = h.memberService.RemoveMember(userID, conversationID, targetUserID)
	if err != nil {
		// จัดการรหัสสถานะตามข้อผิดพลาด
		statusCode := fiber.StatusInternalServerError
		switch err.Error() {
		case "user is not a member of this conversation":
			statusCode = fiber.StatusNotFound
		case "only admins can remove other members", "you are not a member of this conversation":
			statusCode = fiber.StatusForbidden
		case "cannot remove members from direct conversation", "cannot remove the last admin from the conversation":
			statusCode = fiber.StatusBadRequest
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	// 5. ส่งผลลัพธ์กลับ
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Member removed successfully",
	})
}

// ToggleMemberAdmin เปลี่ยนสถานะแอดมินของสมาชิก
func (h *ConversationMemberHandler) ToggleMemberAdmin(c *fiber.Ctx) error {
	// 1. ดึงข้อมูลผู้ใช้จาก context
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// 2. ดึง conversation ID จาก parameter
	conversationIDStr := c.Params("conversationId")
	if conversationIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Conversation ID is required",
		})
	}

	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid conversation ID",
		})
	}

	// 3. ดึง user ID ที่ต้องการเปลี่ยนสถานะ
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
			"message": "Invalid user ID",
		})
	}

	// 4. รับข้อมูลการเปลี่ยนแปลงสถานะ
	var input dto.ToggleAdminRequest
	if err := c.BodyParser(&input); err != nil {
		// ใช้ค่าเริ่มต้นคือ true ถ้าไม่มีข้อมูลส่งมา (toggle)
		input.IsAdmin = true
	}

	// 5. เรียกใช้ service
	isAdmin, err := h.memberService.ToggleAdminStatus(userID, conversationID, targetUserID, input.IsAdmin)
	if err != nil {
		// จัดการรหัสสถานะตามข้อผิดพลาด
		statusCode := fiber.StatusInternalServerError
		switch err.Error() {
		case "user is not a member of this conversation":
			statusCode = fiber.StatusNotFound
		case "only admins can change admin status":
			statusCode = fiber.StatusForbidden
		case "cannot change admin status in direct conversation", "cannot remove admin status from the last admin":
			statusCode = fiber.StatusBadRequest
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	// 6. ส่งผลลัพธ์กลับ
	return c.JSON(fiber.Map{
		"success":  true,
		"message":  "Admin status updated successfully",
		"is_admin": isAdmin,
	})
}
