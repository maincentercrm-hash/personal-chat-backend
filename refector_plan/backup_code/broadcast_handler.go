// interfaces/api/handler/broadcast_handler.go
package handler

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/dto"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/domain/types"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
	"github.com/thizplus/gofiber-chat-api/pkg/utils"
	"github.com/thizplus/gofiber-chat-api/scheduler"
)

// BroadcastHandler จัดการ HTTP request/response สำหรับการส่ง broadcast
type BroadcastHandler struct {
	broadcastService service.BroadcastService
	scheduler        scheduler.BroadcastSchedulerInterface // แก้เป็นแบบนี้
}

// NewBroadcastHandler สร้าง instance ใหม่ของ BroadcastHandler
func NewBroadcastHandler(broadcastService service.BroadcastService, scheduler scheduler.BroadcastSchedulerInterface) *BroadcastHandler {
	return &BroadcastHandler{
		broadcastService: broadcastService,
		scheduler:        scheduler,
	}
}

// CreateBroadcast สร้าง broadcast ใหม่
func (h *BroadcastHandler) CreateBroadcast(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก params
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// รับข้อมูล broadcast
	var input struct {
		Title       string                 `json:"title"`
		MessageType string                 `json:"message_type"`
		Content     string                 `json:"content"`
		MediaURL    string                 `json:"media_url"`
		BubbleType  string                 `json:"bubble_type"`
		BubbleData  map[string]interface{} `json:"bubble_data"`
		TargetType  string                 `json:"target_type"`
		TargetData  map[string]interface{} `json:"target_data"`
		ScheduleAt  *time.Time             `json:"schedule_at"`
		SendNow     bool                   `json:"send_now"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
	}

	// ตรวจสอบข้อมูลที่จำเป็น
	if input.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Title is required",
		})
	}

	if input.MessageType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Message type is required",
		})
	}

	// แปลง BubbleData เป็น types.JSONB
	var bubbleData types.JSONB
	if input.BubbleData != nil {
		bubbleData = types.JSONB(input.BubbleData)
	}

	// ตรวจสอบความถูกต้องของเนื้อหา
	err = h.broadcastService.ValidateBroadcastContent(
		input.MessageType,
		input.Content,
		input.MediaURL,
		input.BubbleType,
		bubbleData,
	)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid broadcast content: " + err.Error(),
		})
	}

	// สร้าง broadcast
	broadcast, err := h.broadcastService.CreateBroadcast(
		businessID,
		userID,
		input.Title,
		input.MessageType,
		input.Content,
		input.MediaURL,
		input.BubbleType,
		bubbleData,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error creating broadcast: " + err.Error(),
		})
	}

	// กำหนดเป้าหมายตาม targetType
	if input.TargetType != "" {
		switch input.TargetType {
		case "all":
			err = h.broadcastService.SetTargetAll(broadcast.ID, businessID, userID)
		case "tags":
			// แปลงข้อมูล tags
			includeTags := []uuid.UUID{}
			excludeTags := []uuid.UUID{}
			matchType := "all"

			if input.TargetData != nil {
				if includeTagsData, ok := input.TargetData["include_tags"].([]interface{}); ok {
					for _, tag := range includeTagsData {
						if tagStr, ok := tag.(string); ok {
							if tagID, err := uuid.Parse(tagStr); err == nil {
								includeTags = append(includeTags, tagID)
							}
						}
					}
				}

				if excludeTagsData, ok := input.TargetData["exclude_tags"].([]interface{}); ok {
					for _, tag := range excludeTagsData {
						if tagStr, ok := tag.(string); ok {
							if tagID, err := uuid.Parse(tagStr); err == nil {
								excludeTags = append(excludeTags, tagID)
							}
						}
					}
				}

				if matchTypeData, ok := input.TargetData["match_type"].(string); ok && (matchTypeData == "all" || matchTypeData == "any") {
					matchType = matchTypeData
				}
			}

			err = h.broadcastService.SetTargetTags(broadcast.ID, businessID, userID, includeTags, excludeTags, matchType)
		case "specific_users":
			// แปลงข้อมูล users
			userIDs := []uuid.UUID{}

			if input.TargetData != nil {
				if usersData, ok := input.TargetData["user_ids"].([]interface{}); ok {
					for _, user := range usersData {
						if userStr, ok := user.(string); ok {
							if userID, err := uuid.Parse(userStr); err == nil {
								userIDs = append(userIDs, userID)
							}
						}
					}
				}
			}

			err = h.broadcastService.SetTargetUsers(broadcast.ID, businessID, userID, userIDs)
		case "customer_profile":
			// แปลงข้อมูลเป็น BroadcastTargetCriteria
			criteria := &dto.BroadcastTargetCriteria{}

			if input.TargetData != nil {
				// ตัวอย่างการแปลง CustomerTypes
				if customerTypesData, ok := input.TargetData["customer_types"].([]interface{}); ok {
					customerTypes := []string{}
					for _, cType := range customerTypesData {
						if cTypeStr, ok := cType.(string); ok {
							customerTypes = append(customerTypes, cTypeStr)
						}
					}
					criteria.CustomerTypes = customerTypes
				}

				// แปลงข้อมูลอื่นๆ ตามต้องการ
				if statusesData, ok := input.TargetData["statuses"].([]interface{}); ok {
					statuses := []string{}
					for _, status := range statusesData {
						if statusStr, ok := status.(string); ok {
							statuses = append(statuses, statusStr)
						}
					}
					criteria.Statuses = statuses
				}

				if lastContactFromStr, ok := input.TargetData["last_contact_from"].(string); ok {
					if lastContactFrom, err := time.Parse(time.RFC3339, lastContactFromStr); err == nil {
						criteria.LastContactFrom = &lastContactFrom
					}
				}

				if lastContactToStr, ok := input.TargetData["last_contact_to"].(string); ok {
					if lastContactTo, err := time.Parse(time.RFC3339, lastContactToStr); err == nil {
						criteria.LastContactTo = &lastContactTo
					}
				}

				if customQuery, ok := input.TargetData["custom_query"].(map[string]interface{}); ok {
					criteria.CustomQuery = types.JSONB(customQuery)
				}
			}

			err = h.broadcastService.SetTargetCustomerProfile(broadcast.ID, businessID, userID, criteria)
		default:
			err = c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid target type",
			})
			return err
		}

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Error setting target: " + err.Error(),
			})
		}
	}

	// กำหนดเวลาส่ง (ถ้ามี)
	if input.ScheduleAt != nil && !input.SendNow {
		// เรียกใช้ service เพื่ออัปเดตในฐานข้อมูล
		err = h.broadcastService.ScheduleBroadcast(broadcast.ID, businessID, userID, *input.ScheduleAt)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Error scheduling broadcast: " + err.Error(),
			})
		}

		// ดึงข้อมูล broadcast ที่อัปเดตแล้ว
		broadcast, err = h.broadcastService.GetBroadcastByID(broadcast.ID, businessID, userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Error fetching updated broadcast: " + err.Error(),
			})
		}

		// เพิ่ม broadcast ลงใน scheduler
		if h.scheduler != nil { // เพิ่มการตรวจสอบ nil ตรงนี้
			err = h.scheduler.ScheduleBroadcast(broadcast)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"message": "Error adding to scheduler: " + err.Error(),
				})
			}
		} else {
			log.Printf("Warning: scheduler is nil, broadcast will not be scheduled")
		}
	}

	// ส่งทันที (ถ้าต้องการ)
	if input.SendNow {
		err = h.broadcastService.SendBroadcast(broadcast.ID, businessID, userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Error sending broadcast: " + err.Error(),
			})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success":   true,
		"message":   "Broadcast created successfully",
		"broadcast": broadcast,
	})
}

// GetBroadcasts ดึงรายการ broadcasts ของธุรกิจ
func (h *BroadcastHandler) GetBroadcasts(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก params
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// รับพารามิเตอร์การแบ่งหน้า
	limit := c.QueryInt("limit", 10)
	page := c.QueryInt("page", 1)
	offset := (page - 1) * limit

	// รับพารามิเตอร์การกรอง
	status := c.Query("status", "")

	// ดึงรายการ broadcasts
	broadcasts, total, err := h.broadcastService.GetBusinessBroadcasts(businessID, userID, status, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error fetching broadcasts: " + err.Error(),
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
		"message":    "Broadcasts retrieved successfully",
		"broadcasts": broadcasts,
		"pagination": pagination,
	})
}

// GetBroadcast ดึงรายละเอียดของ broadcast
func (h *BroadcastHandler) GetBroadcast(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก params
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// ดึง broadcastID จาก params
	broadcastID, err := utils.ParseUUIDParam(c, "broadcastId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// ดึงข้อมูล broadcast
	broadcast, err := h.broadcastService.GetBroadcastByID(broadcastID, businessID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Broadcast not found: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"message":   "Broadcast retrieved successfully",
		"broadcast": broadcast,
	})
}

// UpdateBroadcast อัปเดต broadcast
func (h *BroadcastHandler) UpdateBroadcast(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก params
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// ดึง broadcastID จาก params
	broadcastID, err := utils.ParseUUIDParam(c, "broadcastId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// รับข้อมูลที่ต้องการอัปเดต
	var input struct {
		Title      string                 `json:"title"`
		Content    string                 `json:"content"`
		MediaURL   string                 `json:"media_url"`
		BubbleType string                 `json:"bubble_type"`
		BubbleData map[string]interface{} `json:"bubble_data"`
		TargetType string                 `json:"target_type"`
		TargetData map[string]interface{} `json:"target_data"`
		ScheduleAt *time.Time             `json:"schedule_at"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
	}

	// สร้าง updateData
	updateData := types.JSONB{}

	if input.Title != "" {
		updateData["title"] = input.Title
	}

	if input.Content != "" {
		updateData["content"] = input.Content
	}

	if input.MediaURL != "" {
		updateData["media_url"] = input.MediaURL
	}

	if input.BubbleType != "" {
		updateData["bubble_type"] = input.BubbleType
	}

	if input.BubbleData != nil {
		updateData["bubble_data"] = input.BubbleData
	}

	// อัปเดต broadcast
	updatedBroadcast, err := h.broadcastService.UpdateBroadcast(broadcastID, businessID, userID, updateData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error updating broadcast: " + err.Error(),
		})
	}

	// อัปเดตเป้าหมาย (ถ้ามี)
	if input.TargetType != "" {
		switch input.TargetType {
		case "all":
			err = h.broadcastService.SetTargetAll(broadcastID, businessID, userID)
		case "tags":
			// แปลงข้อมูล tags
			includeTags := []uuid.UUID{}
			excludeTags := []uuid.UUID{}
			matchType := "all"

			if input.TargetData != nil {
				if includeTagsData, ok := input.TargetData["include_tags"].([]interface{}); ok {
					for _, tag := range includeTagsData {
						if tagStr, ok := tag.(string); ok {
							if tagID, err := uuid.Parse(tagStr); err == nil {
								includeTags = append(includeTags, tagID)
							}
						}
					}
				}

				if excludeTagsData, ok := input.TargetData["exclude_tags"].([]interface{}); ok {
					for _, tag := range excludeTagsData {
						if tagStr, ok := tag.(string); ok {
							if tagID, err := uuid.Parse(tagStr); err == nil {
								excludeTags = append(excludeTags, tagID)
							}
						}
					}
				}

				if matchTypeData, ok := input.TargetData["match_type"].(string); ok && (matchTypeData == "all" || matchTypeData == "any") {
					matchType = matchTypeData
				}
			}

			err = h.broadcastService.SetTargetTags(broadcastID, businessID, userID, includeTags, excludeTags, matchType)
		case "specific_users":
			// แปลงข้อมูล users
			userIDs := []uuid.UUID{}

			if input.TargetData != nil {
				if usersData, ok := input.TargetData["user_ids"].([]interface{}); ok {
					for _, user := range usersData {
						if userStr, ok := user.(string); ok {
							if targetUserID, err := uuid.Parse(userStr); err == nil {
								userIDs = append(userIDs, targetUserID)
							}
						}
					}
				}
			}

			err = h.broadcastService.SetTargetUsers(broadcastID, businessID, userID, userIDs)
		case "customer_profile":
			// แปลงข้อมูลเป็น BroadcastTargetCriteria
			criteria := &dto.BroadcastTargetCriteria{}

			if input.TargetData != nil {
				// ตัวอย่างการแปลง CustomerTypes
				if customerTypesData, ok := input.TargetData["customer_types"].([]interface{}); ok {
					customerTypes := []string{}
					for _, cType := range customerTypesData {
						if cTypeStr, ok := cType.(string); ok {
							customerTypes = append(customerTypes, cTypeStr)
						}
					}
					criteria.CustomerTypes = customerTypes
				}

				// แปลงข้อมูลอื่นๆ ตามต้องการ
				if statusesData, ok := input.TargetData["statuses"].([]interface{}); ok {
					statuses := []string{}
					for _, status := range statusesData {
						if statusStr, ok := status.(string); ok {
							statuses = append(statuses, statusStr)
						}
					}
					criteria.Statuses = statuses
				}

				if lastContactFromStr, ok := input.TargetData["last_contact_from"].(string); ok {
					if lastContactFrom, err := time.Parse(time.RFC3339, lastContactFromStr); err == nil {
						criteria.LastContactFrom = &lastContactFrom
					}
				}

				if lastContactToStr, ok := input.TargetData["last_contact_to"].(string); ok {
					if lastContactTo, err := time.Parse(time.RFC3339, lastContactToStr); err == nil {
						criteria.LastContactTo = &lastContactTo
					}
				}

				if customQuery, ok := input.TargetData["custom_query"].(map[string]interface{}); ok {
					criteria.CustomQuery = types.JSONB(customQuery)
				}
			}

			err = h.broadcastService.SetTargetCustomerProfile(broadcastID, businessID, userID, criteria)
		default:
			err = c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid target type",
			})
			return err
		}

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Error setting target: " + err.Error(),
			})
		}
	}

	// ตั้งเวลาส่ง (ถ้ามี)
	if input.ScheduleAt != nil {
		// เรียกใช้ service เพื่ออัปเดตในฐานข้อมูล
		err = h.broadcastService.ScheduleBroadcast(broadcastID, businessID, userID, *input.ScheduleAt)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Error scheduling broadcast: " + err.Error(),
			})
		}

		// ดึงข้อมูล broadcast ที่อัปเดตแล้ว
		updatedBroadcast, err = h.broadcastService.GetBroadcastByID(broadcastID, businessID, userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Error fetching updated broadcast: " + err.Error(),
			})
		}

		// เพิ่ม broadcast ลงใน scheduler
		if h.scheduler != nil { // เพิ่มการตรวจสอบ nil ตรงนี้
			err = h.scheduler.ScheduleBroadcast(updatedBroadcast)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"message": "Error adding to scheduler: " + err.Error(),
				})
			}
		} else {
			log.Printf("Warning: scheduler is nil, broadcast will not be scheduled")
		}
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"message":   "Broadcast updated successfully",
		"broadcast": updatedBroadcast,
	})
}

