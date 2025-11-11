// domain/repository/token_blacklist_repository.go
package repository

import (
	"time"

	"github.com/thizplus/gofiber-chat-api/domain/models"
)

type TokenBlacklistRepository interface {
	Create(blacklist *models.TokenBlacklist) error
	FindByToken(token string) (*models.TokenBlacklist, error)
	IsTokenBlacklisted(token string) (bool, error)
	DeleteExpired(before time.Time) error
}
