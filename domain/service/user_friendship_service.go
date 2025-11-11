// domain/service/user_friendship_service.go
package service

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

type UserFriendshipService interface {
	// ฟีเจอร์หลักของระบบเพื่อน
	SendFriendRequest(userID, friendID uuid.UUID) (*models.UserFriendship, error)
	AcceptFriendRequest(requestID, userID uuid.UUID) (*models.UserFriendship, error)
	RejectFriendRequest(requestID, userID uuid.UUID) (*models.UserFriendship, error)
	RemoveFriend(userID, friendID uuid.UUID) error
	GetFriends(userID uuid.UUID) ([]*models.User, error)
	GetPendingRequests(userID uuid.UUID) ([]*models.UserFriendship, error)

	// ฟีเจอร์การบล็อก
	BlockUser(userID, targetID uuid.UUID) error
	UnblockUser(userID, targetID uuid.UUID) error
	GetBlockedUsers(userID uuid.UUID) ([]*models.User, error)

	// ตรวจสอบความสัมพันธ์
	GetFriendshipStatus(userID, otherUserID uuid.UUID) (string, uuid.UUID, error) // returns status, friendshipID, error
	IsFriend(userID, otherUserID uuid.UUID) (bool, error)
	HasPendingRequest(userID, otherUserID uuid.UUID) (bool, string, error) // returns exists, direction, error
}