// DeleteBroadcast ลบ broadcast
func (h *BroadcastHandler) DeleteBroadcast(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก params
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// ดึง broadcastID จาก params
	broadcastID, err := utils.ParseUUIDParam(c, "broadcastId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// ลบ broadcast
	err = h.broadcastService.DeleteBroadcast(broadcastID, businessID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error deleting broadcast: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Broadcast deleted successfully",
	})
}

// SendBroadcast ส่ง broadcast ทันที
func (h *BroadcastHandler) SendBroadcast(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก params
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// ดึง broadcastID จาก params
	broadcastID, err := utils.ParseUUIDParam(c, "broadcastId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// ส่ง broadcast
	err = h.broadcastService.SendBroadcast(broadcastID, businessID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error sending broadcast: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Broadcast sent successfully",
	})
}

// CancelBroadcast ยกเลิกการส่ง broadcast ที่กำหนดเวลาไว้
func (h *BroadcastHandler) CancelBroadcast(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก params
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// ดึง broadcastID จาก params
	broadcastID, err := utils.ParseUUIDParam(c, "broadcastId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// ยกเลิกการส่ง broadcast ในฐานข้อมูล
	err = h.broadcastService.CancelScheduledBroadcast(broadcastID, businessID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error cancelling broadcast: " + err.Error(),
		})
	}

	// ยกเลิก broadcast ใน scheduler
	if h.scheduler != nil { // เพิ่มการตรวจสอบ nil ตรงนี้
		err = h.scheduler.CancelScheduledBroadcast(broadcastID)
		if err != nil {
			// บันทึก log แต่ไม่ return error เนื่องจากได้ยกเลิกในฐานข้อมูลแล้ว
			log.Printf("Warning: Error removing broadcast from scheduler: %v", err)
		}
	} else {
		log.Printf("Warning: scheduler is nil, broadcast cannot be removed from scheduler")
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Broadcast cancelled successfully",
	})
}

