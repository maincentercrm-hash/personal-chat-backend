// interfaces/api/routes/search_routes.go
package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/handler"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
)

// SetupSearchRoutes กำหนดเส้นทาง API สำหรับการค้นหา
func SetupSearchRoutes(router fiber.Router, searchHandler *handler.SearchHandler) {
	// สร้างกลุ่มเส้นทางค้นหา
	search := router.Group("/search")
	search.Use(middleware.Protected())

	// ค้นหาทั้งผู้ใช้และธุรกิจ
	search.Get("/", searchHandler.SearchAll)
}
