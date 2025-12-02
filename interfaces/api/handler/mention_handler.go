package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
)

// MentionHandler handles mention-related requests
type MentionHandler struct {
	mentionRepo repository.MessageMentionRepository
}

// NewMentionHandler creates a new mention handler
func NewMentionHandler(mentionRepo repository.MessageMentionRepository) *MentionHandler {
	return &MentionHandler{
		mentionRepo: mentionRepo,
	}
}

// GetMyMentions ดึงรายการข้อความที่ mention ตัวเอง (CURSOR-BASED)
func (h *MentionHandler) GetMyMentions(c *fiber.Ctx) error {
	userID, err := middleware.GetUserUUID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized: " + err.Error(),
		})
	}

	// Cursor pagination parameters
	limit := c.QueryInt("limit", 20)
	if limit > 100 {
		limit = 100
	}

	cursor := c.Query("cursor")
	var cursorPtr *string
	if cursor != "" {
		cursorPtr = &cursor
	}

	direction := c.Query("direction", "before")
	if direction != "before" && direction != "after" {
		direction = "before"
	}

	// Get mentions
	mentions, nextCursor, hasMore, err := h.mentionRepo.GetByUserID(
		userID,
		limit,
		cursorPtr,
		direction,
	)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		if err.Error() == "invalid cursor" || err.Error() == "cursor not found" {
			statusCode = fiber.StatusBadRequest
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get mentions: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"mentions": mentions,
			"cursor":   nextCursor,
			"has_more": hasMore,
		},
	})
}
