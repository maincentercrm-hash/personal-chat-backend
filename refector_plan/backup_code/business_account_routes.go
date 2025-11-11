// interfaces/api/routes/business_account_routes.go
package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/handler"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
)

// SetupBusinessAccountRoutes กำหนดเส้นทาง API สำหรับแอคเคาท์ธุรกิจ
func SetupBusinessAccountRoutes(router fiber.Router, businessAccountHandler *handler.BusinessAccountHandler) {
	// สร้างกลุ่มเส้นทางธุรกิจ
	businesses := router.Group("/businesses")
	businesses.Use(middleware.Protected())

	// -- เส้นทางเฉพาะเจาะจงต้องกำหนดก่อนเส้นทางที่มี parameter

	// เส้นทางหลักที่ไม่มี parameter
	businesses.Post("/", businessAccountHandler.Create)           // [success] 5.1 การสร้างบัญชีธุรกิจ [Y]
	businesses.Get("/", businessAccountHandler.GetUserBusinesses) // [success] 5.2 การดึงรายการบัญชีธุรกิจของผู้ใช้ [Y]

	// เส้นทางค้นหาธุรกิจด้วย username (ต้องอยู่ก่อนเส้นทาง /:businessId)
	businesses.Get("/search", businessAccountHandler.SearchByUsername)     // [success] 5.3 การค้นหาธุรกิจด้วย Username [Y]
	businesses.Get("/search-all", businessAccountHandler.SearchBusinesses) // [success] 5.4 การค้นหาธุรกิจทั้งหมด [Y]

	// -- เส้นทางที่มี parameter จะอยู่หลังเส้นทางเฉพาะเจาะจง
	// เส้นทางจัดการธุรกิจหลัก
	businesses.Get("/:businessId", businessAccountHandler.GetDetail) // [success] 5.5 การดูรายละเอียดธุรกิจ [Y]
	businesses.Patch("/:businessId", businessAccountHandler.Update)  // [success] 5.6 การอัปเดตข้อมูลธุรกิจ [Y]
	businesses.Delete("/:businessId", businessAccountHandler.Delete) // [success] 5.9 การลบธุรกิจ [Y]

	// อัปโหลดรูปภาพสำหรับธุรกิจ
	businesses.Put("/:businessId/profile-image", businessAccountHandler.UploadProfileImage) // [success] 5.7 การอัปโหลดรูปโปรไฟล์ธุรกิจ [Y]
	businesses.Put("/:businessId/cover-image", businessAccountHandler.UploadCoverImage)     // [success] 5.8 การอัปโหลดรูปปกธุรกิจ [Y]
}
