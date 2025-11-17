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
	ConversationRepo           repository.ConversationRepository
	ConversationMemberRepo     repository.ConversationMemberRepository
	MessageRepo                repository.MessageRepository
	MessageReadRepo            repository.MessageReadRepository
	StickerRepo                repository.StickerRepository

	// WebSocket Components
	WebSocketHub  *websocket.Hub
	WebSocketPort port.WebSocketPort

	// Services
	StorageService                service.FileStorageService
	AuthService                   service.AuthService
	UserService                   service.UserService
	UserFriendshipService         service.UserFriendshipService
	ConversationService           service.ConversationService
	ConversationMemberService     service.ConversationMemberService
	MessageService                service.MessageService
	MessageReadService            service.MessageReadService
	StickerService                service.StickerService
	NotificationService           service.NotificationService

	// Handlers
	AuthHandler                   *handler.AuthHandler
	UserHandler                   *handler.UserHandler
	FileHandler                   *handler.FileHandler
	UserFriendshipHandler         *handler.UserFriendshipHandler
	ConversationHandler           *handler.ConversationHandler
	ConversationMemberHandler     *handler.ConversationMemberHandler
	MessageHandler                *handler.MessageHandler
	MessageReadHandler            *handler.MessageReadHandler
	StickerHandler                *handler.StickerHandler
	SearchHandler                 *handler.SearchHandler

	// Scheduler
	RedisClient        *redis.Client
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
	container.ConversationRepo = postgres.NewConversationRepository(db)
	container.ConversationMemberRepo = postgres.NewConversationMemberRepository(db)
	container.MessageRepo = postgres.NewMessageRepository(db)
	container.MessageReadRepo = postgres.NewMessageReadRepository(db)
	container.StickerRepo = postgres.NewStickerRepository(db)

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
	container.ConversationService = serviceimpl.NewConversationService(
		container.ConversationRepo,
		container.UserRepo,
		container.MessageRepo,
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


	container.StickerService = serviceimpl.NewStickerService(
		container.StickerRepo,
		container.StorageService,
	)

	container.MessageService = serviceimpl.NewMessageService(
		container.MessageRepo,
		container.MessageReadRepo,
		container.ConversationRepo,
		container.UserRepo,
		container.NotificationService,
	)



	// สร้าง WebSocket Hub ที่มีเฉพาะ services ที่จำเป็น
	container.WebSocketHub = websocket.NewHub(
		container.ConversationService,
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
	)

	// ตั้งค่า NotificationService ใน Hub
	container.WebSocketHub.SetNotificationService(container.NotificationService)

	// สร้าง handlers
	container.AuthHandler = handler.NewAuthHandler(container.AuthService)
	container.UserHandler = handler.NewUserHandler(container.UserService, container.AuthService, container.StorageService)
	container.FileHandler = handler.NewFileHandler(container.StorageService)
	container.UserFriendshipHandler = handler.NewUserFriendshipHandler(container.UserFriendshipService, container.UserService, container.ConversationMemberService, container.NotificationService)
	container.ConversationHandler = handler.NewConversationHandler(container.ConversationService, container.NotificationService)
	container.ConversationMemberHandler = handler.NewConversationMemberHandler(container.ConversationMemberService, container.NotificationService)
	container.MessageHandler = handler.NewMessageHandler(container.MessageService, container.NotificationService)
	container.MessageReadHandler = handler.NewMessageReadHandler(container.MessageReadService, container.NotificationService)
	container.StickerHandler = handler.NewStickerHandler(container.StickerService)
	container.SearchHandler = handler.NewSearchHandler(container.UserService, container.UserFriendshipService)

	return container, nil
}
