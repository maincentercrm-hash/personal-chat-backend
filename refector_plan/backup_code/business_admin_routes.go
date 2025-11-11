// interfaces/api/routes/business_admin_routes.go
package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/handler"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
)

// SetupBusinessAdminRoutes กำหนดเส้นทาง API สำหรับจัดการแอดมินของธุรกิจ
func SetupBusinessAdminRoutes(router fiber.Router, businessAdminHandler *handler.BusinessAdminHandler, businessAdminService service.BusinessAdminService) {
	// สร้างกลุ่มเส้นทางธุรกิจ
	businesses := router.Group("/businesses")
	businesses.Use(middleware.Protected())

	// สร้างกลุ่ม admins สำหรับแต่ละธุรกิจ
	admins := businesses.Group("/:businessId/admins")

	// ตรวจสอบสิทธิ์แอดมินสำหรับทุกเส้นทางใน admins group
	admins.Use(middleware.CheckBusinessAdmin(businessAdminService))

	// เส้นทางแอดมินธุรกิจ
	admins.Get("/", businessAdminHandler.GetAdmins)                     // [success] 6.1 การดึงรายชื่อแอดมิน [Y]
	admins.Post("/", businessAdminHandler.AddAdmin)                     // [success] 6.2 การเพิ่มแอดมิน [Y]
	admins.Delete("/:userId", businessAdminHandler.RemoveAdmin)         // [success] 6.4 การลบแอดมิน [Y]
	admins.Patch("/:userId/role", businessAdminHandler.ChangeAdminRole) // [success] 6.3 การเปลี่ยนบทบาทแอดมิน [Y]
}
