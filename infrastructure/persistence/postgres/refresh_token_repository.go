// /infrastructure/persistence/postgres/refresh_token_repository.go
package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"gorm.io/gorm"
)

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) repository.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(refreshToken *models.RefreshToken) error {
	if refreshToken.ID == uuid.Nil {
		refreshToken.ID = uuid.New()
	}
	if refreshToken.CreatedAt.IsZero() {
		refreshToken.CreatedAt = time.Now()
	}

	return r.db.Create(refreshToken).Error
}

func (r *refreshTokenRepository) FindByToken(token string) (*models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	err := r.db.Where("token = ? AND revoked = ?", token, false).First(&refreshToken).Error
	if err != nil {
		return nil, err
	}
	return &refreshToken, nil
}

func (r *refreshTokenRepository) RevokeByUserID(userID uuid.UUID) error {
	return r.db.Model(&models.RefreshToken{}).
		Where("user_id = ?", userID).
		Update("revoked", true).Error
}

func (r *refreshTokenRepository) DeleteExpired(before time.Time) error {
	return r.db.Where("expires_at < ?", before).Delete(&models.RefreshToken{}).Error
}
