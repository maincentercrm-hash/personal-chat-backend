// interfaces/api/handler/message_handler.go
package handler

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/domain/types"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
	"github.com/thizplus/gofiber-chat-api/pkg/utils"
)

// MessageHandler โครงสร้างของ Handler สำหรับจัดการข้อความ
type MessageHandler struct {
	messageService      service.MessageService
	notificationService service.NotificationService
}

// NewMessageHandler สร้าง Handler ใหม่
func NewMessageHandler(
	messageService service.MessageService,
	notificationService service.NotificationService,
) *MessageHandler {
	return &MessageHandler{
		messageService:      messageService,
		notificationService: notificationService,
	}
}

// SendTextMessage จัดการคำขอส่งข้อความประเภทข้อความ
func (h *MessageHandler) SendTextMessage(c *fiber.Ctx) error {
	// ดึง User ID จาก context ที่ตั้งค่าโดย middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	conversationID, err := utils.ParseUUIDParam(c, "conversationId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// รับข้อมูลข้อความจาก request body
	var input struct {
		TempID   string      `json:"temp_id"`
		Content  string      `json:"content"`
		Metadata types.JSONB `json:"metadata"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body: " + err.Error(),
		})
	}

	// บันทึก temp_id ลงใน metadata ถ้ามี (JSONB เป็น map[string]interface{} อยู่แล้ว)
	metadata := input.Metadata
	if input.TempID != "" {
		if metadata == nil {
			metadata = make(types.JSONB)
		}
		metadata["tempId"] = input.TempID
	}

	// เรียกใช้ service
	message, err := h.messageService.SendTextMessage(conversationID, userID, input.Content, metadata)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		// ตรวจสอบประเภทข้อผิดพลาดเพื่อกำหนด status code ที่เหมาะสม
		if err.Error() == "user is not a member of this conversation" {
			statusCode = fiber.StatusForbidden
		} else if err.Error() == "message content cannot be empty" {
			statusCode = fiber.StatusBadRequest
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	messageJson, err := json.MarshalIndent(message, "", "  ")
	if err != nil {
		fmt.Printf("[ERROR] Failed to marshal message: %v\n", err)
	} else {
		fmt.Printf("[XXXXXXX]Message sent successfully:\n%s\n", string(messageJson))
	}

	h.notificationService.NotifyNewMessage(conversationID, message)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Message sent successfully",
		"data":    message,
	})
}

// SendStickerMessage จัดการคำขอส่งข้อความประเภทสติกเกอร์
func (h *MessageHandler) SendStickerMessage(c *fiber.Ctx) error {
	// ดึง User ID จาก context
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	conversationID, err := utils.ParseUUIDParam(c, "conversationId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// รับข้อมูลสติกเกอร์จาก request body
	var input struct {
		TempID            string      `json:"temp_id"`
		StickerID         uuid.UUID   `json:"sticker_id"`
		StickerSetID      uuid.UUID   `json:"sticker_set_id"`
		MediaURL          string      `json:"media_url"`
		MediaThumbnailURL string      `json:"media_thumbnail_url"`
		Metadata          types.JSONB `json:"metadata"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body: " + err.Error(),
		})
	}

	// บันทึก temp_id ลงใน metadata ถ้ามี (JSONB เป็น map[string]interface{} อยู่แล้ว)
	metadata := input.Metadata
	if input.TempID != "" {
		if metadata == nil {
			metadata = make(types.JSONB)
		}
		metadata["tempId"] = input.TempID
	}

	// เรียกใช้ service
	message, err := h.messageService.SendStickerMessage(
		conversationID,
		userID,
		input.StickerID,
		input.StickerSetID,
		input.MediaURL,
		input.MediaThumbnailURL,
		metadata,
	)

	if err != nil {
		statusCode := fiber.StatusInternalServerError
		// ตรวจสอบประเภทข้อผิดพลาด
		if err.Error() == "user is not a member of this conversation" {
			statusCode = fiber.StatusForbidden
		} else if err.Error() == "sticker URL is required" {
			statusCode = fiber.StatusBadRequest
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	h.notificationService.NotifyNewMessage(conversationID, message)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Sticker sent successfully",
		"data":    message,
	})
}

