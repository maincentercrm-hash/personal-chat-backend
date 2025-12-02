package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/handler"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
)

// SetupMentionRoutes sets up routes for mention-related endpoints
func SetupMentionRoutes(router fiber.Router, mentionHandler *handler.MentionHandler) {
	mentions := router.Group("/mentions")
	mentions.Use(middleware.Protected())

	// GET /api/v1/mentions - Get my mentions (cursor-based)
	mentions.Get("", mentionHandler.GetMyMentions)
}
