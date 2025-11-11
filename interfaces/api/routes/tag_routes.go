// interfaces/api/routes/tag_routes.go
package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/handler"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
)

// SetupTagRoutes ตั้งค่า routes สำหรับ Tag Management
func SetupTagRoutes(
	router fiber.Router,
	tagHandler *handler.TagHandler,
	businessAdminService service.BusinessAdminService,
) {
	// สร้างกลุ่มเส้นทางธุรกิจ
	businesses := router.Group("/businesses")
	businesses.Use(middleware.Protected())

	// Tag Management Routes (Admin/Owner only)
	tags := businesses.Group("/:businessId/tags")

	// ใช้ middleware ตรวจสอบสิทธิ์แอดมิน
	tags.Use(middleware.CheckBusinessAdmin(businessAdminService))

	// ----- จัดการแท็กหลัก -----
	// สร้างแท็กใหม่
	tags.Post("/", tagHandler.CreateTag) // [success] 16.1 การสร้างแท็ก [Y]

	// ดึงแท็กทั้งหมดของธุรกิจ
	tags.Get("/", tagHandler.GetBusinessTags) // [success] 16.2 การดึงแท็กทั้งหมดของธุรกิจ [Y]

	// อัปเดตแท็ก
	tags.Patch("/:tagId", tagHandler.UpdateTag) // [success] 16.3 การอัปเดตแท็ก [Y]

	// ลบแท็ก (owner/admin เท่านั้น)
	tags.Delete("/:tagId", tagHandler.DeleteTag) // [success] 16.11 การลบแท็ก [Y]

	// ----- จัดการความสัมพันธ์ระหว่างแท็กกับผู้ใช้ (แบบที่ 1) -----
	// Base path: /api/v1/businesses/:businessId/tags/:tagId/users
	tagUsers := tags.Group("/:tagId/users")

	// เส้นทางคงที่ต้องมาก่อน
	// ดึงรายชื่อผู้ใช้ที่มีแท็กนี้
	tagUsers.Get("/", tagHandler.GetUsersByTag) // [success] 16.5 การดึงรายชื่อผู้ใช้ที่มีแท็กนี้ [Y]

	// เพิ่มแท็กให้ผู้ใช้หลายคนพร้อมกัน
	tagUsers.Post("/bulk", tagHandler.BulkAddTagToUsers) // [success] 16.7 การเพิ่มแท็กให้ผู้ใช้หลายคนพร้อมกัน [Y]

	// เส้นทางที่มีพารามิเตอร์มาทีหลัง
	// เพิ่มแท็กให้ผู้ใช้
	tagUsers.Post("/:userId", tagHandler.AddTagToUser) // [success] 16.4 การเพิ่มแท็กให้ผู้ใช้ [Y]

	// ลบแท็กจากผู้ใช้
	tagUsers.Delete("/:userId", tagHandler.RemoveTagFromUser) // [success] 16.9 การลบแท็กจากผู้ใช้ [Y]

	// ----- จัดการความสัมพันธ์ระหว่างแท็กกับผู้ใช้ (แบบที่ 2) -----
	// Base path: /api/v1/businesses/:businessId/users/:userId/tags
	userTags := businesses.Group("/:businessId/users/:userId/tags")

	// ใช้ middleware ตรวจสอบสิทธิ์แอดมิน
	userTags.Use(middleware.CheckBusinessAdmin(businessAdminService))

	// ดึงแท็กทั้งหมดของผู้ใช้
	userTags.Get("/", tagHandler.GetUserTags) // [success] 16.6 การดึงแท็กทั้งหมดของผู้ใช้ [Y]

	// เพิ่มแท็กให้ผู้ใช้ (มีผลเหมือนกับ tagUsers.Post("/:userId"))
	userTags.Post("/:tagId", tagHandler.AddTagToUser) // [success] 16.8 การเพิ่มแท็กให้ผู้ใช้ [Y] PATH 2

	// ลบแท็กจากผู้ใช้ (มีผลเหมือนกับ tagUsers.Delete("/:userId"))
	userTags.Delete("/:tagId", tagHandler.RemoveTagFromUser) // [success] 16.10 การลบแท็กจากผู้ใช้ [Y]

	// ----- ส่วนขยายในอนาคต -----
	// Tag Analytics Routes (Optional - สำหรับอนาคต)
	// analytics := tags.Group("/analytics")
	// analytics.Get("/popular", tagHandler.GetPopularTags)
	// analytics.Get("/unused", tagHandler.GetUnusedTags)
	// analytics.Get("/combinations", tagHandler.GetTagCombinations)

	// Tag Search Routes
	// search := tags.Group("/search")
	// search.Get("/", tagHandler.SearchTags)
	// search.Post("/advanced", tagHandler.AdvancedSearchTags)
}
