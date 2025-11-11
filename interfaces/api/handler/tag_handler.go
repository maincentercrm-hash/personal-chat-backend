// interfaces/api/handler/tag_handler.go
package handler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/dto"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"

	"github.com/thizplus/gofiber-chat-api/pkg/utils"
)

// TagHandler จัดการ HTTP request/response สำหรับ Tag
type TagHandler struct {
	tagService          service.TagService
	userTagService      service.UserTagService
	notificationService service.NotificationService
}

// NewTagHandler สร้าง instance ใหม่ของ TagHandler
func NewTagHandler(tagService service.TagService, userTagService service.UserTagService, notificationService service.NotificationService) *TagHandler {
	return &TagHandler{
		tagService:          tagService,
		userTagService:      userTagService,
		notificationService: notificationService,
	}
}

// CreateTag สร้างแท็กใหม่
func (h *TagHandler) CreateTag(c *fiber.Ctx) error {
	// ดึง userID จาก middleware (admin ที่สร้างแท็ก)
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

	// รับข้อมูลจาก request body
	var input struct {
		Name  string `json:"name" validate:"required,max=50"`
		Color string `json:"color"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
	}

	// ตรวจสอบข้อมูลที่จำเป็น
	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Tag name is required",
		})
	}

	// สร้างแท็ก
	tag, err := h.tagService.CreateTag(businessID, adminID, input.Name, input.Color)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error creating tag: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Tag created successfully",
		"data":    tag,
	})
}

// GetBusinessTags ดึงแท็กทั้งหมดของธุรกิจ
// handlers/tag_handler.go
func (h *TagHandler) GetBusinessTags(c *fiber.Ctx) error {
	// ตรวจสอบการยืนยันตัวตน
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

	// ใช้เมธอดใหม่ที่คืนค่าเป็น TagInfo DTOs พร้อม user_count
	tagInfos, err := h.tagService.GetBusinessTagsWithInfo(businessID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error fetching tags: " + err.Error(),
		})
	}

	// สร้าง response ตามรูปแบบ GetBusinessTagsResponse
	response := dto.GetBusinessTagsResponse{
		GenericResponse: dto.GenericResponse{
			Success: true,
			Message: "Tags retrieved successfully",
		},
		Data: struct {
			Tags  []dto.TagInfo `json:"tags"`
			Count int           `json:"count"`
		}{
			Tags:  tagInfos,
			Count: len(tagInfos),
		},
	}

	return c.JSON(response)
}

// UpdateTag อัปเดตแท็ก
func (h *TagHandler) UpdateTag(c *fiber.Ctx) error {
	// ดึง userID จาก middleware (admin ที่อัปเดต)
	userID, err := middleware.GetUserUUID(c)
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

	// ดึง tagID จาก parameter
	tagID, err := utils.ParseUUIDParam(c, "tagId")
	if err != nil {
		return err
	}

	// รับข้อมูลการอัปเดต
	var input struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
	}

	// ตรวจสอบข้อมูลที่จำเป็น
	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Tag name is required",
		})
	}

	// อัปเดตแท็ก
	tag, err := h.tagService.UpdateTag(businessID, tagID, userID, input.Name, input.Color)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error updating tag: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Tag updated successfully",
		"data":    tag,
	})
}

// DeleteTag ลบแท็ก
func (h *TagHandler) DeleteTag(c *fiber.Ctx) error {
	// ดึง userID จาก middleware (admin ที่ลบ)
	userID, err := middleware.GetUserUUID(c)
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

	// ดึง tagID จาก parameter
	tagID, err := utils.ParseUUIDParam(c, "tagId")
	if err != nil {
		return err
	}

	// ลบแท็ก
	err = h.tagService.DeleteTag(businessID, tagID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error deleting tag: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Tag deleted successfully",
	})
}

// AddTagToUser เพิ่มแท็กให้ผู้ใช้
func (h *TagHandler) AddTagToUser(c *fiber.Ctx) error {
	// ดึง userID จาก middleware (admin ที่เพิ่มแท็ก)
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

	// ดึง tagID จาก parameter
	tagID, err := utils.ParseUUIDParam(c, "tagId")
	if err != nil {
		return err
	}

	// ดึง customerUserID จาก parameter
	customerUserID, err := utils.ParseUUIDParam(c, "userId")
	if err != nil {
		return err
	}

	// เพิ่ม logging เพื่อตรวจสอบการทำงาน
	fmt.Printf("Adding tag: businessID=%s, tagID=%s, userID=%s, adminID=%s\n",
		businessID, tagID, customerUserID, adminID)

	// เพิ่มแท็กให้ผู้ใช้
	err = h.tagService.AddTagToUser(businessID, customerUserID, tagID, adminID)
	if err != nil {
		// ถ้าเป็นข้อผิดพลาดว่ามีแท็กอยู่แล้ว ส่ง status 409 Conflict แทน
		if strings.Contains(err.Error(), "already has this tag") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error adding tag to user: " + err.Error(),
		})
	}

	// เพิ่มการแจ้งเตือนผ่าน WebSocket
	if h.notificationService != nil {
		// แจ้งเตือนการอัพเดทแท็ก พร้อมข้อมูลเพิ่มเติม
		h.notificationService.NotifyProfileUpdateTags(businessID, customerUserID, tagID, "add")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Tag added to user successfully",
	})
}

// RemoveTagFromUser ลบแท็กจากผู้ใช้
func (h *TagHandler) RemoveTagFromUser(c *fiber.Ctx) error {
	// ดึง userID จาก middleware (admin ที่ลบแท็ก)
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

	// ดึง tagID จาก parameter
	tagID, err := utils.ParseUUIDParam(c, "tagId")
	if err != nil {
		return err
	}

	// ดึง customerUserID จาก parameter
	customerUserID, err := utils.ParseUUIDParam(c, "userId")
	if err != nil {
		return err
	}

	// ลบแท็กจากผู้ใช้
	err = h.tagService.RemoveTagFromUser(businessID, customerUserID, tagID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error removing tag from user: " + err.Error(),
		})
	}

	// เพิ่มการแจ้งเตือนผ่าน WebSocket
	if h.notificationService != nil {
		// แจ้งเตือนการอัพเดทแท็ก พร้อมข้อมูลเพิ่มเติม
		h.notificationService.NotifyProfileUpdateTags(businessID, customerUserID, tagID, "remove")
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Tag removed from user successfully",
	})
}

// GetUserTags ดึงแท็กของผู้ใช้
func (h *TagHandler) GetUserTags(c *fiber.Ctx) error {
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

	// ดึงแท็กของผู้ใช้
	tags, err := h.tagService.GetUserTags(businessID, customerUserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error fetching user tags: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "User tags retrieved successfully",
		"data": fiber.Map{
			"tags":  tags,
			"count": len(tags),
		},
	})
}

// GetUsersByTag ดึงผู้ใช้ที่มีแท็กนี้
func (h *TagHandler) GetUsersByTag(c *fiber.Ctx) error {
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

	// ดึง tagID จาก parameter
	tagID, err := utils.ParseUUIDParam(c, "tagId")
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

	customers, err := h.tagService.GetUsersByTag(businessID, tagID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error fetching users by tag: " + err.Error(),
		})
	}

	// แปลงข้อมูลให้ตรงกับที่ frontend ต้องการ
	taggedUsers := make([]fiber.Map, 0, len(customers))
	for _, customer := range customers {
		if customer.User != nil { // ตรวจสอบว่ามีข้อมูล User หรือไม่
			taggedUsers = append(taggedUsers, fiber.Map{
				"user_id":           customer.User.ID,
				"display_name":      customer.User.DisplayName,
				"profile_image_url": customer.User.ProfileImageURL,
				"username":          customer.User.Username,
				"email":             customer.User.Email,
			})
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Users retrieved successfully",
		"data":    taggedUsers, // ส่งเฉพาะข้อมูลที่ frontend ต้องการ
	})
}

// BulkAddTagToUsers เพิ่มแท็กให้ผู้ใช้หลายคน
func (h *TagHandler) BulkAddTagToUsers(c *fiber.Ctx) error {
	// ดึง userID จาก middleware (admin ที่เพิ่มแท็ก)
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

	// ดึง tagID จาก parameter
	tagID, err := utils.ParseUUIDParam(c, "tagId")
	if err != nil {
		return err
	}

	// รับข้อมูลรายชื่อผู้ใช้
	var input struct {
		UserIDs []string `json:"user_ids" validate:"required,min=1"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
	}

	if len(input.UserIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "User IDs are required",
		})
	}

	// ตรวจสอบจำนวนผู้ใช้ที่อนุญาต
	if len(input.UserIDs) > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Maximum 100 users allowed per request",
		})
	}

	// แปลง string UUIDs เป็น uuid.UUID
	var userIDs []uuid.UUID
	for _, userIDStr := range input.UserIDs {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Invalid user ID format: " + userIDStr,
			})
		}
		userIDs = append(userIDs, userID)
	}

	// เพิ่มแท็กให้ผู้ใช้หลายคน
	addedUserTags, err := h.userTagService.BulkAddTagToUsers(businessID, tagID, userIDs, adminID)
	if err != nil {
		// อาจมี partial success
		return c.Status(fiber.StatusMultiStatus).JSON(fiber.Map{
			"success": false,
			"message": "Some operations failed: " + err.Error(),
			"data": fiber.Map{
				"successful_count": len(addedUserTags),
				"requested_count":  len(userIDs),
				"added_user_tags":  addedUserTags,
			},
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Tag added to users successfully",
		"data": fiber.Map{
			"successful_count": len(addedUserTags),
			"requested_count":  len(userIDs),
			"added_user_tags":  addedUserTags,
		},
	})
}
