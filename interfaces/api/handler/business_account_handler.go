// interfaces/api/handler/business_account_handler.go
package handler

import (
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/domain/types"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
	"github.com/thizplus/gofiber-chat-api/pkg/utils"
)

// BusinessAccountHandler จัดการ HTTP request/response สำหรับธุรกิจ
type BusinessAccountHandler struct {
	businessAccountService service.BusinessAccountService // ตัวเล็ก
	storageService         service.FileStorageService
}

// NewBusinessAccountHandler สร้าง instance ใหม่ของ BusinessAccountHandler
func NewBusinessAccountHandler(
	businessAccountService service.BusinessAccountService,
	storageService service.FileStorageService,
) *BusinessAccountHandler {
	return &BusinessAccountHandler{
		businessAccountService: businessAccountService,
		storageService:         storageService,
	}
}

// GetDetail ดึงรายละเอียดธุรกิจ
func (h *BusinessAccountHandler) GetDetail(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userIDStr := c.Locals("userID").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid user ID format",
		})
	}

	businessIDStr := c.Params("businessId")
	if businessIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Business ID is required",
		})
	}

	businessID, err := uuid.Parse(businessIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid business ID format",
		})
	}

	// ดึงข้อมูลธุรกิจ
	business, err := h.businessAccountService.GetBusinessByID(businessID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error fetching business: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"message":  "Business retrieved successfully",
		"business": business,
	})
}

// GetUserBusinesses ดึงรายการธุรกิจของผู้ใช้

func (h *BusinessAccountHandler) GetUserBusinesses(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userIDStr := c.Locals("userID").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid user ID format",
		})
	}

	// ดึงธุรกิจที่ผู้ใช้เป็นแอดมินในรูปแบบ DTO แล้ว
	businessDTOs, err := h.businessAccountService.GetUserBusinessDTOs(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error fetching businesses: " + err.Error(),
		})
	}

	// ถ้าไม่มีธุรกิจ
	if len(businessDTOs) == 0 {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "No businesses found",
			"data":    []interface{}{},
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Businesses retrieved successfully",
		"data":    businessDTOs,
	})
}

// Create สร้างธุรกิจใหม่
func (h *BusinessAccountHandler) Create(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userIDStr := c.Locals("userID").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid user ID format",
		})
	}

	// รับข้อมูลธุรกิจจาก request body
	var input types.JSONB
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
	}

	// ตรวจสอบข้อมูลที่จำเป็น
	name, ok := input["name"].(string)
	if !ok || name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Business name is required",
		})
	}

	// ตรวจสอบ username
	username, ok := input["username"].(string)
	if !ok || username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Business username is required",
		})
	}

	// กำหนดค่าเริ่มต้นสำหรับข้อมูลที่ไม่จำเป็น
	description := ""
	if descVal, ok := input["description"].(string); ok {
		description = descVal
	}

	welcomeMessage := ""
	if welcomeVal, ok := input["welcome_message"].(string); ok {
		welcomeMessage = welcomeVal
	}

	// สร้างธุรกิจใหม่
	business, err := h.businessAccountService.CreateBusiness(userID, name, username, description, welcomeMessage)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error creating business: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success":  true,
		"message":  "Business created successfully",
		"business": business,
	})
}

// Update อัพเดทข้อมูลธุรกิจ
func (h *BusinessAccountHandler) Update(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userIDStr := c.Locals("userID").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid user ID format",
		})
	}

	businessIDStr := c.Params("businessId")
	if businessIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Business ID is required",
		})
	}

	businessID, err := uuid.Parse(businessIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid business ID format",
		})
	}

	// รับข้อมูลที่ต้องการอัพเดท
	var input types.JSONB
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
	}

	// ตรวจสอบและกรองข้อมูลที่อนุญาตให้อัพเดท
	updateData := make(types.JSONB)
	if name, ok := input["name"].(string); ok && name != "" {
		updateData["name"] = name
	}

	if description, ok := input["description"].(string); ok {
		updateData["description"] = description
	}

	if profileImageURL, ok := input["profile_image_url"].(string); ok {
		updateData["profile_image_url"] = profileImageURL
	}

	if coverImageURL, ok := input["cover_image_url"].(string); ok {
		updateData["cover_image_url"] = coverImageURL
	}

	/*
		if welcomeMessage, ok := input["welcome_message"].(string); ok {
			updateData["welcome_message"] = welcomeMessage
		}
	*/

	// เพิ่มการรองรับฟิลด์ status
	if status, ok := input["status"].(string); ok && (status == "active" || status == "deleted") {
		updateData["status"] = status
	}

	// ถ้าไม่มีข้อมูลที่จะอัพเดท
	if len(updateData) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "No valid data to update",
		})
	}

	// อัพเดทข้อมูล
	business, err := h.businessAccountService.UpdateBusiness(businessID, userID, updateData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error updating business: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"message":  "Business updated successfully",
		"business": business,
	})
}

// Delete ลบธุรกิจ
func (h *BusinessAccountHandler) Delete(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userIDStr := c.Locals("userID").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid user ID format",
		})
	}

	businessIDStr := c.Params("businessId")
	if businessIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Business ID is required",
		})
	}

	businessID, err := uuid.Parse(businessIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid business ID format",
		})
	}

	// ลบธุรกิจ
	err = h.businessAccountService.DeleteBusiness(businessID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error deleting business: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Business deleted successfully",
	})
}

