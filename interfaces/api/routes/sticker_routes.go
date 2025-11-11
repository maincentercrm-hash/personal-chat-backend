// interfaces/api/routes/sticker_routes.go
package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/handler"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
)

// SetupStickerRoutes กำหนดเส้นทาง API สำหรับสติกเกอร์
func SetupStickerRoutes(router fiber.Router, stickerHandler *handler.StickerHandler) {
	// กลุ่มเส้นทางสำหรับผู้ดูแลระบบ
	adminStickers := router.Group("/admin/stickers")
	adminStickers.Use(middleware.Protected()) // ต้องล็อกอินก่อน
	// TODO: เพิ่ม middleware สำหรับตรวจสอบสิทธิ์แอดมิน

	// การจัดการชุดสติกเกอร์ (สำหรับแอดมิน)
	adminStickers.Post("/sets", stickerHandler.CreateStickerSet)                         // [success] 18.1.1 การสร้างชุดสติกเกอร์ใหม่ [Y]
	adminStickers.Patch("/sets/:stickerSetId", stickerHandler.UpdateStickerSet)          // [success] 18.1.2 การอัปเดตข้อมูลชุดสติกเกอร์ [Y]
	adminStickers.Delete("/sets/:stickerSetId", stickerHandler.DeleteStickerSet)         // [success] 18.1.4 การลบชุดสติกเกอร์ [Y]
	adminStickers.Put("/sets/:stickerSetId/cover", stickerHandler.UploadStickerSetCover) // [success] 18.1.3 การอัปโหลดรูปปกชุดสติกเกอร์ [Y]

	// การจัดการสติกเกอร์ (สำหรับแอดมิน)
	adminStickers.Post("/sets/:stickerSetId/stickers", stickerHandler.AddStickerToSet) // [success] 18.2.1 การเพิ่มสติกเกอร์ใหม่ลงในชุด [Y]
	adminStickers.Patch("/stickers/:stickerId", stickerHandler.UpdateSticker)          // [success] 18.2.2 การอัปเดตข้อมูลสติกเกอร์ [Y]
	adminStickers.Delete("/stickers/:stickerId", stickerHandler.DeleteSticker)         // [success] 18.2.3 การลบสติกเกอร์ [Y]

	// กลุ่มเส้นทางสำหรับผู้ใช้ทั่วไป
	stickers := router.Group("/stickers")
	stickers.Use(middleware.Protected()) // ต้องล็อกอินก่อน

	// ดูข้อมูลชุดสติกเกอร์และสติกเกอร์
	stickers.Get("/sets", stickerHandler.GetAllStickerSets)             // [success] 18.3.1 การดึงข้อมูลชุดสติกเกอร์ทั้งหมด [Y]
	stickers.Get("/sets/default", stickerHandler.GetDefaultStickerSets) // [success] 18.3.2 การดึงข้อมูลชุดสติกเกอร์เริ่มต้น [Y]
	stickers.Get("/sets/:stickerSetId", stickerHandler.GetStickerSet)   // [success] 18.3.3 การดึงข้อมูลชุดสติกเกอร์และสติกเกอร์ในชุด [Y]

	// จัดการชุดสติกเกอร์ของผู้ใช้
	stickers.Get("/my-sets", stickerHandler.GetUserStickerSets)                            // [success] 18.3.4 การดึงชุดสติกเกอร์ของผู้ใช้ [Y]
	stickers.Post("/sets/:stickerSetId/add", stickerHandler.AddStickerSetToUser)           // [success] 18.3.5 การเพิ่มชุดสติกเกอร์ให้ผู้ใช้ [Y]
	stickers.Delete("/sets/:stickerSetId/remove", stickerHandler.RemoveStickerSetFromUser) // [success] 18.3.6 การลบชุดสติกเกอร์ออกจากผู้ใช้ [Y]
	stickers.Put("/sets/:stickerSetId/favorite", stickerHandler.SetStickerSetAsFavorite)   // [success] 18.3.7 การตั้งค่าชุดสติกเกอร์เป็นรายการโปรด [Y]

	// จัดการการใช้งานสติกเกอร์
	stickers.Post("/stickers/:stickerId/usage", stickerHandler.RecordStickerUsage) // [success] 18.4.1 การบันทึกการใช้งานสติกเกอร์ [Y]
	stickers.Get("/recent", stickerHandler.GetUserRecentStickers)                  // [success] 18.4.2 การดึงสติกเกอร์ที่ใช้ล่าสุด [Y]
	stickers.Get("/favorites", stickerHandler.GetUserFavoriteStickers)             // [pending]
	// TODO: อาจเพิ่ม endpoint สำหรับการค้นหาชุดสติกเกอร์ในอนาคต (/sets/search)
}
