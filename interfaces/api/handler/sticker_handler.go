// interfaces/api/handler/sticker_handler.go
package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
	"github.com/thizplus/gofiber-chat-api/pkg/utils"
)

// StickerHandler จัดการคำขอ API เกี่ยวกับสติกเกอร์
type StickerHandler struct {
	stickerService service.StickerService
}

// NewStickerHandler สร้าง handler สำหรับจัดการเกี่ยวกับสติกเกอร์
func NewStickerHandler(stickerService service.StickerService) *StickerHandler {
	return &StickerHandler{
		stickerService: stickerService,
	}
}

// CreateStickerSet สร้างชุดสติกเกอร์ใหม่
func (h *StickerHandler) CreateStickerSet(c *fiber.Ctx) error {
	// ตรวจสอบสิทธิ์ (ในกรณีนี้สมมติว่าเฉพาะแอดมินเท่านั้นที่สามารถสร้างชุดสติกเกอร์ได้)
	// ดึง userID จาก middleware
	_, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// TODO: ตรวจสอบว่าผู้ใช้เป็นแอดมินหรือไม่

	// รับข้อมูลจาก request
	var input struct {
		Name        string `json:"name" validate:"required"`
		Description string `json:"description"`
		Author      string `json:"author"`
		IsOfficial  bool   `json:"is_official"`
		IsDefault   bool   `json:"is_default"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data",
		})
	}

	// ตรวจสอบข้อมูล
	if input.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Sticker set name is required",
		})
	}

	// สร้างชุดสติกเกอร์
	stickerSet, err := h.stickerService.CreateStickerSet(
		input.Name,
		input.Description,
		input.Author,
		input.IsOfficial,
		input.IsDefault,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create sticker set: " + err.Error(),
		})
	}

	// ตอบกลับ
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Sticker set created successfully",
		"data":    stickerSet,
	})
}

// GetStickerSet ดึงข้อมูลชุดสติกเกอร์
func (h *StickerHandler) GetStickerSet(c *fiber.Ctx) error {
	// ดึงและแปลง stickerSetId จาก URL parameter เป็น UUID
	stickerSetID, err := utils.ParseUUIDParam(c, "stickerSetId")
	if err != nil {
		return err
	}

	// ดึงข้อมูลชุดสติกเกอร์
	stickerSet, err := h.stickerService.GetStickerSetByID(stickerSetID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Sticker set not found",
		})
	}

	// ดึงสติกเกอร์ในชุด
	stickers, err := h.stickerService.GetStickersBySetID(stickerSetID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get stickers: " + err.Error(),
		})
	}

	// ตอบกลับ
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"sticker_set": stickerSet,
			"stickers":    stickers,
		},
	})
}

// GetAllStickerSets ดึงข้อมูลชุดสติกเกอร์ทั้งหมด
func (h *StickerHandler) GetAllStickerSets(c *fiber.Ctx) error {
	// ดึง limit และ offset จาก query params
	limit := utils.ParseIntWithLimit(c.Query("limit"), 20, 1, 50)
	offset := utils.ParseInt(c.Query("offset"), 0)

	// ดึงข้อมูลชุดสติกเกอร์ทั้งหมด
	stickerSets, total, err := h.stickerService.GetAllStickerSets(limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get sticker sets: " + err.Error(),
		})
	}

	// ตอบกลับ
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"sticker_sets": stickerSets,
			"count":        total,
			"limit":        limit,
			"offset":       offset,
		},
	})
}

// GetDefaultStickerSets ดึงข้อมูลชุดสติกเกอร์เริ่มต้น
func (h *StickerHandler) GetDefaultStickerSets(c *fiber.Ctx) error {
	// ดึงข้อมูลชุดสติกเกอร์เริ่มต้น
	stickerSets, err := h.stickerService.GetDefaultStickerSets()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get default sticker sets: " + err.Error(),
		})
	}

	// ตอบกลับ
	return c.JSON(fiber.Map{
		"success": true,
		"data":    stickerSets,
	})
}

// UpdateStickerSet อัปเดตข้อมูลชุดสติกเกอร์
func (h *StickerHandler) UpdateStickerSet(c *fiber.Ctx) error {
	// ตรวจสอบสิทธิ์ (ในกรณีนี้สมมติว่าเฉพาะแอดมินเท่านั้นที่สามารถอัปเดตชุดสติกเกอร์ได้)
	// ดึง userID จาก middleware
	_, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// TODO: ตรวจสอบว่าผู้ใช้เป็นแอดมินหรือไม่

	// ดึงและแปลง stickerSetId จาก URL parameter เป็น UUID
	stickerSetID, err := utils.ParseUUIDParam(c, "stickerSetId")
	if err != nil {
		return err
	}

	// รับข้อมูลจาก request
	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Author      string `json:"author"`
		IsOfficial  bool   `json:"is_official"`
		IsDefault   bool   `json:"is_default"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data",
		})
	}

	// อัปเดตชุดสติกเกอร์
	stickerSet, err := h.stickerService.UpdateStickerSet(
		stickerSetID,
		input.Name,
		input.Description,
		input.Author,
		input.IsOfficial,
		input.IsDefault,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update sticker set: " + err.Error(),
		})
	}

	// ตอบกลับ
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Sticker set updated successfully",
		"data":    stickerSet,
	})
}

