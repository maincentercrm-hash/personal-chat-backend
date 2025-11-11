// domain/service/notification_service.go
package service

import (
	"github.com/google/uuid"
)

// WebSocketNotifier interface สำหรับส่ง real-time notifications
type NotificationService interface {
	// Message notifications
	NotifyNewMessage(conversationID uuid.UUID, message interface{})
	NotifyMessageRead(conversationID uuid.UUID, message interface{})
	NotifyMessageReadAll(conversationID uuid.UUID, message interface{})
	NotifyMessageEdited(conversationID uuid.UUID, message interface{})
	NotifyMessageReply(conversationID uuid.UUID, message interface{})
	NotifyMessageDeleted(conversationID uuid.UUID, messageID uuid.UUID)
	NotifyMessageReaction(conversationID uuid.UUID, reaction interface{})

	// Conversation notifications
	NotifyConversationCreated(userIDs []uuid.UUID, conversation interface{}) error
	NotifyConversationUpdated(conversationID uuid.UUID, update interface{})
	NotifyConversationDeleted(conversationID uuid.UUID, memberIDs []uuid.UUID)
	NotifyUserAddedToConversation(conversationID uuid.UUID, userID uuid.UUID)
	NotifyUserRemovedFromConversation(userID, conversationID uuid.UUID)
	NotifyNewConversation(conversation interface{}) error

	// Business notifications
	NotifyBusinessBroadcast(userIDs []uuid.UUID, broadcast interface{})
	NotifyBusinessNewFollower(businessID, followerID uuid.UUID)
	NotifyBusinessWelcomeMessage(userID, businessID uuid.UUID, message interface{})
	NotifyBusinessFollowStatusChanged(businessID, userID uuid.UUID, isFollowing bool)
	NotifyBusinessStatusChanged(businessID uuid.UUID, status string)

	// Customer Profile notifications
	NotifyProfileUpdate(businessID, userID uuid.UUID, profile interface{})
	NotifyProfileUpdateTags(businessID, userID uuid.UUID, tagId uuid.UUID, action string)

	// Friend notifications
	NotifyFriendRequestReceived(request interface{}) error
	NotifyFriendRequestAccepted(friendship interface{}) error
	NotifyFriendRemoved(userID, friendID uuid.UUID)

	// User notifications
	NotifyUserBlocked(blockerID, blockedID uuid.UUID)
	NotifyUserUnblocked(unblockerID, unblockedID uuid.UUID)

	// General notifications
	SendNotification(userIDs []uuid.UUID, notification interface{})
	SendAlert(userID uuid.UUID, alert interface{})
	NotifySystemMessage(userIDs []uuid.UUID, message interface{})
}
