// interfaces/api/routes/customer_profile_routes.go
package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/handler"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
)

// SetupCustomerProfileRoutes ตั้งค่า routes สำหรับ Customer Profile
func SetupCustomerProfileRoutes(
	router fiber.Router,
	customerProfileHandler *handler.CustomerProfileHandler,
	businessAdminService service.BusinessAdminService,
) {
	// สร้างกลุ่มเส้นทางธุรกิจ
	businesses := router.Group("/businesses")
	businesses.Use(middleware.Protected())

	// กลุ่ม routes สำหรับ customer profiles
	// Base path: /api/v1/businesses/:businessId/customers
	customers := businesses.Group("/:businessId/customers")

	// ใช้ middleware ตรวจสอบสิทธิ์แอดมิน
	customers.Use(middleware.CheckBusinessAdmin(businessAdminService))

	// List & Search Routes (เส้นทางคงที่ควรมาก่อน)
	// ดึงรายชื่อลูกค้าทั้งหมดของธุรกิจ (มี pagination)
	// Query params: ?limit=20&offset=0
	customers.Get("/", customerProfileHandler.GetBusinessCustomers) // [success] 15.4 การดึงลูกค้าทั้งหมดของธุรกิจ [Y]

	// ค้นหาลูกค้า
	// Query params: ?q=keyword&limit=20&offset=0
	customers.Get("/search", customerProfileHandler.SearchCustomers) // [success] 15.5 การค้นหาลูกค้า [Y]

	// Customer Profile Management Routes (เส้นทางที่มีพารามิเตอร์มาทีหลัง)
	// สร้างโปรไฟล์ลูกค้าใหม่
	customers.Post("/:userId", customerProfileHandler.CreateCustomerProfile) // [success] 15.1 การสร้างโปรไฟล์ลูกค้า [Y]

	// ดึงโปรไฟล์ลูกค้า
	customers.Get("/:userId", customerProfileHandler.GetCustomerProfile) // [success] 15.2 การดึงโปรไฟล์ลูกค้า [Y]

	// อัปเดตโปรไฟล์ลูกค้า (admin เท่านั้น)
	customers.Patch("/:userId", customerProfileHandler.UpdateCustomerProfile) // [success] 15.3 การอัปเดตโปรไฟล์ลูกค้า [Y]

	// Utility Routes (เส้นทางที่มีพารามิเตอร์และมีความเฉพาะเจาะจงมากขึ้น)
	// อัปเดตเวลาติดต่อล่าสุด last_contact_at (สำหรับ internal use หรือ webhook)
	customers.Put("/:userId/contact", customerProfileHandler.UpdateLastContact) // [success] 15.6 การอัปเดตเวลาติดต่อล่าสุด [Y]

	// Additional Analytics Routes (Optional - สำหรับอนาคต)
	//analytics := customers.Group("/:userId/analytics")
	// สรุปข้อมูลลูกค้า (จำนวนข้อความ, แท็ก, กิจกรรมล่าสุด)
	//analytics.Get("/summary", customerProfileHandler.GetCustomerSummary)
	// กิจกรรมของลูกค้า (ประวัติการสนทนา, การซื้อขาย)
	//analytics.Get("/activity", customerProfileHandler.GetCustomerActivity)
}
