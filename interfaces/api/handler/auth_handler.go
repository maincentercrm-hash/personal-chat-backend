// interfaces/api/handler/auth_handlers.go
package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	// รับข้อมูลการลงทะเบียนจาก request body
	var input map[string]string
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
	}

	// เรียกใช้ service
	user, accessToken, refreshToken, err := h.authService.Register(
		input["username"],
		input["password"],
		input["email"],
		input["display_name"],
	)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	// ส่งข้อมูลกลับ
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success":       true,
		"message":       "User registered successfully",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": fiber.Map{
			"id":           user.ID,
			"username":     user.Username,
			"email":        user.Email,
			"display_name": user.DisplayName,
			"status":       user.Status,
			"created_at":   user.CreatedAt,
		},
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	// รับข้อมูลจาก request body
	var input map[string]string
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data: " + err.Error(),
		})
	}

	// เรียกใช้ service
	user, accessToken, refreshToken, err := h.authService.Login(
		input["username"],
		input["password"],
	)

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	// ส่งข้อมูลกลับ
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success":       true,
		"message":       "Login successful",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": fiber.Map{
			"id":           user.ID,
			"username":     user.Username,
			"email":        user.Email,
			"display_name": user.DisplayName,
			"status":       user.Status,
		},
	})
}

// Logout จัดการการออกจากระบบ
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// ดึง userUUID จาก context
	userUUID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// ดึงข้อมูล token จาก header
	authHeader := c.Get("Authorization")
	tokenString := ""

	// แยกส่วน "Bearer " ออก
	if len(authHeader) > 7 && strings.HasPrefix(authHeader, "Bearer ") {
		tokenString = authHeader[7:]
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid authorization header format",
		})
	}

	// เรียกใช้ service เพื่อทำการ logout ด้วย UUID
	if err := h.authService.Logout(userUUID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error logging out: " + err.Error(),
		})
	}

	// เรียกใช้ service เพื่อ blacklist token ด้วย UUID
	if err := h.authService.BlacklistToken(userUUID, tokenString); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error blacklisting token: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Logged out successfully",
	})
}

// RefreshToken จัดการการรีเฟรช token
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var input map[string]string
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request data",
		})
	}

	refreshTokenString := input["refresh_token"]
	if refreshTokenString == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Refresh token is required",
		})
	}

	// เรียกใช้ service
	accessToken, newRefreshToken, err := h.authService.RefreshToken(refreshTokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	// ส่งข้อมูลกลับ
	return c.JSON(fiber.Map{
		"success":       true,
		"message":       "Token refreshed successfully",
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
	})
}

// GetCurrentUser ดึงข้อมูลผู้ใช้ปัจจุบัน
func (h *AuthHandler) GetCurrentUser(c *fiber.Ctx) error {
	// ดึง userUUID จาก context
	userUUID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// เรียกใช้ service ด้วย UUID
	user, err := h.authService.GetUserByID(userUUID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Error getting user data: " + err.Error(),
		})
	}

	// ส่งข้อมูลกลับ
	return c.JSON(fiber.Map{
		"success": true,
		"user": fiber.Map{
			"id":                user.ID,
			"username":          user.Username,
			"email":             user.Email,
			"display_name":      user.DisplayName,
			"profile_image_url": user.ProfileImageURL,
			"bio":               user.Bio,
			"created_at":        user.CreatedAt,
			"last_active_at":    user.LastActiveAt,
			"status":            user.Status,
		},
	})
}
