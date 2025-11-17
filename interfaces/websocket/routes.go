// interfaces/websocket/routes.go
package websocket

import (
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/middleware"
)

// RegisterWebSocketRoutes registers WebSocket routes
func RegisterWebSocketRoutes(app *fiber.App, hub *Hub, authHandler fiber.Handler) {
	log.Println("Registering WebSocket routes...")

	// ตรวจสอบว่า Hub มี services ที่จำเป็นหรือไม่
	if hub.conversationService == nil {
		log.Println("Warning: ConversationService is nil, auto-subscription to conversations might not work properly")
	}

	// WebSocket upgrade middleware
	app.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client requested upgrade to the WebSocket protocol
		if websocket.IsWebSocketUpgrade(c) {
			log.Printf("WebSocket upgrade request detected for path: %s", c.Path())
			c.Locals("allowed", true)
			return c.Next()
		}

		log.Printf("Not a WebSocket upgrade request for path: %s", c.Path())
		return fiber.ErrUpgradeRequired
	})

	// User WebSocket endpoint
	app.Get("/ws/user", func(c *fiber.Ctx) error {
		log.Printf("Auth middleware for /ws/user called")

		// ดึง token จาก query parameter หรือ Authorization header
		token := c.Query("token")

		if token == "" {
			authHeader := c.Get("Authorization")
			if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				token = authHeader[7:]
				log.Printf("Using token from Authorization header")
			}
		}

		if token == "" {
			log.Printf("Missing JWT token")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "Missing JWT token",
			})
		}

		// Validate token
		userUUID, err := middleware.ValidateTokenStringToUUID(token)
		if err != nil {
			log.Printf("Token validation error: %v", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid or expired JWT token",
			})
		}

		log.Printf("Authentication successful for user: %s", userUUID.String())

		// Store user info in locals
		c.Locals("userID", userUUID.String())
		c.Locals("userUUID", userUUID)

		return c.Next()
	}, websocket.New(func(c *websocket.Conn) {
		log.Printf("User WebSocket handler called")

		// ดึง userUUID จาก locals
		userUUID, ok := c.Locals("userUUID").(uuid.UUID)
		if !ok {
			log.Printf("Error: userUUID not found in locals")
			c.WriteMessage(websocket.CloseMessage, []byte("User not authenticated"))
			c.Close()
			return
		}

		log.Printf("WebSocket connection established for user: %s", userUUID.String())

		// สร้าง client
		client := &Client{
			ID:           uuid.New(),
			UserID:       userUUID,
			Conn:         c,
			Send:         make(chan []byte, 256),
			Hub:          hub,
			IsAlive:      true,
			LastPingTime: time.Now(),
			RateLimiter:  NewRateLimiter(60, time.Minute),
			messageCount: 0,
			lastReset:    time.Now(),
		}

		log.Printf("Registering client %s for user %s", client.ID, userUUID.String())
		hub.register <- client

		// Start goroutines
		go client.WritePump()
		client.ReadPump()

		log.Printf("User WebSocket connection closed")
	}))

}
