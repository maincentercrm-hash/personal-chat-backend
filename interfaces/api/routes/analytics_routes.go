// interfaces/api/routes/analytics_routes.go
package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/handler"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
)

// SetupAnalyticsRoutes กำหนดเส้นทาง API สำหรับข้อมูลวิเคราะห์
func SetupAnalyticsRoutes(
	router fiber.Router,
	analyticsHandler *handler.AnalyticsHandler,
	businessAdminService service.BusinessAdminService,
) {
	// สร้างกลุ่มเส้นทางธุรกิจ
	businesses := router.Group("/businesses")
	businesses.Use(middleware.Protected())

	// สร้างกลุ่มเส้นทางสำหรับข้อมูลวิเคราะห์
	analytics := businesses.Group("/:businessId/analytics")

	// ใช้ middleware ตรวจสอบสิทธิ์แอดมิน
	analytics.Use(middleware.CheckBusinessAdmin(businessAdminService))

	// ดึงข้อมูลวิเคราะห์
	analytics.Get("/daily", analyticsHandler.GetDailyAnalytics)     // [success] 14.1 การดึงข้อมูลวิเคราะห์รายวัน [Y]
	analytics.Get("/summary", analyticsHandler.GetSummaryAnalytics) // [success] 14.2 การดึงข้อมูลสรุป [Y]

	// บันทึกเหตุการณ์สำหรับการวิเคราะห์
	// ใช้ API key แทนการยืนยันตัวตนปกติ
	apiTrack := router.Group("/businesses/:businessId/track")
	apiTrack.Use(middleware.VerifyAPIKey())              // ต้องสร้าง middleware นี้เพิ่มเติม
	apiTrack.Post("/event", analyticsHandler.TrackEvent) // [pending] TODO: ตัวจัดเก็บข้อมูล event คล้ายๆ google tag manager หรือ facebook pixel
}
