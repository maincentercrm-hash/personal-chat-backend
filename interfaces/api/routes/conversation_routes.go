// interfaces/api/routes/conversation_routes.go
package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/handler"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
)

// SetupConversationRoutes กำหนดเส้นทางสำหรับการสนทนาส่วนตัว
func SetupConversationRoutes(
	router fiber.Router,
	conversationHandler *handler.ConversationHandler,
	conversationMemberHandler *handler.ConversationMemberHandler,
) {
	// สร้างกลุ่มเส้นทางการสนทนา
	conversations := router.Group("/conversations")
	conversations.Use(middleware.Protected())

	// เส้นทางหลัก - เฉพาะการสนทนาส่วนตัว
	conversations.Post("/", conversationHandler.Create)              // [success] 8.1 การสร้างการสนทนา [direct,group,business]
	conversations.Get("/", conversationHandler.GetUserConversations) // [success] 8.2 การดึงรายการการสนทนา [Y]

	// เส้นทางเฉพาะการสนทนา
	conversations.Patch("/:conversationId", conversationHandler.UpdateConversation)             // [success] 8.3 การอัปเดตข้อมูลการสนทนา [Y]
	conversations.Get("/:conversationId/messages", conversationHandler.GetConversationMessages) // [succcess] 8.4 การดึงข้อความในการสนทนา [Y]
	conversations.Patch("/:conversationId/pin", conversationHandler.TogglePinConversation)      // [success] 8.5 การเปลี่ยนสถานะปักหมุดของการสนทนา [Y]
	conversations.Patch("/:conversationId/mute", conversationHandler.ToggleMuteConversation)    // [success] 8.6 การเปลี่ยนสถานะการปิดเสียงของการสนทนา [Y]

	// การจัดการสมาชิกในกลุ่ม
	conversations.Post("/:conversationId/members", conversationMemberHandler.AddConversationMember)              // [success] 9.1 การเพิ่มสมาชิกในการสนทนา [Y]
	conversations.Get("/:conversationId/members", conversationMemberHandler.GetConversationMembers)              // [success] 9.2 การดึงรายชื่อสมาชิกในการสนทนา [Y]
	conversations.Delete("/:conversationId/members/:userId", conversationMemberHandler.RemoveConversationMember) // [success] 9.4 การลบสมาชิกจากการสนทนา [Y]
	conversations.Patch("/:conversationId/members/:userId/admin", conversationMemberHandler.ToggleMemberAdmin)   // [success] 9.3 การเปลี่ยนสถานะแอดมินของสมาชิก [Y]
}
