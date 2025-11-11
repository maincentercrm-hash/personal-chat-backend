// interfaces/api/handler/broadcast_delivery_handler.go
package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
	"github.com/thizplus/gofiber-chat-api/pkg/utils"
)

// BroadcastDeliveryHandler จัดการ HTTP request/response สำหรับ broadcast delivery
type BroadcastDeliveryHandler struct {
	broadcastDeliveryService service.BroadcastDeliveryService
}

// NewBroadcastDeliveryHandler สร้าง instance ใหม่ของ BroadcastDeliveryHandler
func NewBroadcastDeliveryHandler(broadcastDeliveryService service.BroadcastDeliveryService) *BroadcastDeliveryHandler {
	return &BroadcastDeliveryHandler{
		broadcastDeliveryService: broadcastDeliveryService,
	}
}

// TrackOpen บันทึกการเปิดอ่าน broadcast
func (h *BroadcastDeliveryHandler) TrackOpen(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง deliveryID จาก params
	deliveryID, err := utils.ParseUUIDParam(c, "deliveryId")
	if err != nil {
		return err
	}

	// ดึงข้อมูล delivery เพื่อตรวจสอบสิทธิ์
	delivery, err := h.broadcastDeliveryService.GetByID(deliveryID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Delivery not found: " + err.Error(),
		})
	}

	// ตรวจสอบว่า userID ตรงกับ user_id ใน delivery หรือไม่
	if delivery.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "You don't have permission to track this delivery",
		})
	}

	// บันทึกการเปิดอ่าน
	err = h.broadcastDeliveryService.MarkAsOpened(deliveryID, time.Now())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error tracking open: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Open tracked successfully",
	})
}

// TrackClick บันทึกการคลิก broadcast
func (h *BroadcastDeliveryHandler) TrackClick(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง deliveryID จาก params
	deliveryID, err := utils.ParseUUIDParam(c, "deliveryId")
	if err != nil {
		return err
	}

	// ดึงข้อมูล delivery เพื่อตรวจสอบสิทธิ์
	delivery, err := h.broadcastDeliveryService.GetByID(deliveryID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Delivery not found: " + err.Error(),
		})
	}

	// ตรวจสอบว่า userID ตรงกับ user_id ใน delivery หรือไม่
	if delivery.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "You don't have permission to track this delivery",
		})
	}

	// บันทึกการคลิก
	err = h.broadcastDeliveryService.MarkAsClicked(deliveryID, time.Now())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error tracking click: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Click tracked successfully",
	})
}

// GetUserDeliveries ดึงรายการ broadcast deliveries ของผู้ใช้
func (h *BroadcastDeliveryHandler) GetUserDeliveries(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// รับพารามิเตอร์การแบ่งหน้า
	limit := c.QueryInt("limit", 10)
	page := c.QueryInt("page", 1)
	offset := (page - 1) * limit

	// ดึงรายการ deliveries
	deliveries, total, err := h.broadcastDeliveryService.GetByUserID(userID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error fetching deliveries: " + err.Error(),
		})
	}

	// สร้างข้อมูลการแบ่งหน้า
	pagination := map[string]interface{}{
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
	}

	return c.JSON(fiber.Map{
		"success":    true,
		"message":    "Deliveries retrieved successfully",
		"deliveries": deliveries,
		"pagination": pagination,
	})
}