// SendImageMessage จัดการคำขอส่งข้อความประเภทรูปภาพ
func (h *MessageHandler) SendImageMessage(c *fiber.Ctx) error {
	// ดึง User ID จาก context
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	conversationID, err := utils.ParseUUIDParam(c, "conversationId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// รับข้อมูลรูปภาพจาก request body
	var input struct {
		TempID            string      `json:"temp_id"`
		MediaURL          string      `json:"media_url"`
		MediaThumbnailURL string      `json:"media_thumbnail_url"`
		Caption           string      `json:"caption"`
		Metadata          types.JSONB `json:"metadata"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body: " + err.Error(),
		})
	}

	// บันทึก temp_id ลงใน metadata ถ้ามี (JSONB เป็น map[string]interface{} อยู่แล้ว)
	metadata := input.Metadata
	if input.TempID != "" {
		if metadata == nil {
			metadata = make(types.JSONB)
		}
		metadata["tempId"] = input.TempID
	}

	// เรียกใช้ service
	message, err := h.messageService.SendImageMessage(
		conversationID,
		userID,
		input.MediaURL,
		input.MediaThumbnailURL,
		input.Caption,
		metadata,
	)

	if err != nil {
		statusCode := fiber.StatusInternalServerError
		// ตรวจสอบประเภทข้อผิดพลาด
		if err.Error() == "user is not a member of this conversation" {
			statusCode = fiber.StatusForbidden
		} else if err.Error() == "image URL is required" {
			statusCode = fiber.StatusBadRequest
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	h.notificationService.NotifyNewMessage(conversationID, message)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Image sent successfully",
		"data":    message,
	})
}

// SendFileMessage จัดการคำขอส่งข้อความประเภทไฟล์
func (h *MessageHandler) SendFileMessage(c *fiber.Ctx) error {
	// ดึง User ID จาก context
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	conversationID, err := utils.ParseUUIDParam(c, "conversationId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// รับข้อมูลไฟล์จาก request body
	var input struct {
		TempID   string      `json:"temp_id"`
		MediaURL string      `json:"media_url"`
		FileName string      `json:"file_name"`
		FileSize int64       `json:"file_size"`
		FileType string      `json:"file_type"`
		Metadata types.JSONB `json:"metadata"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body: " + err.Error(),
		})
	}

	// บันทึก temp_id ลงใน metadata ถ้ามี (JSONB เป็น map[string]interface{} อยู่แล้ว)
	metadata := input.Metadata
	if input.TempID != "" {
		if metadata == nil {
			metadata = make(types.JSONB)
		}
		metadata["tempId"] = input.TempID
	}

	// เรียกใช้ service
	message, err := h.messageService.SendFileMessage(
		conversationID,
		userID,
		input.MediaURL,
		input.FileName,
		input.FileSize,
		input.FileType,
		metadata,
	)

	if err != nil {
		statusCode := fiber.StatusInternalServerError
		// ตรวจสอบประเภทข้อผิดพลาด
		if err.Error() == "user is not a member of this conversation" {
			statusCode = fiber.StatusForbidden
		} else if err.Error() == "file URL is required" {
			statusCode = fiber.StatusBadRequest
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	h.notificationService.NotifyNewMessage(conversationID, message)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "File sent successfully",
		"data":    message,
	})
}

