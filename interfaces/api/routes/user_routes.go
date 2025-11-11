// interfaces/api/routes/user_routes.go
package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/handler"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
)

// SetupUserRoutes กำหนดเส้นทาง API สำหรับผู้ใช้
func SetupUserRoutes(router fiber.Router, userHandler *handler.UserHandler) {
	// สร้างกลุ่มเส้นทาง User
	userRoutes := router.Group("/users")
	userRoutes.Use(middleware.Protected())

	// เส้นทางค้นหาผู้ใช้
	userRoutes.Get("/search", userHandler.SearchUsers)                // [success] 2.5 การค้นหาผู้ใช้ [Y]
	userRoutes.Get("/status", userHandler.GetStatus)                  // [success] 2.6 การดูสถานะผู้ใช้ [Y]
	userRoutes.Get("/search-by-email", userHandler.SearchUserByEmail) // [์new] ยังไม่ได้เพิ่มใน postman

	// ดึงข้อมูลผู้ใช้ปัจจุบัน
	userRoutes.Get("/me", userHandler.GetCurrentUser) // [success] 2.1 การดึงข้อมูลผู้ใช้ปัจจุบัน [Y]

	// เส้นทางอื่นๆ
	userRoutes.Get("/:userId", userHandler.GetProfile) // [success] 2.2 การดูโปรไฟล์ผู้ใช้ [Y]

	// แก้ไขโปรไฟล์
	userRoutes.Patch("/:userId", userHandler.UpdateProfile)                  // [success] 2.3 การอัปเดตโปรไฟล์ [Y]
	userRoutes.Put("/:userId/profile-image", userHandler.UploadProfileImage) // [success] 2.4 การอัปโหลดรูปโปรไฟล์ [Y]
}
