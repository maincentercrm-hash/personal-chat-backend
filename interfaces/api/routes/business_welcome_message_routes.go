// interfaces/api/routes/business_welcome_message_routes.go
package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/handler"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
)

// SetupBusinessWelcomeMessageRoutes กำหนดเส้นทางสำหรับการจัดการข้อความต้อนรับของธุรกิจ
func SetupBusinessWelcomeMessageRoutes(
	router fiber.Router,
	welcomeMessageHandler *handler.BusinessWelcomeMessageHandler,
	businessAdminService service.BusinessAdminService,
) {
	// สร้างกลุ่มเส้นทางธุรกิจ
	businesses := router.Group("/businesses")
	businesses.Use(middleware.Protected())

	// สร้างกลุ่มย่อยสำหรับข้อความต้อนรับ
	welcomeMessages := businesses.Group("/:businessId/welcome-messages")

	// เส้นทางที่ต้องการสิทธิ์แอดมินธุรกิจ
	adminRoutes := welcomeMessages.Group("/")
	adminRoutes.Use(middleware.CheckBusinessAdmin(businessAdminService))

	// เส้นทางการจัดการข้อความต้อนรับ

	/*
		19.1 การสร้างข้อความต้อนรับใหม่ [Y]
		19.2 การสร้างข้อความต้อนรับประเภทรูปภาพ [Y]
		19.3 การสร้างข้อความต้อนรับแบบมีปุ่มกด [Y]
		19.4 การสร้างข้อความต้อนรับแบบการ์ด [Y]
		19.5 การสร้างข้อความต้อนรับแบบ Carousel [Y]
		19.6 การสร้างข้อความต้อนรับแบบ Command [Y]
		19.7 การสร้างข้อความต้อนรับแบบ Inactive [Y]
	*/
	adminRoutes.Post("/", welcomeMessageHandler.CreateWelcomeMessage)                                       // [success] การสร้างข้อความต้อนรับใหม่ [Y]
	adminRoutes.Get("/", welcomeMessageHandler.GetWelcomeMessages)                                          // [success] 19.8 การดึงข้อความต้อนรับทั้งหมด [Y]
	adminRoutes.Get("/by-trigger", welcomeMessageHandler.GetWelcomeMessagesByTriggerType)                   // [success] 19.11 การดึงข้อความต้อนรับตามประเภททริกเกอร์ [Y]
	adminRoutes.Get("/:welcomeMessageId", welcomeMessageHandler.GetWelcomeMessageByID)                      // [success] 19.10 การดึงข้อความต้อนรับตาม ID [Y]
	adminRoutes.Patch("/:welcomeMessageId", welcomeMessageHandler.UpdateWelcomeMessage)                     // [success] 19.12 การอัปเดตข้อความต้อนรับ [Y]
	adminRoutes.Delete("/:welcomeMessageId", welcomeMessageHandler.DeleteWelcomeMessage)                    // [success] 19.16 การลบข้อความต้อนรับ [Y]
	adminRoutes.Patch("/:welcomeMessageId/status", welcomeMessageHandler.SetWelcomeMessageActive)           // [success] 19.13 การเปิด/ปิดการใช้งานข้อความต้อนรับ [Y]
	adminRoutes.Patch("/:welcomeMessageId/sort-order", welcomeMessageHandler.UpdateWelcomeMessageSortOrder) // [success] 19.14 การอัปเดตลำดับการแสดงผล [Y]

	// เส้นทางที่ไม่ต้องการสิทธิ์แอดมิน (สำหรับการติดตามการใช้งาน)
	welcomeMessages.Post("/:welcomeMessageId/track-click", welcomeMessageHandler.TrackMessageClick) // [pending] 19.15 การบันทึกการคลิกปุ่มในข้อความต้อนรับ [P]
}
