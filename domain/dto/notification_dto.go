// domain/dto/notification_dto.go

package dto

import (
	"time"

	"github.com/google/uuid"
)

type FriendAcceptNotification struct {
	FriendshipID    uuid.UUID `json:"friendship_id"`
	UserID          uuid.UUID `json:"user_id"`
	Username        string    `json:"username"`
	DisplayName     string    `json:"display_name"`
	ProfileImageURL string    `json:"profile_image_url"`
	AcceptedAt      time.Time `json:"accepted_at"`
}