// GetBroadcastStats ดึงสถิติของ broadcast
func (h *BroadcastHandler) GetBroadcastStats(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก params
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// ดึง broadcastID จาก params
	broadcastID, err := utils.ParseUUIDParam(c, "broadcastId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// ดึงสถิติ broadcast
	stats, err := h.broadcastService.GetBroadcastStats(broadcastID, businessID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error fetching broadcast stats: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Broadcast stats retrieved successfully",
		"stats":   stats,
	})
}

// GetBroadcastDeliveries ดึงรายการการส่ง broadcast
func (h *BroadcastHandler) GetBroadcastDeliveries(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก params
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// ดึง broadcastID จาก params
	broadcastID, err := utils.ParseUUIDParam(c, "broadcastId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// รับพารามิเตอร์การแบ่งหน้า
	limit := c.QueryInt("limit", 10)
	page := c.QueryInt("page", 1)
	offset := (page - 1) * limit

	// รับพารามิเตอร์การกรอง
	status := c.Query("status", "")

	// ดึงรายการการส่ง
	deliveries, total, err := h.broadcastService.GetBroadcastDeliveries(broadcastID, businessID, userID, status, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error fetching broadcast deliveries: " + err.Error(),
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
		"message":    "Broadcast deliveries retrieved successfully",
		"deliveries": deliveries,
		"pagination": pagination,
	})
}

