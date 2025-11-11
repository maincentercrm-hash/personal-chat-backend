// interfaces/api/handler/file_handler.go
package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/domain/service"
)

// FileHandler จัดการ API endpoints เกี่ยวกับการอัปโหลดไฟล์
type FileHandler struct {
	storageService service.FileStorageService
}

// NewFileHandler สร้าง FileHandler ใหม่
func NewFileHandler(storageService service.FileStorageService) *FileHandler {
	return &FileHandler{
		storageService: storageService,
	}
}

// UploadImage จัดการการอัปโหลดรูปภาพ
func (h *FileHandler) UploadImage(c *fiber.Ctx) error {
	// รับไฟล์จาก request
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ไม่พบไฟล์รูปภาพในคำขอ",
		})
	}

	// กำหนด folder สำหรับเก็บรูปภาพ (ถ้ามีการส่งมา)
	folder := c.FormValue("folder", "images")

	// อัปโหลดรูปภาพโดยใช้ storage service
	result, err := h.storageService.UploadImage(file, folder)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "ไม่สามารถอัปโหลดรูปภาพได้: " + err.Error(),
		})
	}

	// ส่งผลลัพธ์กลับไป
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "อัปโหลดรูปภาพสำเร็จ",
		"data":    result,
	})
}

// UploadFile จัดการการอัปโหลดไฟล์ทั่วไป
func (h *FileHandler) UploadFile(c *fiber.Ctx) error {
	// รับไฟล์จาก request
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ไม่พบไฟล์ในคำขอ",
		})
	}

	// กำหนด folder สำหรับเก็บไฟล์ (ถ้ามีการส่งมา)
	folder := c.FormValue("folder", "files")

	// อัปโหลดไฟล์โดยใช้ storage service
	result, err := h.storageService.UploadFile(file, folder)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "ไม่สามารถอัปโหลดไฟล์ได้: " + err.Error(),
		})
	}

	// ส่งผลลัพธ์กลับไป
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "อัปโหลดไฟล์สำเร็จ",
		"data":    result,
	})
}
