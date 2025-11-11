// interfaces/api/routes/routes.go - คงไว้แบบเดิม ไม่ต้องแก้ไข
package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/handler"
)

// SetupRoutes กำหนดเส้นทาง API ทั้งหมดของแอปพลิเคชัน
func SetupRoutes(
	app *fiber.App,
	authHandler *handler.AuthHandler,
	fileHandler *handler.FileHandler,
	userFriendshipHandler *handler.UserFriendshipHandler,
	businessAccountHandler *handler.BusinessAccountHandler,
	businessFollowHandler *handler.BusinessFollowHandler,
	businessAdminHandler *handler.BusinessAdminHandler,
	businessAdminService service.BusinessAdminService,

	userHandler *handler.UserHandler,
	conversationHandler *handler.ConversationHandler,
	conversationMemberHandler *handler.ConversationMemberHandler,

	messageHandler *handler.MessageHandler,
	messageReadHandler *handler.MessageReadHandler,

	customerProfileHandler *handler.CustomerProfileHandler,
	tagHandler *handler.TagHandler,
	userTagHandler *handler.UserTagHandler,

	analyticsHandler *handler.AnalyticsHandler,
	stickerHandler *handler.StickerHandler,

	businessWelcomeMessageHandler *handler.BusinessWelcomeMessageHandler,
	broadcastHandler *handler.BroadcastHandler,
	broadcastDeliveryHandler *handler.BroadcastDeliveryHandler,

	searchHandler *handler.SearchHandler,

) {
	// สร้าง API group
	api := app.Group("/api/v1")

	// Health check route
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "API is running",
		})
	})

	// กำหนดเส้นทางต่างๆ
	SetupAuthRoutes(api, authHandler)
	SetupFileRoutes(api, fileHandler)

	SetupBusinessAccountRoutes(api, businessAccountHandler)
	SetupBusinessAdminRoutes(api, businessAdminHandler, businessAdminService)
	SetupBusinessFollowRoutes(api, businessFollowHandler)

	SetupUserRoutes(api, userHandler)
	SetupUserFriendshipRoutes(api, userFriendshipHandler)
	SetupConversationRoutes(api, conversationHandler, conversationMemberHandler)

	SetupMessageRoutes(api, messageHandler)
	SetupMessageReadRoutes(api, messageReadHandler)

	SetupCustomerProfileRoutes(api, customerProfileHandler, businessAdminService)
	SetupTagRoutes(api, tagHandler, businessAdminService)
	SetupUserTagRoutes(api, userTagHandler, businessAdminService)

	SetupAnalyticsRoutes(api, analyticsHandler, businessAdminService)
	SetupStickerRoutes(api, stickerHandler)

	SetupBusinessConversationRoutes(api, conversationHandler, businessAdminService)
	SetupBusinessMessageRoutes(api, messageHandler, businessAdminService)

	SetupBusinessWelcomeMessageRoutes(api, businessWelcomeMessageHandler, businessAdminService)

	SetupBroadcastRoutes(api, broadcastHandler, businessAdminService)
	SetupBroadcastDeliveryRoutes(api, broadcastDeliveryHandler)

	SetupSearchRoutes(api, searchHandler)

}
