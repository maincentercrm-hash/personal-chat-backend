// application/serviceimpl/presence_service.go
package serviceimpl

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"github.com/thizplus/gofiber-chat-api/domain/service"
)

type presenceService struct {
	redis              *redis.Client
	userRepo           repository.UserRepository
	userFriendshipRepo repository.UserFriendshipRepository
	ctx                context.Context
}

const (
	// Redis key prefixes
	onlineKeyPrefix = "user:online:"
	lastSeenPrefix  = "user:lastseen:"

	// TTL for online status (5 minutes)
	onlineTTL = 5 * time.Minute
)

// NewPresenceService creates a new PresenceService
func NewPresenceService(
	redis *redis.Client,
	userRepo repository.UserRepository,
	userFriendshipRepo repository.UserFriendshipRepository,
) service.PresenceService {
	return &presenceService{
		redis:              redis,
		userRepo:           userRepo,
		userFriendshipRepo: userFriendshipRepo,
		ctx:                context.Background(),
	}
}

// SetUserOnline marks a user as online in Redis
func (s *presenceService) SetUserOnline(userID uuid.UUID) error {
	key := onlineKeyPrefix + userID.String()

	// Set with TTL
	err := s.redis.Set(s.ctx, key, "1", onlineTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to set user online: %w", err)
	}

	// Update last active in database
	return s.UpdateLastActive(userID)
}

// SetUserOffline marks a user as offline
func (s *presenceService) SetUserOffline(userID uuid.UUID) error {
	key := onlineKeyPrefix + userID.String()

	// Delete online key
	err := s.redis.Del(s.ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to set user offline: %w", err)
	}

	// Store last seen time
	lastSeenKey := lastSeenPrefix + userID.String()
	now := time.Now().Unix()
	err = s.redis.Set(s.ctx, lastSeenKey, now, 0).Err() // No expiry
	if err != nil {
		return fmt.Errorf("failed to store last seen: %w", err)
	}

	// Update last active in database
	return s.UpdateLastActive(userID)
}

// UpdateLastActive updates user's last active timestamp in database
func (s *presenceService) UpdateLastActive(userID uuid.UUID) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	now := time.Now()
	user.LastActiveAt = &now

	err = s.userRepo.Update(user)
	if err != nil {
		return fmt.Errorf("failed to update last active: %w", err)
	}

	return nil
}

// IsUserOnline checks if a user is online
func (s *presenceService) IsUserOnline(userID uuid.UUID) (bool, error) {
	key := onlineKeyPrefix + userID.String()

	val, err := s.redis.Get(s.ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check online status: %w", err)
	}

	return val == "1", nil
}

// GetUserPresence gets a user's presence information
func (s *presenceService) GetUserPresence(userID uuid.UUID) (*service.UserPresence, error) {
	// Check online status from Redis
	isOnline, err := s.IsUserOnline(userID)
	if err != nil {
		return nil, err
	}

	presence := &service.UserPresence{
		UserID:   userID,
		IsOnline: isOnline,
	}

	// Get user from database for last_active_at
	user, err := s.userRepo.FindByID(userID)
	if err == nil {
		presence.LastActiveAt = user.LastActiveAt
	}

	// If offline, get last seen from Redis
	if !isOnline {
		lastSeenKey := lastSeenPrefix + userID.String()
		lastSeenUnix, err := s.redis.Get(s.ctx, lastSeenKey).Int64()
		if err == nil {
			lastSeen := time.Unix(lastSeenUnix, 0)
			presence.LastSeenAt = &lastSeen
		}
	}

	return presence, nil
}

// GetMultipleUserPresence gets presence for multiple users
func (s *presenceService) GetMultipleUserPresence(userIDs []uuid.UUID) (map[uuid.UUID]*service.UserPresence, error) {
	result := make(map[uuid.UUID]*service.UserPresence)

	// Build Redis keys
	keys := make([]string, len(userIDs))
	for i, userID := range userIDs {
		keys[i] = onlineKeyPrefix + userID.String()
	}

	// Get all online statuses in one call (pipeline)
	pipe := s.redis.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))
	for i, key := range keys {
		cmds[i] = pipe.Get(s.ctx, key)
	}
	_, _ = pipe.Exec(s.ctx)

	// Process results
	for i, cmd := range cmds {
		userID := userIDs[i]
		isOnline := false

		val, err := cmd.Result()
		if err == nil && val == "1" {
			isOnline = true
		}

		result[userID] = &service.UserPresence{
			UserID:   userID,
			IsOnline: isOnline,
		}
	}

	// Get last_active_at from database for all users
	// Note: We could optimize this with a FindByIDs method, but for now iterate
	for _, userID := range userIDs {
		user, err := s.userRepo.FindByID(userID)
		if err == nil {
			if presence, exists := result[userID]; exists {
				presence.LastActiveAt = user.LastActiveAt
			}
		}
	}

	return result, nil
}

// GetOnlineUsers gets all online users
func (s *presenceService) GetOnlineUsers() ([]uuid.UUID, error) {
	// Scan for all online keys
	pattern := onlineKeyPrefix + "*"
	var cursor uint64
	var onlineUsers []uuid.UUID

	for {
		keys, nextCursor, err := s.redis.Scan(s.ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan online users: %w", err)
		}

		for _, key := range keys {
			// Extract UUID from key
			userIDStr := key[len(onlineKeyPrefix):]
			userID, err := uuid.Parse(userIDStr)
			if err == nil {
				onlineUsers = append(onlineUsers, userID)
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return onlineUsers, nil
}

// GetOnlineFriends gets online friends of a user
func (s *presenceService) GetOnlineFriends(userID uuid.UUID) ([]*service.UserPresence, error) {
	// Get accepted friendships
	friendships, err := s.userFriendshipRepo.FindAcceptedFriendships(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get friendships: %w", err)
	}

	// Extract friend IDs
	var friendIDs []uuid.UUID
	for _, friendship := range friendships {
		if friendship.UserID == userID {
			friendIDs = append(friendIDs, friendship.FriendID)
		} else {
			friendIDs = append(friendIDs, friendship.UserID)
		}
	}

	if len(friendIDs) == 0 {
		return []*service.UserPresence{}, nil
	}

	// Get presence for all friends
	presenceMap, err := s.GetMultipleUserPresence(friendIDs)
	if err != nil {
		return nil, err
	}

	// Filter only online friends
	var onlineFriends []*service.UserPresence
	for _, friendID := range friendIDs {
		if presence, exists := presenceMap[friendID]; exists && presence.IsOnline {
			onlineFriends = append(onlineFriends, presence)
		}
	}

	return onlineFriends, nil
}
