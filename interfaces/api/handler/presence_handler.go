// interfaces/api/handler/presence_handler.go
package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
)

type PresenceHandler struct {
	presenceService service.PresenceService
}

func NewPresenceHandler(presenceService service.PresenceService) *PresenceHandler {
	return &PresenceHandler{
		presenceService: presenceService,
	}
}

// GetUserPresence gets a single user's presence
func (h *PresenceHandler) GetUserPresence(c *fiber.Ctx) error {
	userIDStr := c.Params("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid user ID",
		})
	}

	presence, err := h.presenceService.GetUserPresence(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get user presence",
			"error":   err.Error(),
		})
	}

	// Format response ตาม spec (with backward compatibility)
	status := "offline"
	if presence.IsOnline {
		status = "online"
	}

	// ใช้ LastSeenAt ถ้ามี ไม่งั้นใช้ LastActiveAt
	lastSeen := presence.LastSeenAt
	if lastSeen == nil {
		lastSeen = presence.LastActiveAt
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"user_id":        presence.UserID,
			"status":         status,                  // 🆕 เพิ่ม
			"is_online":      presence.IsOnline,       // ✅ เก็บไว้ (backward compatible)
			"last_seen":      lastSeen,                // 🆕 เพิ่ม
			"last_active_at": presence.LastActiveAt,   // ✅ เก็บไว้ (backward compatible)
		},
	})
}

// GetMultipleUserPresence gets presence for multiple users
func (h *PresenceHandler) GetMultipleUserPresence(c *fiber.Ctx) error {
	var req struct {
		UserIDs []string `json:"user_ids"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	// Parse UUIDs
	var userIDs []uuid.UUID
	for _, idStr := range req.UserIDs {
		id, err := uuid.Parse(idStr)
		if err == nil {
			userIDs = append(userIDs, id)
		}
	}

	presenceMap, err := h.presenceService.GetMultipleUserPresence(userIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get presence",
			"error":   err.Error(),
		})
	}

	// Convert map to slice with formatted response
	presences := make([]fiber.Map, 0, len(presenceMap))
	for _, presence := range presenceMap {
		status := "offline"
		if presence.IsOnline {
			status = "online"
		}

		lastSeen := presence.LastSeenAt
		if lastSeen == nil {
			lastSeen = presence.LastActiveAt
		}

		presences = append(presences, fiber.Map{
			"user_id":        presence.UserID,
			"status":         status,                  // 🆕 เพิ่ม
			"is_online":      presence.IsOnline,       // ✅ เก็บไว้
			"last_seen":      lastSeen,                // 🆕 เพิ่ม
			"last_active_at": presence.LastActiveAt,   // ✅ เก็บไว้
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    presences,
	})
}

// GetOnlineFriends gets user's online friends
func (h *PresenceHandler) GetOnlineFriends(c *fiber.Ctx) error {
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized",
		})
	}

	friends, err := h.presenceService.GetOnlineFriends(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get online friends",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    friends,
	})
}
