// interfaces/api/routes/broadcast_delivery_routes.go

package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/handler"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
)

// SetupBroadcastDeliveryRoutes กำหนดเส้นทาง API สำหรับจัดการ broadcast deliveries
func SetupBroadcastDeliveryRoutes(router fiber.Router, broadcastDeliveryHandler *handler.BroadcastDeliveryHandler) {
	// สร้างกลุ่มเส้นทางสำหรับ broadcast deliveries
	deliveries := router.Group("/broadcast-deliveries")
	deliveries.Use(middleware.Protected())

	// เส้นทาง broadcast deliveries ของผู้ใช้
	deliveries.Get("/", broadcastDeliveryHandler.GetUserDeliveries) // [success] 21.1 ดึงรายการ Broadcast Deliveries ของผู้ใช้ [Y]

	// เส้นทางสำหรับการติดตาม (tracking)
	deliveries.Post("/:deliveryId/track-open", broadcastDeliveryHandler.TrackOpen)   // [pending]
	deliveries.Post("/:deliveryId/track-click", broadcastDeliveryHandler.TrackClick) // [pending]
}
