// domain/service/presence_service.go
package service

import (
	"time"

	"github.com/google/uuid"
)

// UserPresence represents a user's online presence
type UserPresence struct {
	UserID       uuid.UUID  `json:"user_id"`
	IsOnline     bool       `json:"is_online"`
	LastActiveAt *time.Time `json:"last_active_at,omitempty"`
	LastSeenAt   *time.Time `json:"last_seen_at,omitempty"` // เวลาที่เห็นครั้งล่าสุด
}

// PresenceService manages user online presence
type PresenceService interface {
	// SetUserOnline marks a user as online
	SetUserOnline(userID uuid.UUID) error

	// SetUserOffline marks a user as offline
	SetUserOffline(userID uuid.UUID) error

	// UpdateLastActive updates user's last active timestamp
	UpdateLastActive(userID uuid.UUID) error

	// IsUserOnline checks if a user is online
	IsUserOnline(userID uuid.UUID) (bool, error)

	// GetUserPresence gets a user's presence information
	GetUserPresence(userID uuid.UUID) (*UserPresence, error)

	// GetMultipleUserPresence gets presence for multiple users
	GetMultipleUserPresence(userIDs []uuid.UUID) (map[uuid.UUID]*UserPresence, error)

	// GetOnlineUsers gets all online users
	GetOnlineUsers() ([]uuid.UUID, error)

	// GetOnlineFriends gets online friends of a user
	GetOnlineFriends(userID uuid.UUID) ([]*UserPresence, error)
}