// UploadStickerSetCover อัปโหลดรูปปกชุดสติกเกอร์
func (h *StickerHandler) UploadStickerSetCover(c *fiber.Ctx) error {
	// ตรวจสอบสิทธิ์ (ในกรณีนี้สมมติว่าเฉพาะแอดมินเท่านั้นที่สามารถอัปโหลดรูปปกได้)
	// ดึง userID จาก middleware
	_, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// TODO: ตรวจสอบว่าผู้ใช้เป็นแอดมินหรือไม่

	// ดึงและแปลง stickerSetId จาก URL parameter เป็น UUID
	stickerSetID, err := utils.ParseUUIDParam(c, "stickerSetId")
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

	// อัปโหลดรูปปก
	stickerSet, err := h.stickerService.UploadStickerSetCover(stickerSetID, file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to upload cover image: " + err.Error(),
		})
	}

	// ตอบกลับ
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Cover image uploaded successfully",
		"data": fiber.Map{
			"sticker_set":     stickerSet,
			"cover_image_url": stickerSet.CoverImageURL,
		},
	})
}

// DeleteStickerSet ลบชุดสติกเกอร์
func (h *StickerHandler) DeleteStickerSet(c *fiber.Ctx) error {
	// ตรวจสอบสิทธิ์ (ในกรณีนี้สมมติว่าเฉพาะแอดมินเท่านั้นที่สามารถลบชุดสติกเกอร์ได้)
	// ดึง userID จาก middleware
	_, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// TODO: ตรวจสอบว่าผู้ใช้เป็นแอดมินหรือไม่

	// ดึงและแปลง stickerSetId จาก URL parameter เป็น UUID
	stickerSetID, err := utils.ParseUUIDParam(c, "stickerSetId")
	if err != nil {
		return err
	}

	// ลบชุดสติกเกอร์
	if err := h.stickerService.DeleteStickerSet(stickerSetID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete sticker set: " + err.Error(),
		})
	}

	// ตอบกลับ
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Sticker set deleted successfully",
	})
}

// AddStickerToSet เพิ่มสติกเกอร์ใหม่ลงในชุด
func (h *StickerHandler) AddStickerToSet(c *fiber.Ctx) error {
	// ตรวจสอบสิทธิ์ (ในกรณีนี้สมมติว่าเฉพาะแอดมินเท่านั้นที่สามารถเพิ่มสติกเกอร์ได้)
	// ดึง userID จาก middleware
	_, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// TODO: ตรวจสอบว่าผู้ใช้เป็นแอดมินหรือไม่

	// ดึงและแปลง stickerSetId จาก URL parameter เป็น UUID
	stickerSetID, err := utils.ParseUUIDParam(c, "stickerSetId")
	if err != nil {
		return err
	}

	// รับข้อมูลจาก form
	name := c.FormValue("name")
	sortOrderStr := c.FormValue("sort_order")
	isAnimatedStr := c.FormValue("is_animated")

	// แปลงค่า sort_order เป็น int
	sortOrder := utils.ParseInt(sortOrderStr, 0)

	// แปลงค่า is_animated เป็น bool
	isAnimated := isAnimatedStr == "true"

	// รับไฟล์ที่อัปโหลด
	file, err := c.FormFile("sticker_image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "No image file uploaded",
		})
	}

	// เพิ่มสติกเกอร์
	sticker, err := h.stickerService.AddStickerToSet(stickerSetID, name, file, isAnimated, sortOrder)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to add sticker: " + err.Error(),
		})
	}

	// ตอบกลับ
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Sticker added successfully",
		"data":    sticker,
	})
}

