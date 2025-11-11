// domain/repository/refresh_token_repository.go
package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

type RefreshTokenRepository interface {
	Create(refreshToken *models.RefreshToken) error
	FindByToken(token string) (*models.RefreshToken, error)
	RevokeByUserID(userID uuid.UUID) error
	DeleteExpired(before time.Time) error
	// เพิ่ม method อื่นๆ ตามที่จำเป็น
}
