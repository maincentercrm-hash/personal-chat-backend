// interfaces/api/handler/message_read_handler.go
package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
)

// MessageReadHandler โครงสร้างของ Handler สำหรับจัดการการอ่านข้อความ
type MessageReadHandler struct {
	messageReadService  service.MessageReadService
	notificationService service.NotificationService
}

// NewMessageReadHandler สร้าง Handler ใหม่
func NewMessageReadHandler(
	messageReadService service.MessageReadService,
	notificationService service.NotificationService,
) *MessageReadHandler {
	return &MessageReadHandler{
		messageReadService:  messageReadService,
		notificationService: notificationService,
	}
}

// MarkMessageAsRead จัดการคำขอมาร์คข้อความว่าอ่านแล้ว
func (h *MessageReadHandler) MarkMessageAsRead(c *fiber.Ctx) error {
	// ดึง User UUID จาก context
	userUUID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	messageIDStr := c.Params("messageId")
	if messageIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Message ID is required",
		})
	}

	// แปลง messageID string เป็น UUID
	messageUUID, err := uuid.Parse(messageIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid message ID format",
		})
	}

	// เรียกใช้ service ด้วย UUID
	conversationID, err := h.messageReadService.MarkMessageAsRead(messageUUID, userUUID)
	if err != nil {
		// จัดการข้อผิดพลาด
		statusCode := fiber.StatusInternalServerError

		if err.Error() == "message not found" {
			statusCode = fiber.StatusNotFound
		} else if err.Error() == "you are not a member of this conversation" {
			statusCode = fiber.StatusForbidden
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	// ถ้ามี notificationService ให้ส่งการแจ้งเตือนผ่าน WebSocket
	if h.notificationService != nil && conversationID != uuid.Nil {
		readData := map[string]interface{}{
			"message_id":      messageUUID.String(),
			"user_id":         userUUID.String(),
			"conversation_id": conversationID.String(),
			"read_at":         time.Now(),
		}

		// ส่งการแจ้งเตือนไปยังสมาชิกในการสนทนา
		h.notificationService.NotifyMessageRead(conversationID, readData)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Message marked as read",
	})
}

// GetMessageReads จัดการคำขอดูรายชื่อผู้ที่อ่านข้อความแล้ว
func (h *MessageReadHandler) GetMessageReads(c *fiber.Ctx) error {
	// ดึง User UUID จาก context
	userUUID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	messageIDStr := c.Params("messageId")
	if messageIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Message ID is required",
		})
	}

	// แปลง messageID string เป็น UUID
	messageUUID, err := uuid.Parse(messageIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid message ID format",
		})
	}

	// เรียกใช้ service ด้วย UUID
	reads, err := h.messageReadService.GetMessageReads(messageUUID, userUUID)
	if err != nil {
		// จัดการข้อผิดพลาด
		statusCode := fiber.StatusInternalServerError

		if err.Error() == "message not found" {
			statusCode = fiber.StatusNotFound
		} else if err.Error() == "you are not a member of this conversation" {
			statusCode = fiber.StatusForbidden
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	// แปลงข้อมูลเป็นรูปแบบที่เหมาะสมสำหรับการส่งกลับ
	var result []fiber.Map
	for _, read := range reads {
		result = append(result, fiber.Map{
			"user_id": read.UserID.String(), // แปลง UUID เป็น string เพื่อส่งกลับเป็น JSON
			"read_at": read.ReadAt,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Reads retrieved successfully",
		"data":    result,
	})
}

// MarkAllMessagesAsRead จัดการคำขอมาร์คข้อความทั้งหมดในการสนทนาว่าอ่านแล้ว
func (h *MessageReadHandler) MarkAllMessagesAsRead(c *fiber.Ctx) error {
	// ดึง User UUID จาก context
	userUUID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	conversationIDStr := c.Params("conversationId")
	if conversationIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Conversation ID is required",
		})
	}

	// แปลง conversationID string เป็น UUID
	conversationUUID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid conversation ID format",
		})
	}

	// เรียกใช้ service ด้วย UUID
	markedCount, err := h.messageReadService.MarkAllMessagesAsRead(conversationUUID, userUUID)
	if err != nil {
		// จัดการข้อผิดพลาด
		statusCode := fiber.StatusInternalServerError

		if err.Error() == "you are not a member of this conversation" {
			statusCode = fiber.StatusForbidden
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	// ถ้ามี notificationService ให้ส่งการแจ้งเตือนผ่าน WebSocket
	if h.notificationService != nil && conversationUUID != uuid.Nil {
		h.notificationService.NotifyMessageReadAll(conversationUUID, fiber.Map{
			"conversation_id": conversationUUID.String(),
			"user_id":         userUUID.String(),
			"read_at":         time.Now(),
			"marked_count":    markedCount,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "All messages marked as read",
		"data": fiber.Map{
			"marked_count": markedCount,
		},
	})
}

// GetUnreadCount จัดการคำขอดูจำนวนข้อความที่ยังไม่ได้อ่านในการสนทนา
func (h *MessageReadHandler) GetUnreadCount(c *fiber.Ctx) error {
	// ดึง User UUID จาก context
	userUUID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	conversationIDStr := c.Params("conversationId")
	if conversationIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Conversation ID is required",
		})
	}

	// แปลง conversationID string เป็น UUID
	conversationUUID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid conversation ID format",
		})
	}

	// เรียกใช้ service ด้วย UUID
	count, err := h.messageReadService.GetUnreadCount(conversationUUID, userUUID)
	if err != nil {
		// จัดการข้อผิดพลาด
		statusCode := fiber.StatusInternalServerError

		if err.Error() == "you are not a member of this conversation" {
			statusCode = fiber.StatusForbidden
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Unread count retrieved successfully",
		"data": fiber.Map{
			"unread_count": count,
		},
	})
}