// UpdateSticker อัปเดตข้อมูลสติกเกอร์
func (h *StickerHandler) UpdateSticker(c *fiber.Ctx) error {
	// ตรวจสอบสิทธิ์ (ในกรณีนี้สมมติว่าเฉพาะแอดมินเท่านั้นที่สามารถอัปเดตสติกเกอร์ได้)
	// ดึง userID จาก middleware
	_, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// TODO: ตรวจสอบว่าผู้ใช้เป็นแอดมินหรือไม่

	// ดึงและแปลง stickerId จาก URL parameter เป็น UUID
	stickerID, err := utils.ParseUUIDParam(c, "stickerId")
	if err != nil {
		return err
	}

	// รับข้อมูลจาก request
	var input struct {
		Name      string `json:"name"`
		SortOrder int    `json:"sort_order"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data",
		})
	}

	// อัปเดตสติกเกอร์
	sticker, err := h.stickerService.UpdateSticker(stickerID, input.Name, input.SortOrder)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update sticker: " + err.Error(),
		})
	}

	// ตอบกลับ
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Sticker updated successfully",
		"data":    sticker,
	})
}

// DeleteSticker ลบสติกเกอร์
func (h *StickerHandler) DeleteSticker(c *fiber.Ctx) error {
	// ตรวจสอบสิทธิ์ (ในกรณีนี้สมมติว่าเฉพาะแอดมินเท่านั้นที่สามารถลบสติกเกอร์ได้)
	// ดึง userID จาก middleware
	_, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// TODO: ตรวจสอบว่าผู้ใช้เป็นแอดมินหรือไม่

	// ดึงและแปลง stickerId จาก URL parameter เป็น UUID
	stickerID, err := utils.ParseUUIDParam(c, "stickerId")
	if err != nil {
		return err
	}

	// ลบสติกเกอร์
	if err := h.stickerService.DeleteSticker(stickerID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete sticker: " + err.Error(),
		})
	}

	// ตอบกลับ
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Sticker deleted successfully",
	})
}

// AddStickerSetToUser เพิ่มชุดสติกเกอร์ให้ผู้ใช้
func (h *StickerHandler) AddStickerSetToUser(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึงและแปลง stickerSetId จาก URL parameter เป็น UUID
	stickerSetID, err := utils.ParseUUIDParam(c, "stickerSetId")
	if err != nil {
		return err
	}

	// เพิ่มชุดสติกเกอร์ให้ผู้ใช้
	if err := h.stickerService.AddStickerSetToUser(userID, stickerSetID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to add sticker set to user: " + err.Error(),
		})
	}

	// ตอบกลับ
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Sticker set added to user successfully",
	})
}

// GetUserStickerSets ดึงชุดสติกเกอร์ของผู้ใช้
func (h *StickerHandler) GetUserStickerSets(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึงชุดสติกเกอร์ของผู้ใช้
	stickerSets, err := h.stickerService.GetUserStickerSets(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get user sticker sets: " + err.Error(),
		})
	}

	// ตอบกลับ
	return c.JSON(fiber.Map{
		"success": true,
		"data":    stickerSets,
	})
}

// SetStickerSetAsFavorite ตั้งค่าชุดสติกเกอร์เป็นรายการโปรด
func (h *StickerHandler) SetStickerSetAsFavorite(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึงและแปลง stickerSetId จาก URL parameter เป็น UUID
	stickerSetID, err := utils.ParseUUIDParam(c, "stickerSetId")
	if err != nil {
		return err
	}

	// รับค่า is_favorite จาก request
	var input struct {
		IsFavorite bool `json:"is_favorite"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data",
		})
	}

	// ตั้งค่าชุดสติกเกอร์เป็นรายการโปรด
	if err := h.stickerService.SetStickerSetAsFavorite(userID, stickerSetID, input.IsFavorite); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to set sticker set as favorite: " + err.Error(),
		})
	}

	// ตอบกลับ
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Sticker set favorite status updated successfully",
	})
}

// RemoveStickerSetFromUser ลบชุดสติกเกอร์ออกจากผู้ใช้
func (h *StickerHandler) RemoveStickerSetFromUser(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึงและแปลง stickerSetId จาก URL parameter เป็น UUID
	stickerSetID, err := utils.ParseUUIDParam(c, "stickerSetId")
	if err != nil {
		return err
	}

	// ลบชุดสติกเกอร์ออกจากผู้ใช้
	if err := h.stickerService.RemoveStickerSetFromUser(userID, stickerSetID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to remove sticker set from user: " + err.Error(),
		})
	}

	// ตอบกลับ
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Sticker set removed from user successfully",
	})
}

// RecordStickerUsage บันทึกการใช้งานสติกเกอร์
func (h *StickerHandler) RecordStickerUsage(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึงและแปลง stickerId จาก URL parameter เป็น UUID
	stickerID, err := utils.ParseUUIDParam(c, "stickerId")
	if err != nil {
		return err
	}

	// บันทึกการใช้งานสติกเกอร์
	if err := h.stickerService.RecordStickerUsage(userID, stickerID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to record sticker usage: " + err.Error(),
		})
	}

	// ตอบกลับ
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Sticker usage recorded successfully",
	})
}

// GetUserRecentStickers ดึงสติกเกอร์ที่ใช้ล่าสุดของผู้ใช้
func (h *StickerHandler) GetUserRecentStickers(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึง limit จาก query params
	limit := utils.ParseIntWithLimit(c.Query("limit"), 20, 1, 50)

	// ดึงสติกเกอร์ที่ใช้ล่าสุดของผู้ใช้
	stickers, err := h.stickerService.GetUserRecentStickers(userID, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get user recent stickers: " + err.Error(),
		})
	}

	// ตอบกลับ
	return c.JSON(fiber.Map{
		"success": true,
		"data":    stickers,
	})
}

// GetUserFavoriteStickers ดึงสติกเกอร์โปรดของผู้ใช้
func (h *StickerHandler) GetUserFavoriteStickers(c *fiber.Ctx) error {
	// ดึง userID จาก middleware
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึงสติกเกอร์โปรดของผู้ใช้
	stickers, err := h.stickerService.GetUserFavoriteStickers(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get user favorite stickers: " + err.Error(),
		})
	}

	// ตอบกลับ
	return c.JSON(fiber.Map{
		"success": true,
		"data":    stickers,
	})
}
