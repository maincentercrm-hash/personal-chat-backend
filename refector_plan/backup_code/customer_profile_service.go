// domain/service/customer_profile_service.go
package service

import (
	"github.com/google/uuid"

	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/types"
)

type CustomerProfileService interface {
	CreateCustomerProfile(businessID, userID uuid.UUID, nickname, notes, customerType string) (*models.CustomerProfile, error)
	GetCustomerProfile(businessID, userID uuid.UUID) (*models.CustomerProfile, error)
	UpdateCustomerProfile(businessID, userID, adminID uuid.UUID, updateData types.JSONB) (*models.CustomerProfile, error)
	GetBusinessCustomers(businessID uuid.UUID, limit, offset int) ([]*models.CustomerProfile, int64, error)
	SearchCustomers(businessID uuid.UUID, query string, limit, offset int) ([]*models.CustomerProfile, int64, error)
}
