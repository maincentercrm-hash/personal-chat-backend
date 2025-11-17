// pkg/di/container.go
package di

import (
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/thizplus/gofiber-chat-api/application/serviceimpl"
	"github.com/thizplus/gofiber-chat-api/domain/port"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/infrastructure/adapter"
	"github.com/thizplus/gofiber-chat-api/infrastructure/persistence/postgres"
	"github.com/thizplus/gofiber-chat-api/interfaces/api/handler"
	"github.com/thizplus/gofiber-chat-api/interfaces/websocket"

	"gorm.io/gorm"
)

// Container เก็บ dependencies ทั้งหมดของแอปพลิเคชัน
type Container struct {
	// Repositories
	UserRepo                   repository.UserRepository
	RefreshTokenRepo           repository.RefreshTokenRepository
	TokenBlacklistRepo         repository.TokenBlacklistRepository
	UserFriendshipRepo         repository.UserFriendshipRepository
	BusinessAccountRepo        repository.BusinessAccountRepository
	BusinessAdminRepo          repository.BusinessAdminRepository
	ConversationRepo           repository.ConversationRepository
	ConversationMemberRepo     repository.ConversationMemberRepository
	MessageRepo                repository.MessageRepository
	MessageReadRepo            repository.MessageReadRepository
	AnalyticsRepo              repository.AnalyticsDailyRepository
	BusinessFollowRepo         repository.BusinessFollowRepository
	CustomerProfileRepo        repository.CustomerProfileRepository
	TagRepo                    repository.TagRepository
	UserTagRepo                repository.UserTagRepository
	StickerRepo                repository.StickerRepository
	BusinessWelcomeMessageRepo repository.BusinessWelcomeMessageRepository
	BroadcastRepo              repository.BroadcastRepository
	BroadcastDeliveryRepo      repository.BroadcastDeliveryRepository

	// WebSocket Components
	WebSocketHub  *websocket.Hub
	WebSocketPort port.WebSocketPort

	// Services
	StorageService                service.FileStorageService
	AuthService                   service.AuthService
	UserService                   service.UserService
	UserFriendshipService         service.UserFriendshipService
	BusinessAccountService        service.BusinessAccountService
	BusinessAdminService          service.BusinessAdminService
	ConversationService           service.ConversationService
	ConversationMemberService     service.ConversationMemberService
	MessageService                service.MessageService
	MessageReadService            service.MessageReadService
	AnalyticsService              service.AnalyticsService
	BusinessFollowService         service.BusinessFollowService
	CustomerProfileService        service.CustomerProfileService
	TagService                    service.TagService
	UserTagService                service.UserTagService
	StickerService                service.StickerService
	BusinessWelcomeMessageService service.BusinessWelcomeMessageService
	BroadcastService              service.BroadcastService
	BroadcastDeliveryService      service.BroadcastDeliveryService
	NotificationService           service.NotificationService

	// Handlers
	AuthHandler                   *handler.AuthHandler
	UserHandler                   *handler.UserHandler
	FileHandler                   *handler.FileHandler
	UserFriendshipHandler         *handler.UserFriendshipHandler
	BusinessAccountHandler        *handler.BusinessAccountHandler
	BusinessAdminHandler          *handler.BusinessAdminHandler
	ConversationHandler           *handler.ConversationHandler
	ConversationMemberHandler     *handler.ConversationMemberHandler
	MessageHandler                *handler.MessageHandler
	MessageReadHandler            *handler.MessageReadHandler
	AnalyticsHandler              *handler.AnalyticsHandler
	BusinessFollowHandler         *handler.BusinessFollowHandler
	CustomerProfileHandler        *handler.CustomerProfileHandler
	TagHandler                    *handler.TagHandler
	UserTagHandler                *handler.UserTagHandler
	StickerHandler                *handler.StickerHandler
	BusinessWelcomeMessageHandler *handler.BusinessWelcomeMessageHandler
	BroadcastHandler              *handler.BroadcastHandler
	BroadcastDeliveryHandler      *handler.BroadcastDeliveryHandler
	SearchHandler                 *handler.SearchHandler

	// Scheduler
	RedisClient        *redis.Client
	BroadcastScheduler *scheduler.BroadcastScheduler
}

