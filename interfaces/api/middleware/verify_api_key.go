// interfaces/api/middleware/verify_api_key.go
package middleware

import (
	"os"

	"github.com/gofiber/fiber/v2"
)

// VerifyAPIKey ยืนยัน API key สำหรับการติดตามข้อมูล
func VerifyAPIKey() fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKey := c.Get("X-API-Key")

		if apiKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "API key is required",
			})
		}

		// ในตัวอย่างนี้ เราจะใช้ API key จาก environment variable
		// ในทางปฏิบัติจริง คุณควรใช้ระบบที่ซับซ้อนกว่านี้ เช่น API key repository
		validApiKey := os.Getenv("API_TRACK_KEY")
		if validApiKey == "" {
			validApiKey = "default-api-key-for-development"
		}

		if apiKey != validApiKey {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid API key",
			})
		}

		// อาจเพิ่มการตรวจสอบว่า API key นี้เป็นของ business ที่ระบุใน URL หรือไม่
		// โดยดึง businessId จาก params แล้วตรวจสอบกับฐานข้อมูล

		return c.Next()
	}
}
