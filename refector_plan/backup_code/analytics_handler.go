// interfaces/api/handler/analytics_handler.go
package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
	"github.com/thizplus/gofiber-chat-api/pkg/utils"
)

// AnalyticsHandler จัดการ HTTP request/response สำหรับข้อมูลวิเคราะห์
type AnalyticsHandler struct {
	analyticsService service.AnalyticsService
}

// NewAnalyticsHandler สร้าง instance ใหม่ของ AnalyticsHandler
func NewAnalyticsHandler(analyticsService service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
	}
}

// GetDailyAnalytics ดึงข้อมูลวิเคราะห์รายวัน
func (h *AnalyticsHandler) GetDailyAnalytics(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จากพารามิเตอร์
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err
	}

	// ดึงพารามิเตอร์วันที่
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// กำหนดค่าเริ่มต้นถ้าไม่ได้ระบุ
	startDate := time.Now().AddDate(0, 0, -30)
	endDate := time.Now()

	// แปลงค่าวันที่ถ้ามีการระบุ
	if startDateStr != "" {
		parsedStartDate, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			startDate = parsedStartDate
		}
	}

	if endDateStr != "" {
		parsedEndDate, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			endDate = parsedEndDate
		}
	}

	// ตรวจสอบว่าช่วงวันที่ถูกต้อง
	if startDate.After(endDate) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Start date must be before end date",
		})
	}

	// ดึงข้อมูลวิเคราะห์
	analytics, err := h.analyticsService.GetDailyAnalytics(businessID, userID, startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error getting analytics data: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"analytics":  analytics,
			"start_date": startDate.Format("2006-01-02"),
			"end_date":   endDate.Format("2006-01-02"),
		},
	})
}

// GetSummaryAnalytics สรุปข้อมูลวิเคราะห์
func (h *AnalyticsHandler) GetSummaryAnalytics(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จากพารามิเตอร์
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err
	}

	// ดึงจำนวนวันจาก query param
	days := utils.ParseInt(c.Query("days"), 30)

	// ดึงข้อมูลสรุป
	summary, err := h.analyticsService.GetSummaryAnalytics(businessID, userID, days)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error getting analytics summary: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    summary,
	})
}

// TrackEvent บันทึกเหตุการณ์สำหรับการวิเคราะห์
func (h *AnalyticsHandler) TrackEvent(c *fiber.Ctx) error {
	// ดึง businessID จากพารามิเตอร์
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err
	}

	// ดึงข้อมูลเหตุการณ์
	var input struct {
		EventType string    `json:"event_type"`
		UserID    uuid.UUID `json:"user_id,omitempty"`
		Count     int       `json:"count,omitempty"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data",
		})
	}

	// ตรวจสอบประเภทเหตุการณ์และบันทึกข้อมูล
	switch input.EventType {
	case "new_follower":
		err = h.analyticsService.TrackNewFollower(businessID)
	case "unfollow":
		err = h.analyticsService.TrackUnfollow(businessID)
	case "message_received":
		err = h.analyticsService.TrackMessageReceived(businessID, input.Count)
	case "message_sent":
		err = h.analyticsService.TrackMessageSent(businessID, input.Count)
	case "active_user":
		err = h.analyticsService.TrackActiveUser(businessID, input.UserID)
	case "broadcast_open":
		err = h.analyticsService.TrackBroadcastOpen(businessID)
	case "broadcast_click":
		err = h.analyticsService.TrackBroadcastClick(businessID)
	case "rich_menu_click":
		err = h.analyticsService.TrackRichMenuClick(businessID)
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid event type",
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error tracking event: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Event tracked successfully",
	})
}
