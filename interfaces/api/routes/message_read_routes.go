// interfaces/api/routes/message_read_routes.go
package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/handler"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
)

// SetupMessageReadRoutes กำหนดเส้นทาง API สำหรับการอ่านข้อความ
func SetupMessageReadRoutes(router fiber.Router, messageReadHandler *handler.MessageReadHandler) {
	// สร้างกลุ่มเส้นทางข้อความ
	messages := router.Group("/messages")
	messages.Use(middleware.Protected())

	// เส้นทางสำหรับการอ่านข้อความ
	messages.Post("/:messageId/read", messageReadHandler.MarkMessageAsRead) // [success] 11.1 การมาร์คข้อความว่าอ่านแล้ว [Y]
	messages.Get("/:messageId/reads", messageReadHandler.GetMessageReads)   // [success] 11.2 การดูรายชื่อผู้ที่อ่านข้อความแล้ว [Y]

	// สร้างกลุ่มเส้นทางการสนทนา
	conversations := router.Group("/conversations")
	conversations.Use(middleware.Protected())

	// เส้นทางสำหรับการจัดการข้อความทั้งหมดในการสนทนา
	conversations.Post("/:conversationId/read_all", messageReadHandler.MarkAllMessagesAsRead) // [success] 11.3 การมาร์คข้อความทั้งหมดในการสนทนาว่าอ่านแล้ว [Y]
	conversations.Get("/:conversationId/unread_count", messageReadHandler.GetUnreadCount)     // [success] 11.4 การดูจำนวนข้อความที่ยังไม่อ่าน [Y]
}