// NewContainer สร้าง container ใหม่พร้อมกับ dependencies ทั้งหมด
func NewContainer(db *gorm.DB, storageService service.FileStorageService, redisClient *redis.Client) (*Container, error) {
	container := &Container{
		StorageService: storageService,
		RedisClient:    redisClient,
	}

	// สร้าง repositories
	container.UserRepo = postgres.NewUserRepository(db)
	container.RefreshTokenRepo = postgres.NewRefreshTokenRepository(db)
	container.TokenBlacklistRepo = postgres.NewTokenBlacklistRepository(db)
	container.UserFriendshipRepo = postgres.NewUserFriendshipRepository(db)
	container.BusinessAccountRepo = postgres.NewBusinessAccountRepository(db)
	container.BusinessAdminRepo = postgres.NewBusinessAdminRepository(db)
	container.ConversationRepo = postgres.NewConversationRepository(db)
	container.ConversationMemberRepo = postgres.NewConversationMemberRepository(db)
	container.MessageRepo = postgres.NewMessageRepository(db)
	container.MessageReadRepo = postgres.NewMessageReadRepository(db)
	container.AnalyticsRepo = postgres.NewAnalyticsDailyRepository(db)
	container.BusinessFollowRepo = postgres.NewBusinessFollowRepository(db)
	container.CustomerProfileRepo = postgres.NewCustomerProfileRepository(db)
	container.TagRepo = postgres.NewTagRepository(db)
	container.UserTagRepo = postgres.NewUserTagRepository(db)
	container.StickerRepo = postgres.NewStickerRepository(db)
	container.BusinessWelcomeMessageRepo = postgres.NewBusinessWelcomeMessageRepository(db)
	container.BroadcastRepo = postgres.NewBroadcastRepository(db)
	container.BroadcastDeliveryRepo = postgres.NewBroadcastDeliveryRepository(db)

	log.Println("เชื่อมต่อกับบริการจัดเก็บไฟล์สำเร็จ")

	// สร้าง basic services
	container.AuthService = serviceimpl.NewAuthService(
		container.UserRepo,
		container.RefreshTokenRepo,
		container.TokenBlacklistRepo,
	)

	container.UserService = serviceimpl.NewUserService(container.UserRepo)
	container.UserFriendshipService = serviceimpl.NewUserFriendshipService(
		container.UserFriendshipRepo,
		container.UserRepo,
	)
	container.BusinessAccountService = serviceimpl.NewBusinessAccountService(
		container.BusinessAccountRepo,
		container.BusinessAdminRepo,
		container.UserRepo,
	)
	container.BusinessAdminService = serviceimpl.NewBusinessAdminService(
		container.BusinessAdminRepo,
		container.BusinessAccountRepo,
		container.UserRepo,
	)
	container.ConversationService = serviceimpl.NewConversationService(
		container.ConversationRepo,
		container.UserRepo,
		container.BusinessAccountRepo,
		container.MessageRepo,
		container.BusinessAdminRepo,
		container.CustomerProfileRepo,
	)
	container.ConversationMemberService = serviceimpl.NewConversationMemberService(
		container.ConversationRepo,
		container.UserRepo,
		container.MessageRepo,
	)

	container.MessageReadService = serviceimpl.NewMessageReadService(
		container.MessageRepo,
		container.MessageReadRepo,
		container.ConversationRepo,
	)

	container.AnalyticsService = serviceimpl.NewAnalyticsService(
		container.AnalyticsRepo,
		container.BusinessAccountRepo,
		container.BusinessAdminRepo,
	)
	container.BusinessFollowService = serviceimpl.NewBusinessFollowService(
		container.BusinessFollowRepo,
		container.BusinessAccountRepo,
		container.UserRepo,
		container.AnalyticsRepo,
	)

	container.CustomerProfileService = serviceimpl.NewCustomerProfileService(
		container.CustomerProfileRepo,
		container.BusinessAdminRepo,
		container.BusinessAccountRepo,
		container.UserRepo,
		container.UserTagRepo,
	)
	container.TagService = serviceimpl.NewTagService(
		container.TagRepo,
		container.UserTagRepo,
		container.CustomerProfileRepo,
		container.BusinessAdminRepo,
		container.BusinessAccountRepo,
	)
	container.UserTagService = serviceimpl.NewUserTagService(
		container.UserTagRepo,
		container.TagRepo,
		container.CustomerProfileRepo,
		container.BusinessAdminRepo,
		container.BusinessAccountRepo,
		container.UserRepo,
	)
	container.StickerService = serviceimpl.NewStickerService(
		container.StickerRepo,
		container.StorageService,
	)

	container.MessageService = serviceimpl.NewMessageService(
		container.MessageRepo,
		container.MessageReadRepo,
		container.ConversationRepo,
		container.UserRepo,
		container.BusinessAccountRepo,
		container.BusinessAdminRepo,
	)

	container.BusinessWelcomeMessageService = serviceimpl.NewBusinessWelcomeMessageService(
		container.BusinessWelcomeMessageRepo,
		container.BusinessAccountRepo,
		container.UserRepo,
		container.BusinessAdminRepo,
		container.MessageService,
		container.ConversationRepo,
	)
	container.BroadcastService = serviceimpl.NewBroadcastService(
		container.BroadcastRepo,
		container.BroadcastDeliveryRepo,
		container.BusinessAccountRepo,
		container.BusinessAdminRepo,
		container.UserRepo,
		container.BusinessFollowRepo,
		container.UserTagRepo,
		container.TagRepo,
		container.CustomerProfileRepo,
		container.MessageService,
		container.UserTagService,
		container.ConversationRepo,
	)
	container.BroadcastDeliveryService = serviceimpl.NewBroadcastDeliveryService(
		container.BroadcastDeliveryRepo,
	)

	container.BroadcastScheduler = scheduler.NewBroadcastScheduler(
		redisClient,
		container.BroadcastRepo,
		container.BroadcastService,
	)

	// สร้าง WebSocket Hub ที่มีเฉพาะ services ที่จำเป็น
	container.WebSocketHub = websocket.NewHub(
		container.ConversationService,
		container.BusinessAdminService,
		nil, // NotificationService จะถูกตั้งค่าภายหลัง
	)

	// สร้าง WebSocketAdapter
	container.WebSocketPort = adapter.NewWebSocketAdapter(container.WebSocketHub)

	// สร้าง NotificationService
	container.NotificationService = serviceimpl.NewNotificationService(
		container.WebSocketPort,
		container.UserRepo,
		container.MessageRepo,
		container.ConversationRepo,
		container.BusinessAccountRepo,
	)

	// ตั้งค่า NotificationService ใน Hub
	container.WebSocketHub.SetNotificationService(container.NotificationService)

	// สร้าง handlers
	container.AuthHandler = handler.NewAuthHandler(container.AuthService)
	container.UserHandler = handler.NewUserHandler(container.UserService, container.AuthService, container.StorageService)
	container.FileHandler = handler.NewFileHandler(container.StorageService)
	container.UserFriendshipHandler = handler.NewUserFriendshipHandler(container.UserFriendshipService, container.UserService, container.ConversationMemberService, container.NotificationService)
	container.BusinessAccountHandler = handler.NewBusinessAccountHandler(container.BusinessAccountService, container.StorageService)
	container.BusinessAdminHandler = handler.NewBusinessAdminHandler(container.BusinessAdminService)
	container.ConversationHandler = handler.NewConversationHandler(container.ConversationService, container.BusinessAdminService, container.BusinessWelcomeMessageService, container.NotificationService)
	container.ConversationMemberHandler = handler.NewConversationMemberHandler(container.ConversationMemberService)
	container.MessageHandler = handler.NewMessageHandler(container.MessageService, container.NotificationService)
	container.MessageReadHandler = handler.NewMessageReadHandler(container.MessageReadService, container.NotificationService)
	container.AnalyticsHandler = handler.NewAnalyticsHandler(container.AnalyticsService)
	container.BusinessFollowHandler = handler.NewBusinessFollowHandler(container.BusinessFollowService)
	container.CustomerProfileHandler = handler.NewCustomerProfileHandler(container.CustomerProfileService, container.NotificationService)
	container.TagHandler = handler.NewTagHandler(container.TagService, container.UserTagService, container.NotificationService)
	container.UserTagHandler = handler.NewUserTagHandler(container.UserTagService)
	container.StickerHandler = handler.NewStickerHandler(container.StickerService)
	container.BusinessWelcomeMessageHandler = handler.NewBusinessWelcomeMessageHandler(container.BusinessWelcomeMessageService)
	container.BroadcastHandler = handler.NewBroadcastHandler(container.BroadcastService, container.BroadcastScheduler)
	container.BroadcastDeliveryHandler = handler.NewBroadcastDeliveryHandler(container.BroadcastDeliveryService)
	container.SearchHandler = handler.NewSearchHandler(container.UserService, container.UserFriendshipService, container.BusinessAccountService)

	return container, nil
}
