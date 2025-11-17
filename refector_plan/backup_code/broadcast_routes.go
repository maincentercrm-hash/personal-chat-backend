// interfaces/api/routes/broadcast_routes.go

package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/handler"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
)

// SetupBroadcastRoutes กำหนดเส้นทาง API สำหรับจัดการ broadcasts
func SetupBroadcastRoutes(router fiber.Router, broadcastHandler *handler.BroadcastHandler, businessAdminService service.BusinessAdminService) {
	// สร้างกลุ่มเส้นทางธุรกิจ
	businesses := router.Group("/businesses")
	businesses.Use(middleware.Protected())

	// สร้างกลุ่ม broadcasts สำหรับแต่ละธุรกิจ
	broadcasts := businesses.Group("/:businessId/broadcasts")

	// ตรวจสอบสิทธิ์แอดมินสำหรับทุกเส้นทางใน broadcasts group
	broadcasts.Use(middleware.CheckBusinessAdmin(businessAdminService))

	// เส้นทาง broadcasts
	broadcasts.Get("/", broadcastHandler.GetBroadcasts) // [success] 20.7 ดึงรายการ Broadcasts [Y]

	/*
		20.1 การสร้าง Broadcast [Y]
		20.2 การสร้าง Broadcast พร้อมรูปภาพ [Y]
		20.3 การสร้าง Broadcast แบบ Flex/Carousel [Y]
		20.4 การสร้าง Broadcast ส่งเฉพาะกลุ่ม (Tags) [Y]
		20.5 การสร้าง Broadcast ส่งเฉพาะบุคคล [Y]
		20.6 การสร้าง Broadcast ตามเงื่อนไข Customer Profile [Y]
	*/
	broadcasts.Post("/", broadcastHandler.CreateBroadcast) // [success]

	broadcasts.Post("/estimate-target", broadcastHandler.GetEstimatedTargetCount)       // [success] 20.17 ประมาณจำนวนผู้รับ [Y]
	broadcasts.Get("/:broadcastId", broadcastHandler.GetBroadcast)                      // [success] 20.8 ดึงรายละเอียด Broadcast [Y]
	broadcasts.Put("/:broadcastId", broadcastHandler.UpdateBroadcast)                   // [success] 20.9 อัปเดต Broadcast [Y]
	broadcasts.Delete("/:broadcastId", broadcastHandler.DeleteBroadcast)                // [success] 20.10 ลบ Broadcast [Y]
	broadcasts.Post("/:broadcastId/send", broadcastHandler.SendBroadcast)               // [success] 20.11 ส่ง Broadcast ทันที [Y]
	broadcasts.Post("/:broadcastId/cancel", broadcastHandler.CancelBroadcast)           // [success] 20.12 ยกเลิก Broadcast ที่ตั้งเวลาไว้ [Y]
	broadcasts.Get("/:broadcastId/stats", broadcastHandler.GetBroadcastStats)           // [success] 20.13 ดูสถิติ Broadcast [Y]
	broadcasts.Get("/:broadcastId/deliveries", broadcastHandler.GetBroadcastDeliveries) // [success] 20.14 ดูรายการส่ง Broadcast [Y]
	broadcasts.Get("/:broadcastId/preview", broadcastHandler.PreviewBroadcast)          // [success] 20.15 ดูตัวอย่าง Broadcast [Y]
	broadcasts.Post("/:broadcastId/duplicate", broadcastHandler.DuplicateBroadcast)     // [success] 20.16 คัดลอก Broadcast [Y]

}
