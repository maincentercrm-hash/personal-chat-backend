// domain/service/auth_service.go

package service

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

type AuthService interface {
	Register(username, password, email, displayName string) (*models.User, string, string, error)
	Login(username, password string) (*models.User, string, string, error)
	RefreshToken(refreshToken string) (string, string, error)
	Logout(userID uuid.UUID) error                       // เปลี่ยนเป็น UUID
	BlacklistToken(userID uuid.UUID, token string) error // เปลี่ยนเป็น UUID
	GetUserByID(userID uuid.UUID) (*models.User, error)  // เปลี่ยนเป็น UUID
}
