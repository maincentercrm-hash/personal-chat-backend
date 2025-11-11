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
	if hub.businessAdminService == nil {
		log.Println("Warning: BusinessAdminService is nil, business WebSocket endpoint might not work properly")
	}
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

	// Business WebSocket endpoint - เพิ่มการตรวจสอบว่า businessAdminService พร้อมใช้งานหรือไม่
	app.Get("/ws/business/:businessId", func(c *fiber.Ctx) error {
		log.Printf("Auth middleware for /ws/business/%s called", c.Params("businessId"))

		// ดึง token จาก query parameter หรือ Authorization header
		token := c.Query("token")

		if token == "" {
			authHeader := c.Get("Authorization")
			if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				token = authHeader[7:]
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
		log.Printf("Business WebSocket handler called")

		// ดึง userUUID จาก locals
		userUUID, ok := c.Locals("userUUID").(uuid.UUID)
		if !ok {
			log.Printf("Error: userUUID not found in locals")
			c.WriteMessage(websocket.CloseMessage, []byte("User not authenticated"))
			c.Close()
			return
		}

		// ดึง business ID จาก params
		businessID, err := uuid.Parse(c.Params("businessId"))
		if err != nil {
			log.Printf("Invalid business ID: %s", c.Params("businessId"))
			c.WriteMessage(websocket.CloseMessage, []byte("Invalid business ID"))
			c.Close()
			return
		}

		// ตรวจสอบว่า businessAdminService พร้อมใช้งานหรือไม่
		if hub.businessAdminService == nil {
			log.Printf("BusinessAdminService is nil, cannot check admin permission")
			c.WriteMessage(websocket.CloseMessage, []byte("Service unavailable"))
			c.Close()
			return
		}

		// Verify user is admin of business
		isAdmin, err := hub.businessAdminService.CheckAdminPermission(
			userUUID,
			businessID,
			[]string{"owner", "admin", "moderator"},
		)
		if err != nil || !isAdmin {
			log.Printf("User %s not authorized for business %s", userUUID, businessID)
			c.WriteMessage(websocket.CloseMessage, []byte("Not authorized for this business"))
			c.Close()
			return
		}

		log.Printf("Business WebSocket connection established for user: %s, business: %s",
			userUUID.String(), businessID.String())

		// สร้าง client
		client := &Client{
			ID:           uuid.New(),
			UserID:       userUUID,
			BusinessID:   &businessID,
			Conn:         c,
			Send:         make(chan []byte, 256),
			Hub:          hub,
			IsAlive:      true,
			LastPingTime: time.Now(),
			RateLimiter:  NewRateLimiter(120, time.Minute),
			messageCount: 0,
			lastReset:    time.Now(),
		}

		log.Printf("Registering business client %s for user %s, business %s",
			client.ID, userUUID.String(), businessID.String())
		hub.register <- client

		// Start goroutines
		go client.WritePump()
		client.ReadPump()

		log.Printf("Business WebSocket connection closed")
	}))

	// WebSocket stats endpoint
	app.Get("/api/v1/websocket/stats", middleware.Protected(), func(c *fiber.Ctx) error {
		stats := hub.GetStats()
		return c.JSON(fiber.Map{
			"success": true,
			"data":    stats,
		})
	})

	// WebSocket health check
	app.Get("/api/v1/websocket/health", func(c *fiber.Ctx) error {
		stats := hub.GetStats()
		healthy := stats["total_connections"].(int) < 10000 // Max 10k connections

		status := fiber.StatusOK
		if !healthy {
			status = fiber.StatusServiceUnavailable
		}

		return c.Status(status).JSON(fiber.Map{
			"healthy":     healthy,
			"connections": stats["total_connections"],
			"uptime":      stats["uptime"],
		})
	})

	// WebSocket connections status endpoint
	app.Get("/api/v1/websocket/connections", middleware.Protected(), func(c *fiber.Ctx) error {
		// ดึงรายละเอียดการเชื่อมต่อทั้งหมด
		connections := hub.GetAllConnections()
		return c.JSON(fiber.Map{
			"success": true,
			"data":    connections,
		})
	})

	// WebSocket conversation subscribers endpoint
	app.Get("/api/v1/websocket/conversations/:conversationId/subscribers", middleware.Protected(), func(c *fiber.Ctx) error {
		conversationID, err := uuid.Parse(c.Params("conversationId"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid conversation ID format",
			})
		}

		subscribers := hub.GetConversationSubscribers(conversationID)
		return c.JSON(fiber.Map{
			"success": true,
			"data":    subscribers,
		})
	})

	app.Get("/api/v1/websocket/users/:userId/subscribers", middleware.Protected(), func(c *fiber.Ctx) error {
		userID, err := uuid.Parse(c.Params("userId"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid user ID format",
			})
		}

		subscribers := hub.GetUserStatusSubscribers(userID)
		return c.JSON(fiber.Map{
			"success": true,
			"data":    subscribers,
		})
	})

	log.Println("WebSocket routes registered successfully")
}
