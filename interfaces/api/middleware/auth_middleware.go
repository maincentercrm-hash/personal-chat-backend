// middleware/auth_middleware.go
package middleware

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Protected เป็น middleware สำหรับป้องกันเส้นทางที่ต้องการการยืนยันตัวตน
func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// ดึงค่า JWT secret
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = "default-jwt-secret-for-development-only"
			fmt.Println("WARNING: JWT_SECRET not set in environment, using default value")
		}

		// ดึง Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Missing JWT token",
			})
		}

		// ตรวจสอบรูปแบบ Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Malformed JWT token",
			})
		}

		tokenString := parts[1]

		// สร้าง parser ด้วย options
		parser := jwt.NewParser(jwt.WithValidMethods([]string{"HS256"}))

		// แยกวิเคราะห์และตรวจสอบ token
		token, err := parser.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid or expired JWT token",
			})
		}

		// ตรวจสอบว่า token ถูกต้อง
		if !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid JWT token",
			})
		}

		// เก็บ token ไว้ใน locals
		c.Locals("user", token)

		// ดึง user ID จาก claims และเก็บไว้ใน locals
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if userIDStr, ok := claims["id"].(string); ok {
				// เก็บ string ID เพื่อความเข้ากันได้กับโค้ดเดิม
				c.Locals("userID", userIDStr)

				// แปลงเป็น UUID และเก็บไว้ใน locals
				userUUID, err := uuid.Parse(userIDStr)
				if err == nil {
					c.Locals("userUUID", userUUID)
				} else {
					// ID ในโทเคนไม่ใช่ UUID ที่ถูกต้อง
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error":   true,
						"message": "Invalid UUID format in token",
					})
				}
			}
		}

		return c.Next()
	}
}

// GetUserIDString ดึง user ID เป็น string (สำหรับความเข้ากันได้กับโค้ดเดิม)
func GetUserIDString(c *fiber.Ctx) (string, bool) {
	userID, ok := c.Locals("userID").(string)
	return userID, ok
}

// GetUserUUID ดึง user UUID จาก context
func GetUserUUID(c *fiber.Ctx) (uuid.UUID, error) {
	// ลองดึง UUID โดยตรงจาก context ก่อน
	if userUUID, ok := c.Locals("userUUID").(uuid.UUID); ok {
		return userUUID, nil
	}

	// ถ้าไม่มี userUUID ลองดึง userID (string) และแปลง
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return uuid.Nil, fmt.Errorf("user ID not found in context")
	}

	// แปลง string เป็น UUID
	return uuid.Parse(userIDStr)
}

// GetUserUUIDOrError ดึง UUID หรือส่งกลับ error response
func GetUserUUIDOrError(c *fiber.Ctx) (uuid.UUID, error) {
	// ลองดึง UUID โดยตรงจาก context ก่อน
	if userUUID, ok := c.Locals("userUUID").(uuid.UUID); ok {
		return userUUID, nil
	}

	// ถ้าไม่มี userUUID ลองดึง userID (string)
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return uuid.Nil, c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: User ID not found",
		})
	}

	// แปลง string เป็น UUID
	parsed, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid user ID format",
		})
	}

	return parsed, nil
}

// ValidateTokenStringToUUID ตรวจสอบ token และส่งคืน UUID
func ValidateTokenStringToUUID(tokenString string) (uuid.UUID, error) {
	userIDStr, err := ValidateTokenString(tokenString)
	if err != nil {
		return uuid.Nil, err
	}

	return uuid.Parse(userIDStr)
}

// ValidateTokenString ตรวจสอบ token และส่งคืน string (เพื่อความเข้ากันได้กับโค้ดเดิม)
func ValidateTokenString(tokenString string) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-jwt-secret-for-development-only"
	}

	// สร้าง parser ด้วย options
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"HS256"}))

	// แยกวิเคราะห์และตรวจสอบ token
	token, err := parser.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return "", err
	}

	// ตรวจสอบว่า token ถูกต้องและดึงข้อมูล claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if userID, ok := claims["id"].(string); ok {
			return userID, nil
		}
		return "", fmt.Errorf("user ID not found in token claims")
	}

	return "", fmt.Errorf("invalid token")
}

// GenerateJWTFromUUID สร้าง JWT token จาก UUID
func GenerateJWTFromUUID(userUUID uuid.UUID) (string, error) {
	return GenerateJWT(userUUID.String())
}

// GenerateJWT สร้าง JWT token จาก string ID (เพื่อความเข้ากันได้กับโค้ดเดิม)
func GenerateJWT(userID string) (string, error) {
	// ตรวจสอบว่า userID เป็น UUID ที่ถูกต้อง
	_, err := uuid.Parse(userID)
	if err != nil {
		return "", fmt.Errorf("invalid UUID format: %w", err)
	}

	// ดึงค่า JWT secret
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-jwt-secret-for-development-only"
		fmt.Println("WARNING: JWT_SECRET not set in environment, using default value")
	}

	// ดึงระยะเวลาหมดอายุของ token (หน่วยเป็นนาที)
	expiry := 1440 // 24 ชั่วโมง
	if expiryString := os.Getenv("JWT_ACCESS_EXPIRY"); expiryString != "" {
		fmt.Sscanf(expiryString, "%d", &expiry)
	}

	// สร้าง claims
	now := time.Now()
	claims := jwt.MapClaims{
		"id":  userID,
		"exp": jwt.NewNumericDate(now.Add(time.Duration(expiry) * time.Minute)),
		"iat": jwt.NewNumericDate(now),
	}

	// สร้าง token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// ลงชื่อ token
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
