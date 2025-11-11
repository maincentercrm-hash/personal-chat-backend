// interfaces/api/routes/business_message_routes.go
package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/handler"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
)

// SetupBusinessMessageRoutes กำหนดเส้นทางสำหรับการส่งข้อความในนามธุรกิจ
func SetupBusinessMessageRoutes(
	router fiber.Router,
	messageHandler *handler.MessageHandler,
	businessAdminService service.BusinessAdminService,
) {
	// สร้างกลุ่มเส้นทางธุรกิจ
	businesses := router.Group("/businesses")
	businesses.Use(middleware.Protected())

	// สร้างกลุ่มย่อยสำหรับการส่งข้อความในนามธุรกิจ
	businessMessages := businesses.Group("/:businessId/conversations/:conversationId/messages")

	// ใช้ middleware ตรวจสอบสิทธิ์แอดมิน
	businessMessages.Use(middleware.CheckBusinessAdmin(businessAdminService))

	// เส้นทางส่งข้อความในนามธุรกิจ
	businessMessages.Post("/text", messageHandler.SendBusinessTextMessage)       // [success] 19.1 การส่งข้อความประเภทข้อความในนามธุรกิจ [Y]
	businessMessages.Post("/sticker", messageHandler.SendBusinessStickerMessage) // [success] 19.2 การส่งข้อความประเภทสติกเกอร์ในนามธุรกิจ [Y]
	businessMessages.Post("/image", messageHandler.SendBusinessImageMessage)     // [success] 19.3 การส่งข้อความประเภทรูปภาพในนามธุรกิจ [Y]
	businessMessages.Post("/file", messageHandler.SendBusinessFileMessage)       // [success] 19.4 การส่งข้อความประเภทไฟล์ในนามธุรกิจ [Y]
}
