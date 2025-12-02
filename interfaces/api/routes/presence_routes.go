// interfaces/api/routes/presence_routes.go
package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/handler"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
)

// SetupPresenceRoutes sets up presence/online status routes
func SetupPresenceRoutes(router fiber.Router, presenceHandler *handler.PresenceHandler) {
	presence := router.Group("/presence")
	presence.Use(middleware.Protected())

	// Get single user presence
	presence.Get("/user/:userId", presenceHandler.GetUserPresence)

	// Get multiple users presence (batch)
	presence.Post("/users", presenceHandler.GetMultipleUserPresence)

	// Get online friends
	presence.Get("/friends/online", presenceHandler.GetOnlineFriends)
}
