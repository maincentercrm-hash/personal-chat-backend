// interfaces/api/routes/user_tag_routes.go
package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/handler"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
)

// SetupUserTagRoutes ตั้งค่า routes สำหรับ UserTag Management (Advanced)
func SetupUserTagRoutes(
	router fiber.Router,
	userTagHandler *handler.UserTagHandler,
	businessAdminService service.BusinessAdminService,
) {
	// สร้างกลุ่มเส้นทางธุรกิจ
	businesses := router.Group("/businesses")
	businesses.Use(middleware.Protected())

	// กลุ่ม routes สำหรับ user-tags (Advanced Management)
	// Base path: /api/v1/businesses/:businessId/user-tags
	userTags := businesses.Group("/:businessId/user-tags")

	// ใช้ middleware ตรวจสอบสิทธิ์แอดมิน
	userTags.Use(middleware.CheckBusinessAdmin(businessAdminService))

	// =========== 1. กลุ่มเส้นทางเฉพาะ (Fixed paths) ===========

	// กลุ่มตรวจสอบผู้ใช้ที่มีหลาย Tag
	userTags.Get("/multiple", userTagHandler.GetUsersWithMultipleTags) // [success] 17.10 การดึงผู้ใช้ที่มีหลายแท็ก [Y]

	// Bulk Operations Routes
	bulk := userTags.Group("/bulk")
	bulk.Post("/tag/:tagId/users", userTagHandler.BulkAddTagToUsers)        // [success] 17.7 การเพิ่มแท็กให้ผู้ใช้หลายคน (Advanced) [Y]
	bulk.Delete("/tag/:tagId/users", userTagHandler.BulkRemoveTagFromUsers) // [success] 17.8 การลบแท็กจากผู้ใช้หลายคน [Y]

	// Advanced Search Routes
	search := userTags.Group("/search")
	search.Post("/", userTagHandler.SearchUsersByTags) // [success] 17.9 การค้นหาผู้ใช้ตามเงื่อนไขซับซ้อน [Y]

	// Statistics & Analytics Routes
	stats := userTags.Group("/statistics")
	stats.Get("/", userTagHandler.GetTagStatistics) // [success] 17.11 การดึงสถิติการใช้แท็ก [Y]

	// การวิเคราะห์ขั้นสูง
	userTags.Get("/analytics", userTagHandler.GetUserTagAnalytics) // [success] 17.12 การดึงข้อมูลการใช้แท็กของผู้ใช้ [Y]

	// Data Export Routes
	export := userTags.Group("/export")
	export.Get("/", userTagHandler.ExportUserTags) // [success] 17.13 การส่งออกข้อมูลแท็กของผู้ใช้ CSV [Y]

	// Activity & Monitoring Routes
	activity := userTags.Group("/activity")
	activity.Get("/", userTagHandler.GetRecentTagActivity) // [pendding]

	// Validation Routes
	validate := userTags.Group("/validate") // [success] 17.15 การตรวจสอบความถูกต้องของการแท็ก [Y]
	validate.Post("/", userTagHandler.ValidateTagging)

	// =========== 2. เส้นทางที่ใช้ Tag ID ===========

	// ดึงผู้ใช้ที่มีแท็กนี้ (with pagination)
	userTags.Get("/tag/:tagId/users", userTagHandler.GetUsersByTag) // [success] 17.4 การดึงรายชื่อผู้ใช้ที่มีแท็กนี้ [Y]

	// =========== 3. เส้นทางที่ใช้ User ID ===========

	// ดึงแท็กทั้งหมดของผู้ใช้
	userTags.Get("/:userId", userTagHandler.GetUserTags) // [success] 17.3 การดึงแท็กทั้งหมดของผู้ใช้ [Y]

	// แทนที่แท็กทั้งหมดของผู้ใช้ด้วยแท็กใหม่
	userTags.Put("/:userId/tags", userTagHandler.ReplaceUserTags) // [success] 17.6 การแทนที่แท็กทั้งหมดของผู้ใช้ด้วยแท็กใหม่ [Y]

	// ตรวจสอบว่าผู้ใช้มีแท็กนี้หรือไม่
	userTags.Get("/:userId/tags/:tagId/check", userTagHandler.CheckUserHasTag) // [success] 17.5 การตรวจสอบว่าผู้ใช้มีแท็กนี้หรือไม่ [Y]

	// เพิ่มแท็กให้ผู้ใช้
	userTags.Post("/:userId/tags/:tagId", userTagHandler.AddTagToUser) // [success] 17.1 การเพิ่มแท็กให้ผู้ใช้ (Advanced)

	// ลบแท็กจากผู้ใช้
	userTags.Delete("/:userId/tags/:tagId", userTagHandler.RemoveTagFromUser) // [success] 17.2 การลบแท็กจากผู้ใช้ (Advanced)

	// =========== 4. เส้นทางที่ถูก Comment ไว้ (Future Features) ===========

	// Advanced Analytics Routes (Future Features)
	// analytics := userTags.Group("/analytics")
	// แนวโน้มการใช้แท็กตามเวลา
	// analytics.Get("/trends", userTagHandler.GetTagTrends)
	// การกระจายจำนวนแท็กของผู้ใช้
	// analytics.Get("/distribution", userTagHandler.GetTagDistribution)
	// การจับคู่แท็กที่เกิดขึ้นบ่อย
	// analytics.Get("/combinations", userTagHandler.GetTagCombinations)
	// สถิติการใช้แท็กตาม Admin
	// analytics.Get("/admin-usage", userTagHandler.GetAdminTagUsage)

	// Management Routes (Admin Tools)
	// manage := userTags.Group("/manage")
	// ผู้ใช้ที่ยังไม่มีแท็ก
	// manage.Get("/untagged", userTagHandler.GetUntaggedUsers)
	// ทำความสะอาดแท็กที่ไม่ใช้แล้ว
	// manage.Post("/cleanup", userTagHandler.CleanupUnusedTags)
	// รวมแท็กที่คล้ายกัน
	// manage.Post("/merge", userTagHandler.MergeTags)
}
