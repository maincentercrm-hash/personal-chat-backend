// interfaces/api/routes/business_conversation_routes.go
package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/handler"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
)

// SetupBusinessConversationRoutes กำหนดเส้นทางสำหรับการสนทนาในนามธุรกิจ
func SetupBusinessConversationRoutes(
	router fiber.Router,
	conversationHandler *handler.ConversationHandler,
	businessAdminService service.BusinessAdminService,
) {
	// สร้างกลุ่มเส้นทางธุรกิจ
	businesses := router.Group("/businesses")
	businesses.Use(middleware.Protected())

	// สร้างกลุ่มเส้นทางการสนทนาของธุรกิจ
	businessConversations := businesses.Group("/:businessId/conversations")

	// ใช้ middleware ตรวจสอบสิทธิ์แอดมินเฉพาะกับการสนทนาธุรกิจ
	businessConversations.Use(middleware.CheckBusinessAdmin(businessAdminService))

	// ดูการสนทนาทั้งหมดของธุรกิจ (admin/owner เห็นเหมือนกัน)
	businessConversations.Get("/", conversationHandler.GetBusinessConversations) // [success] 13.1 การดึงการสนทนาทั้งหมดของธุรกิจ [Y]

	// ดูข้อความในการสนทนาธุรกิจ
	businessConversations.Get("/:conversationId/messages", conversationHandler.GetBusinessConversationMessages) // [success] 13.2 การดูข้อความในการสนทนาธุรกิจ [Y]
}