// SearchByUsername ค้นหาธุรกิจด้วย username
func (h *BusinessAccountHandler) SearchByUsername(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userIDStr := c.Locals("userID").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid user ID format",
		})
	}

	username := c.Query("username")
	if username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Username is required",
		})
	}

	// ดึงพารามิเตอร์ exact_match จาก query
	exactMatch := false
	if c.Query("exact_match") == "true" {
		exactMatch = true
	}

	// ค้นหาธุรกิจ
	var business *models.BusinessAccount
	var searchErr error

	if exactMatch {
		// ค้นหาแบบตรงกับทั้งหมด
		business, searchErr = h.businessAccountService.GetBusinessByUsernameExact(username, userID)
	} else {
		// ค้นหาแบบเดิม
		business, searchErr = h.businessAccountService.GetBusinessByUsername(username, userID)
	}

	if searchErr != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Business not found: " + searchErr.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"message":  "Business found",
		"business": business,
	})
}

// interfaces/api/handler/business_account_handler.go
// เพิ่มเมธอดเหล่านี้ในคลาส BusinessAccountHandler

// UploadProfileImage อัปโหลดรูปโปรไฟล์ธุรกิจ
func (h *BusinessAccountHandler) UploadProfileImage(c *fiber.Ctx) error {
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

	// รับไฟล์ที่อัปโหลด
	file, err := c.FormFile("profile_image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "No image file uploaded",
		})
	}

	// ตรวจสอบขนาดไฟล์ (เช่น ไม่เกิน 5MB)
	maxSize := 5 * 1024 * 1024 // 5MB
	if file.Size > int64(maxSize) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "File too large (max 5MB)",
		})
	}

	// ตรวจสอบประเภทไฟล์
	ext := strings.ToLower(filepath.Ext(file.Filename))
	validExt := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
	}

	if !validExt[ext] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid file type. Only JPG, JPEG, PNG and GIF are allowed",
		})
	}

	// อัปโหลดไฟล์โดยใช้ storageService
	result, err := h.storageService.UploadImage(file, "business_profile_images")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to upload image to cloud storage",
			"error":   err.Error(),
		})
	}

	// ใช้ URL ที่ได้จากการอัปโหลด
	imageURL := result.URL

	// อัปเดต URL ในฐานข้อมูล
	if err := h.businessAccountService.UploadBusinessProfileImage(businessID, userID, imageURL); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error updating business profile image: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Business profile image uploaded successfully",
		"data": fiber.Map{
			"profile_image_url": imageURL,
			"public_id":         result.PublicID,
		},
	})
}

// UploadCoverImage อัปโหลดรูปปกธุรกิจ
func (h *BusinessAccountHandler) UploadCoverImage(c *fiber.Ctx) error {
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

	// รับไฟล์ที่อัปโหลด
	file, err := c.FormFile("cover_image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "No image file uploaded",
		})
	}

	// ตรวจสอบขนาดไฟล์ (เช่น ไม่เกิน 5MB)
	maxSize := 5 * 1024 * 1024 // 5MB
	if file.Size > int64(maxSize) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "File too large (max 5MB)",
		})
	}

	// ตรวจสอบประเภทไฟล์
	ext := strings.ToLower(filepath.Ext(file.Filename))
	validExt := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
	}

	if !validExt[ext] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid file type. Only JPG, JPEG, PNG and GIF are allowed",
		})
	}

	// อัปโหลดไฟล์โดยใช้ storageService
	result, err := h.storageService.UploadImage(file, "business_cover_images")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to upload image to cloud storage",
			"error":   err.Error(),
		})
	}

	// ใช้ URL ที่ได้จากการอัปโหลด
	imageURL := result.URL

	// อัปเดต URL ในฐานข้อมูล
	if err := h.businessAccountService.UploadBusinessCoverImage(businessID, userID, imageURL); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error updating business cover image: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Business cover image uploaded successfully",
		"data": fiber.Map{
			"cover_image_url": imageURL,
			"public_id":       result.PublicID,
		},
	})
}

// SearchBusinesses ค้นหาธุรกิจ
func (h *BusinessAccountHandler) SearchBusinesses(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึงพารามิเตอร์การค้นหา
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Search query is required",
		})
	}

	// ดึง limit และ offset จาก query params
	limit := utils.ParseIntWithLimit(c.Query("limit"), 20, 1, 50)
	offset := utils.ParseInt(c.Query("offset"), 0)

	// ดึงพารามิเตอร์ exact_match จาก query
	exactMatch := false
	if c.Query("exact_match") == "true" {
		exactMatch = true
	}

	// ค้นหาธุรกิจ
	var businesses []*models.BusinessAccount
	var total int64
	var searchErr error

	if exactMatch {
		// ค้นหาแบบตรงกับทั้งหมด
		businesses, total, searchErr = h.businessAccountService.SearchBusinessesExact(query, limit, offset, userID)
	} else {
		// ค้นหาแบบเดิม
		businesses, total, searchErr = h.businessAccountService.SearchBusinesses(query, limit, offset, userID)
	}

	if searchErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error searching businesses: " + searchErr.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"businesses": businesses,
			"total":      total,
			"limit":      limit,
			"offset":     offset,
			"query":      query,
		},
	})
}
