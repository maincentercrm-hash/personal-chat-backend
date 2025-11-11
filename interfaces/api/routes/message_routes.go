// interfaces/api/routes/message_routes.go
package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/handler"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
)

// SetupMessageRoutes กำหนดเส้นทาง API สำหรับข้อความ
func SetupMessageRoutes(router fiber.Router, messageHandler *handler.MessageHandler) {
	// สร้างกลุ่มเส้นทางข้อความ
	messages := router.Group("/messages")
	messages.Use(middleware.Protected())

	// เส้นทางจัดการข้อความ
	messages.Patch("/:messageId", messageHandler.EditMessage)                          // [success] 10.5 การแก้ไขข้อความ [Y]
	messages.Get("/:messageId/edit-history", messageHandler.GetMessageEditHistory)     // [success] 10.6 การดูประวัติการแก้ไขข้อความ [Y]
	messages.Delete("/:messageId", messageHandler.DeleteMessage)                       // [success] 10.7 การลบข้อความ [Y]
	messages.Get("/:messageId/delete-history", messageHandler.GetMessageDeleteHistory) // [success] 10.8 การดูประวัติการลบข้อความ [Y]
	messages.Post("/:messageId/reply", messageHandler.ReplyToMessage)                  // [success] 10.9 การตอบกลับข้อความ [Y]

	// เส้นทางส่งข้อความประเภทต่างๆ ของบัญชีธรรมดา
	conversations := router.Group("/conversations")
	conversations.Use(middleware.Protected())

	conversations.Post("/:conversationId/messages/text", messageHandler.SendTextMessage)       //  [success] 10.1 การส่งข้อความประเภทข้อความ [Y]
	conversations.Post("/:conversationId/messages/sticker", messageHandler.SendStickerMessage) //  [success] 10.2 การส่งข้อความประเภทสติกเกอร์ [Y]
	conversations.Post("/:conversationId/messages/image", messageHandler.SendImageMessage)     //  [success] 10.3 การส่งข้อความประเภทรูปภาพ [Y]
	conversations.Post("/:conversationId/messages/file", messageHandler.SendFileMessage)       //  [success] 10.4 การส่งข้อความประเภทไฟล์ [Y]
}
