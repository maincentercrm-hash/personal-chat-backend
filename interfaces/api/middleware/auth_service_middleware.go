// interfaces/api/middleware/auth_service_middleware.go
package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/domain/service"
)

// AuthMiddleware struct เพื่อใช้ service
type AuthMiddleware struct {
	authService service.AuthService
}

// NewAuthMiddleware สร้าง auth middleware instance
func NewAuthMiddleware(authService service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// Protected เป็น method ที่ return Protected middleware function
func (am *AuthMiddleware) Protected() fiber.Handler {
	return Protected() // ใช้ Protected function ที่มีอยู่แล้ว
}
