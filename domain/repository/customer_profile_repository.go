// domain/repository/customer_profile_repository.go

package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

type CustomerProfileRepository interface {
	Create(profile *models.CustomerProfile) error
	GetByBusinessAndUser(businessID, userID uuid.UUID) (*models.CustomerProfile, error)
	Update(profile *models.CustomerProfile) error
	GetByBusinessID(businessID uuid.UUID, limit, offset int) ([]*models.CustomerProfile, int64, error)
	SearchByBusinessID(businessID uuid.UUID, query string, limit, offset int) ([]*models.CustomerProfile, int64, error)
	FindByConditions(businessID uuid.UUID, customerTypes []string, lastContactFrom, lastContactTo *time.Time, statuses []string, customQuery map[string]interface{}) ([]*models.CustomerProfile, error)
}
