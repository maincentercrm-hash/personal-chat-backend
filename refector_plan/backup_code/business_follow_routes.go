// interfaces/api/routes/business_follow_routes.go
package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/handler"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
)

// SetupBusinessFollowRoutes กำหนดเส้นทาง API สำหรับการติดตามธุรกิจ
func SetupBusinessFollowRoutes(router fiber.Router, businessFollowHandler *handler.BusinessFollowHandler) {
	// สร้างกลุ่มเส้นทางธุรกิจ
	businesses := router.Group("/businesses")
	businesses.Use(middleware.Protected())

	// เส้นทางจัดการการติดตามธุรกิจ
	businesses.Post("/:businessId/follow", businessFollowHandler.FollowBusiness)          // [success] 7.1 การติดตามธุรกิจ [Y]
	businesses.Delete("/:businessId/follow", businessFollowHandler.UnfollowBusiness)      // [success] 7.6 การเลิกติดตามธุรกิจ [Y]
	businesses.Get("/:businessId/followers", businessFollowHandler.GetBusinessFollowers)  // [success] 7.3 การดึงรายชื่อผู้ติดตามของธุรกิจ [Y]
	businesses.Get("/:businessId/follow/status", businessFollowHandler.CheckFollowStatus) // [success] 7.2 การตรวจสอบสถานะการติดตาม [Y]

	// เส้นทางสำหรับผู้ใช้และธุรกิจที่ติดตาม
	users := router.Group("/users")
	users.Use(middleware.Protected())
	users.Get("/followed-businesses", businessFollowHandler.GetUserFollowedBusinesses)         // [success] 7.4 การดึงรายชื่อธุรกิจที่ติดตาม [Y]
	users.Get("/:userId/followed-businesses", businessFollowHandler.GetUserFollowedBusinesses) // [success] 7.5 การดึงรายชื่อธุรกิจที่ผู้ใช้เฉพาะเจาะจงติดตาม [Y]
}
