// domain/dto/activity_dto.go
package dto

import (
	"time"

	"github.com/thizplus/gofiber-chat-api/domain/types"
)

// ActivityDTO represents a group activity
type ActivityDTO struct {
	ID             string       `json:"id"`
	ConversationID string       `json:"conversation_id"`
	Type           string       `json:"type"`
	Actor          *UserInfoDTO `json:"actor"`
	Target         *UserInfoDTO `json:"target,omitempty"`
	OldValue       types.JSONB  `json:"old_value,omitempty"`
	NewValue       types.JSONB  `json:"new_value,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
}

// UserInfoDTO represents basic user information for activities
type UserInfoDTO struct {
	ID              string `json:"id"`
	Username        string `json:"username"`
	DisplayName     string `json:"display_name"`
	ProfileImageURL string `json:"profile_image_url,omitempty"`
}
