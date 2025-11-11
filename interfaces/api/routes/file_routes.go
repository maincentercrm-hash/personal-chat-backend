// interfaces/api/routes/file_routes.go
package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/handler"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
)

// SetupFileRoutes ตั้งค่าเส้นทางสำหรับการจัดการไฟล์
func SetupFileRoutes(router fiber.Router, fileHandler *handler.FileHandler) {
	// กำหนดเส้นทางการอัปโหลดไฟล์ (ต้องมีการยืนยันตัวตน)
	upload := router.Group("/upload")
	upload.Use(middleware.Protected())

	// กำหนดเส้นทาง upload
	upload.Post("/image", fileHandler.UploadImage) // [success] 3.1 การอัปโหลดรูปภาพ [Y]
	upload.Post("/file", fileHandler.UploadFile)   // [success] 3.2 การอัปโหลดไฟล์ [Y]
	// ถ้ามีเพิ่มเติม เช่น
	// upload.Post("/avatar", fileHandler.UploadAvatar)
}
