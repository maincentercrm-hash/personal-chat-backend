// interfaces/api/routes/auth_routes.go
package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/handler"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
)

// SetupAuthRoutes กำหนดเส้นทางสำหรับการยืนยันตัวตน
func SetupAuthRoutes(router fiber.Router, authHandler *handler.AuthHandler) {
	// เส้นทางที่ไม่ต้องการการยืนยันตัวตน
	authRoutes := router.Group("/auth")
	authRoutes.Post("/register", authHandler.Register)          // [success] 1.1 การลงทะเบียนสร้างผู้ใช้ใหม่ [Y]
	authRoutes.Post("/login", authHandler.Login)                // [success] 1.2 การเข้าสู่ระบบ [Y]
	authRoutes.Post("/refresh-token", authHandler.RefreshToken) // [success] 1.4 การต่ออายุ Token [Y]
	// authRoutes.Post("/reset-password", authHandler.ResetPassword) // [pendding]

	// เส้นทางที่ต้องการการยืนยันตัวตน
	authRoutes.Get("/user", middleware.Protected(), authHandler.GetCurrentUser) // [success] 1.3 การดึงข้อมูลผู้ใช้ปัจจุบัน [Y]
	authRoutes.Post("/logout", middleware.Protected(), authHandler.Logout)      // [success] 1.5 การออกจากระบบ [Y]
}
