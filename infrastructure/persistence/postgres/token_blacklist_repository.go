// /infrastructure/persistence/postgres/token_blacklist_repository.go
package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"gorm.io/gorm"
)

type tokenBlacklistRepository struct {
	db *gorm.DB
}

func NewTokenBlacklistRepository(db *gorm.DB) repository.TokenBlacklistRepository {
	return &tokenBlacklistRepository{db: db}
}

func (r *tokenBlacklistRepository) Create(blacklist *models.TokenBlacklist) error {
	if blacklist.ID == uuid.Nil {
		blacklist.ID = uuid.New()
	}
	if blacklist.CreatedAt.IsZero() {
		blacklist.CreatedAt = time.Now()
	}

	return r.db.Create(blacklist).Error
}

func (r *tokenBlacklistRepository) FindByToken(token string) (*models.TokenBlacklist, error) {
	var blacklist models.TokenBlacklist
	err := r.db.Where("token = ?", token).First(&blacklist).Error
	if err != nil {
		return nil, err
	}
	return &blacklist, nil
}

func (r *tokenBlacklistRepository) IsTokenBlacklisted(token string) (bool, error) {
	var count int64
	err := r.db.Model(&models.TokenBlacklist{}).Where("token = ?", token).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *tokenBlacklistRepository) DeleteExpired(before time.Time) error {
	return r.db.Where("expired_at < ?", before).Delete(&models.TokenBlacklist{}).Error
}
