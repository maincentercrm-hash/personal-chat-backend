// utils/parse.go
package utils

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ParseInt แปลงสตริงเป็นจำนวนเต็ม พร้อมค่าเริ่มต้นเมื่อเกิดข้อผิดพลาด
func ParseInt(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}

	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}

	return val
}

// ParseIntWithLimit แปลงสตริงเป็นจำนวนเต็ม พร้อมจำกัดค่าสูงสุดและต่ำสุด
func ParseIntWithLimit(s string, defaultValue, min, max int) int {
	val := ParseInt(s, defaultValue)

	if val < min {
		return min
	}

	if max > 0 && val > max {
		return max
	}

	return val
}

// ParseUUID แปลง string เป็น UUID
func ParseUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}

// ParseUUIDOrError แปลง string เป็น UUID หรือส่งกลับ error response
func ParseUUIDOrError(c *fiber.Ctx, id string, errorMessage string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": errorMessage,
		})
	}
	return parsed, nil
}

// ParseUUIDParam แปลง URL parameter เป็น UUID หรือส่งกลับ error response
func ParseUUIDParam(c *fiber.Ctx, paramName string) (uuid.UUID, error) {
	idStr := c.Params(paramName)
	if idStr == "" {
		return uuid.Nil, c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": paramName + " is required",
		})
	}

	return ParseUUIDOrError(c, idStr, "Invalid "+paramName+" format")
}

// ParseUUIDQuery แปลง query parameter เป็น UUID หรือส่งกลับ error response
func ParseUUIDQuery(c *fiber.Ctx, queryName string, required bool) (uuid.UUID, error) {
	idStr := c.Query(queryName)
	if idStr == "" {
		if required {
			return uuid.Nil, c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": queryName + " is required",
			})
		}
		return uuid.Nil, nil // ถ้าไม่จำเป็นและไม่มีค่า ก็ส่งค่า uuid.Nil
	}

	return ParseUUIDOrError(c, idStr, "Invalid "+queryName+" format")
}

// ParseUUIDArray แปลง string array เป็น UUID array
func ParseUUIDArray(strIDs []string) []uuid.UUID {
	var uuids []uuid.UUID
	for _, idStr := range strIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue // ข้ามค่าที่ไม่ถูกต้อง
		}
		uuids = append(uuids, id)
	}
	return uuids
}

// ValidateUUIDArray แปลงและตรวจสอบความถูกต้องของ UUID array
func ValidateUUIDArray(c *fiber.Ctx, strIDs []string, fieldName string) ([]uuid.UUID, error) {
	uuids := ParseUUIDArray(strIDs)
	if len(uuids) == 0 {
		return nil, c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "No valid " + fieldName + " provided",
		})
	}
	return uuids, nil
}
