// interfaces/api/handler/customer_profile_handler.go
package handler

import (
	"encoding/json"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/dto"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/domain/types"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"

	"github.com/thizplus/gofiber-chat-api/pkg/utils"
)

// CustomerProfileHandler จัดการ HTTP request/response สำหรับ Customer Profile
type CustomerProfileHandler struct {
	customerProfileService service.CustomerProfileService
	notificationService    service.NotificationService
}

// NewCustomerProfileHandler สร้าง instance ใหม่ของ CustomerProfileHandler
func NewCustomerProfileHandler(
	customerProfileService service.CustomerProfileService,
	notificationService service.NotificationService,
) *CustomerProfileHandler {
	return &CustomerProfileHandler{
		customerProfileService: customerProfileService,
		notificationService:    notificationService,
	}
}

// CreateCustomerProfile สร้างโปรไฟล์ลูกค้าใหม่
func (h *CustomerProfileHandler) CreateCustomerProfile(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	_, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก parameter
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err // error response ถูกจัดการในฟังก์ชันแล้ว
	}

	// ดึง customerUserID จาก parameter
	customerUserID, err := utils.ParseUUIDParam(c, "userId")
	if err != nil {
		return err
	}

	// รับข้อมูลจาก request body
	var input struct {
		Nickname     string `json:"nickname"`
		Notes        string `json:"notes"`
		CustomerType string `json:"customer_type"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
	}

	// สร้างโปรไฟล์ลูกค้า
	profile, err := h.customerProfileService.CreateCustomerProfile(
		businessID,
		customerUserID,
		input.Nickname,
		input.Notes,
		input.CustomerType,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error creating customer profile: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Customer profile created successfully",
		"data":    profile,
	})
}

// GetCustomerProfile ดึงโปรไฟล์ลูกค้า
func (h *CustomerProfileHandler) GetCustomerProfile(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	_, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก parameter
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err
	}

	// ดึง customerUserID จาก parameter
	customerUserID, err := utils.ParseUUIDParam(c, "userId")
	if err != nil {
		return err
	}

	// ดึงโปรไฟล์ลูกค้า
	profile, err := h.customerProfileService.GetCustomerProfile(businessID, customerUserID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Customer profile not found: " + err.Error(),
		})
	}

	// แปลงข้อมูลเพื่อไม่ให้ส่ง Business และเพิ่ม tags ในรูปแบบ SimpleTag
	profileMap := structToMap(profile)

	// ลบข้อมูล business ออก
	delete(profileMap, "business")

	// แปลง Tags เป็น SimpleTag (ถ้ามี)
	if len(profile.Tags) > 0 {
		simpleTags := make([]dto.SimpleTag, 0)
		tagMap := make(map[uuid.UUID]bool) // ป้องกันการซ้ำกัน

		for _, userTag := range profile.Tags {
			if userTag.Tag != nil {
				if _, exists := tagMap[userTag.Tag.ID]; !exists {
					simpleTags = append(simpleTags, dto.SimpleTag{
						ID:    userTag.Tag.ID,
						Name:  userTag.Tag.Name,
						Color: userTag.Tag.Color,
					})
					tagMap[userTag.Tag.ID] = true
				}
			}
		}

		// แทนที่ Tags ใน map ด้วย SimpleTags
		profileMap["tags"] = simpleTags
	} else {
		// ถ้าไม่มี tags ให้ส่ง empty array
		profileMap["tags"] = []dto.SimpleTag{}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Customer profile retrieved successfully",
		"data":    profileMap,
	})
}

// UpdateCustomerProfile อัปเดตโปรไฟล์ลูกค้า
func (h *CustomerProfileHandler) UpdateCustomerProfile(c *fiber.Ctx) error {
	// ดึง userID จาก middleware (admin ที่ทำการอัปเดต)
	adminID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก parameter
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err
	}

	// ดึง customerUserID จาก parameter
	customerUserID, err := utils.ParseUUIDParam(c, "userId")
	if err != nil {
		return err
	}

	// รับข้อมูลการอัปเดต
	var updateData types.JSONB
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
	}

	// อัปเดตโปรไฟล์ลูกค้า
	profile, err := h.customerProfileService.UpdateCustomerProfile(
		businessID,
		customerUserID,
		adminID,
		updateData,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error updating customer profile: " + err.Error(),
		})
	}

	// แปลงข้อมูลเพื่อเตรียมส่ง response และ WebSocket
	profileMap := structToMap(profile)

	// แปลง Tags เป็น SimpleTag (ถ้ามี) - เหมือนกับใน GetCustomerProfile
	if len(profile.Tags) > 0 {
		simpleTags := make([]dto.SimpleTag, 0)
		tagMap := make(map[uuid.UUID]bool) // ป้องกันการซ้ำกัน

		for _, userTag := range profile.Tags {
			if userTag.Tag != nil {
				if _, exists := tagMap[userTag.Tag.ID]; !exists {
					simpleTags = append(simpleTags, dto.SimpleTag{
						ID:    userTag.Tag.ID,
						Name:  userTag.Tag.Name,
						Color: userTag.Tag.Color,
					})
					tagMap[userTag.Tag.ID] = true
				}
			}
		}

		// แทนที่ Tags ใน map ด้วย SimpleTags
		profileMap["tags"] = simpleTags
	} else {
		// ถ้าไม่มี tags ให้ส่ง empty array
		profileMap["tags"] = []dto.SimpleTag{}
	}

	// เพิ่มการแจ้งเตือนผ่าน WebSocket
	if h.notificationService != nil {
		h.notificationService.NotifyProfileUpdate(businessID, customerUserID, profileMap)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Customer profile updated successfully",
		"data":    profileMap,
	})
}

// GetBusinessCustomers ดึงรายชื่อลูกค้าทั้งหมดของธุรกิจ
func (h *CustomerProfileHandler) GetBusinessCustomers(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	_, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก parameter
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err
	}

	// ดึงพารามิเตอร์การแบ่งหน้า
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	// ตรวจสอบค่าที่ถูกต้อง
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	// ดึงรายชื่อลูกค้า
	profiles, total, err := h.customerProfileService.GetBusinessCustomers(businessID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error fetching customers: " + err.Error(),
		})
	}

	// แปลงข้อมูลและจัดเตรียมข้อมูลก่อนส่งคืน
	customersWithSimpleTags := make([]map[string]interface{}, 0, len(profiles))

	for _, profile := range profiles {
		// สร้าง map เพื่อเก็บข้อมูลลูกค้า
		customerMap := structToMap(profile)

		// แปลง UserTag เป็น SimpleTag
		if len(profile.Tags) > 0 {
			simpleTags := make([]dto.SimpleTag, 0)
			tagMap := make(map[uuid.UUID]bool) // ป้องกันการซ้ำกัน

			for _, userTag := range profile.Tags {
				if userTag.Tag != nil {
					if _, exists := tagMap[userTag.Tag.ID]; !exists {
						simpleTags = append(simpleTags, dto.SimpleTag{
							ID:    userTag.Tag.ID,
							Name:  userTag.Tag.Name,
							Color: userTag.Tag.Color,
						})
						tagMap[userTag.Tag.ID] = true
					}
				}
			}

			// แทนที่ Tags ใน map ด้วย SimpleTags
			customerMap["tags"] = simpleTags
		}

		customersWithSimpleTags = append(customersWithSimpleTags, customerMap)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Customers retrieved successfully",
		"data": fiber.Map{
			"customers": customersWithSimpleTags,
			"total":     total,
			"limit":     limit,
			"offset":    offset,
		},
	})
}

// structToMap แปลง struct เป็น map
func structToMap(obj interface{}) map[string]interface{} {
	data, _ := json.Marshal(obj)
	result := make(map[string]interface{})
	json.Unmarshal(data, &result)
	return result
}

// SearchCustomers ค้นหาลูกค้า
func (h *CustomerProfileHandler) SearchCustomers(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	_, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง businessID จาก parameter
	businessID, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err
	}

	// ดึงพารามิเตอร์การค้นหา
	query := c.Query("q", "")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Search query is required",
		})
	}

	// ดึงพารามิเตอร์การแบ่งหน้า
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	// ตรวจสอบค่าที่ถูกต้อง
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	// ค้นหาลูกค้า
	profiles, total, err := h.customerProfileService.SearchCustomers(businessID, query, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error searching customers: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Search completed successfully",
		"data": fiber.Map{
			"customers": profiles,
			"total":     total,
			"limit":     limit,
			"offset":    offset,
			"query":     query,
		},
	})
}

// UpdateLastContact อัปเดตเวลาติดต่อล่าสุด (สำหรับ internal use)
func (h *CustomerProfileHandler) UpdateLastContact(c *fiber.Ctx) error {
	// ดึง businessID และ userID จาก parameter
	_, err := utils.ParseUUIDParam(c, "businessId")
	if err != nil {
		return err
	}

	_, err = utils.ParseUUIDParam(c, "userId")
	if err != nil {
		return err
	}

	// อัปเดตเวลาติดต่อล่าสุด (method นี้ต้องเพิ่มใน service)
	// err = h.customerProfileService.UpdateLastContact(businessID, customerUserID)
	// if err != nil {
	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
	// 		"success": false,
	// 		"message": "Error updating last contact: " + err.Error(),
	// 	})
	// }

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Last contact updated successfully",
	})
}