// PreviewBroadcast ดูตัวอย่างข้อความก่อนส่ง
func (h *BroadcastHandler) PreviewBroadcast(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก params
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// ดึง broadcastID จาก params
	broadcastID, err := utils.ParseUUIDParam(c, "broadcastId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// ดูตัวอย่าง broadcast
	preview, err := h.broadcastService.PreviewBroadcast(broadcastID, businessID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error previewing broadcast: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Broadcast preview generated successfully",
		"preview": preview,
	})
}

// DuplicateBroadcast สร้าง broadcast ใหม่โดยคัดลอกจาก broadcast เดิม
func (h *BroadcastHandler) DuplicateBroadcast(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก params
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// ดึง broadcastID จาก params
	broadcastID, err := utils.ParseUUIDParam(c, "broadcastId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// สร้าง broadcast ใหม่โดยคัดลอกจาก broadcast เดิม
	newBroadcast, err := h.broadcastService.DuplicateBroadcast(broadcastID, businessID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error duplicating broadcast: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"message":   "Broadcast duplicated successfully",
		"broadcast": newBroadcast,
	})
}

// GetEstimatedTargetCount ประมาณจำนวนผู้รับตามเงื่อนไขที่กำหนด
func (h *BroadcastHandler) GetEstimatedTargetCount(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก params
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// รับข้อมูลเงื่อนไขเป้าหมาย
	var input struct {
		TargetType     string                 `json:"target_type"`
		TargetCriteria map[string]interface{} `json:"target_criteria"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
	}

	// ตรวจสอบข้อมูลที่จำเป็น
	if input.TargetType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Target type is required",
		})
	}

	// แปลงข้อมูลเป็น BroadcastTargetCriteria
	criteria := &dto.BroadcastTargetCriteria{}

	// ตามประเภทเป้าหมาย
	switch input.TargetType {
	case "tags":
		// แปลงข้อมูล tags
		includeTags := []uuid.UUID{}
		excludeTags := []uuid.UUID{}
		matchType := "all"

		if input.TargetCriteria != nil {
			if includeTagsData, ok := input.TargetCriteria["include_tags"].([]interface{}); ok {
				for _, tag := range includeTagsData {
					if tagStr, ok := tag.(string); ok {
						if tagID, err := uuid.Parse(tagStr); err == nil {
							includeTags = append(includeTags, tagID)
						}
					}
				}
			}

			if excludeTagsData, ok := input.TargetCriteria["exclude_tags"].([]interface{}); ok {
				for _, tag := range excludeTagsData {
					if tagStr, ok := tag.(string); ok {
						if tagID, err := uuid.Parse(tagStr); err == nil {
							excludeTags = append(excludeTags, tagID)
						}
					}
				}
			}

			if matchTypeData, ok := input.TargetCriteria["match_type"].(string); ok && (matchTypeData == "all" || matchTypeData == "any") {
				matchType = matchTypeData
			}
		}

		criteria.IncludeTags = includeTags
		criteria.ExcludeTags = excludeTags
		criteria.TagMatchType = matchType

	case "specific_users":
		// แปลงข้อมูล users
		userIDs := []uuid.UUID{}

		if input.TargetCriteria != nil {
			if usersData, ok := input.TargetCriteria["user_ids"].([]interface{}); ok {
				for _, user := range usersData {
					if userStr, ok := user.(string); ok {
						if targetUserID, err := uuid.Parse(userStr); err == nil {
							userIDs = append(userIDs, targetUserID)
						}
					}
				}
			}
		}

		criteria.UserIDs = userIDs

	case "customer_profile":
		if input.TargetCriteria != nil {
			// ตัวอย่างการแปลง CustomerTypes
			if customerTypesData, ok := input.TargetCriteria["customer_types"].([]interface{}); ok {
				customerTypes := []string{}
				for _, cType := range customerTypesData {
					if cTypeStr, ok := cType.(string); ok {
						customerTypes = append(customerTypes, cTypeStr)
					}
				}
				criteria.CustomerTypes = customerTypes
			}

			// แปลงข้อมูลอื่นๆ ตามต้องการ
			if statusesData, ok := input.TargetCriteria["statuses"].([]interface{}); ok {
				statuses := []string{}
				for _, status := range statusesData {
					if statusStr, ok := status.(string); ok {
						statuses = append(statuses, statusStr)
					}
				}
				criteria.Statuses = statuses
			}

			if lastContactFromStr, ok := input.TargetCriteria["last_contact_from"].(string); ok {
				if lastContactFrom, err := time.Parse(time.RFC3339, lastContactFromStr); err == nil {
					criteria.LastContactFrom = &lastContactFrom
				}
			}

			if lastContactToStr, ok := input.TargetCriteria["last_contact_to"].(string); ok {
				if lastContactTo, err := time.Parse(time.RFC3339, lastContactToStr); err == nil {
					criteria.LastContactTo = &lastContactTo
				}
			}

			if customQuery, ok := input.TargetCriteria["custom_query"].(map[string]interface{}); ok {
				criteria.CustomQuery = types.JSONB(customQuery)
			}
		}
	}

	// ประมาณจำนวนผู้รับ
	count, err := h.broadcastService.GetEstimatedTargetCount(businessID, userID, input.TargetType, criteria)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error estimating target count: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Target count estimated successfully",
		"count":   count,
	})
}