// EditMessage จัดการคำขอแก้ไขข้อความ
// EditMessage จัดการคำขอแก้ไขข้อความ
func (h *MessageHandler) EditMessage(c *fiber.Ctx) error {
	// ดึง User ID จาก context
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	messageID, err := utils.ParseUUIDParam(c, "messageId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// รับข้อมูลการแก้ไขจาก request body
	var input struct {
		Content string `json:"content"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body: " + err.Error(),
		})
	}

	// เรียกใช้ service
	message, err := h.messageService.EditMessage(messageID, userID, input.Content)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		// ตรวจสอบประเภทข้อผิดพลาด
		if err.Error() == "message not found" {
			statusCode = fiber.StatusNotFound
		} else if err.Error() == "only message owner can edit messages" {
			statusCode = fiber.StatusForbidden
		} else if err.Error() == "cannot edit deleted message" || err.Error() == "only text messages can be edited" {
			statusCode = fiber.StatusBadRequest
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	// 🔥 เพิ่มส่วนนี้: ส่ง WebSocket notification สำหรับการแก้ไขข้อความ
	h.notificationService.NotifyMessageEdited(message.ConversationID, message)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Message updated successfully",
		"data":    message,
	})
}

// DeleteMessage จัดการคำขอลบข้อความ
func (h *MessageHandler) DeleteMessage(c *fiber.Ctx) error {
	// ดึง User ID จาก context
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	messageID, err := utils.ParseUUIDParam(c, "messageId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// เรียกใช้ service
	err = h.messageService.DeleteMessage(messageID, userID)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		// ตรวจสอบประเภทข้อผิดพลาด
		if err.Error() == "message not found" {
			statusCode = fiber.StatusNotFound
		} else if err.Error() == "only message owner or conversation admin can delete messages" {
			statusCode = fiber.StatusForbidden
		} else if err.Error() == "message is already deleted" {
			statusCode = fiber.StatusBadRequest
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Message deleted successfully",
	})
}

// GetMessageEditHistory จัดการคำขอดูประวัติการแก้ไขข้อความ
func (h *MessageHandler) GetMessageEditHistory(c *fiber.Ctx) error {
	// ดึง User ID จาก context
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	messageID, err := utils.ParseUUIDParam(c, "messageId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// เรียกใช้ service
	history, err := h.messageService.GetMessageEditHistory(messageID, userID)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		// ตรวจสอบประเภทข้อผิดพลาด
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

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Edit history retrieved successfully",
		"data":    history,
	})
}

// GetMessageDeleteHistory จัดการคำขอดูประวัติการลบข้อความ
func (h *MessageHandler) GetMessageDeleteHistory(c *fiber.Ctx) error {
	// ดึง User ID จาก context
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	messageID, err := utils.ParseUUIDParam(c, "messageId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// เรียกใช้ service
	history, err := h.messageService.GetMessageDeleteHistory(messageID, userID)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		// ตรวจสอบประเภทข้อผิดพลาด
		if err.Error() == "message not found" {
			statusCode = fiber.StatusNotFound
		} else if err.Error() == "only admins can view delete history" {
			statusCode = fiber.StatusForbidden
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Delete history retrieved successfully",
		"data":    history,
	})
}

// ReplyToMessage จัดการคำขอตอบกลับข้อความ
func (h *MessageHandler) ReplyToMessage(c *fiber.Ctx) error {
	// ดึง User ID จาก context
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	replyToID, err := utils.ParseUUIDParam(c, "messageId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// รับข้อมูลการตอบกลับจาก request body
	var input struct {
		MessageType       string      `json:"message_type"`
		Content           string      `json:"content"`
		MediaURL          string      `json:"media_url"`
		MediaThumbnailURL string      `json:"media_thumbnail_url"`
		Metadata          types.JSONB `json:"metadata"`
		SenderType        string      `json:"sender_type"` // เพิ่ม field นี้
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body: " + err.Error(),
		})
	}

	// เรียกใช้ service
	message, err := h.messageService.ReplyToMessage(
		replyToID,
		userID,
		input.MessageType,
		input.Content,
		input.MediaURL,
		input.MediaThumbnailURL,
		input.Metadata,
	)

	if err != nil {
		statusCode := fiber.StatusInternalServerError
		// ตรวจสอบประเภทข้อผิดพลาด
		if err.Error() == "message not found" {
			statusCode = fiber.StatusNotFound
		} else if err.Error() == "you are not a member of this conversation" {
			statusCode = fiber.StatusForbidden
		} else if err.Error() == "cannot reply to deleted message" || err.Error() == "invalid message type" {
			statusCode = fiber.StatusBadRequest
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	h.notificationService.NotifyNewMessage(message.ConversationID, message)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Reply sent successfully",
		"data":    message,
	})
}

