# Pin Message Feature - Implementation Plan

## Overview
เพิ่มฟีเจอร์การปักหมุดข้อความ (Pin Message) ใน conversation โดยรองรับ 2 ประเภท:
- **Personal Pin**: ปักหมุดเฉพาะตัวเอง มองเห็นเฉพาะผู้ที่ปักหมุด
- **Public Pin**: ปักหมุดสาธารณะ มองเห็นได้ทุกคนใน conversation (เฉพาะ admin/owner)

## 1. Database Schema

### 1.1 สร้างตารางใหม่: `pinned_messages`

```sql
CREATE TABLE pinned_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    pin_type VARCHAR(20) NOT NULL CHECK (pin_type IN ('personal', 'public')),
    pinned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Constraints
    UNIQUE(message_id, user_id, pin_type),

    -- Indexes
    CREATE INDEX idx_pinned_messages_conversation_id ON pinned_messages(conversation_id),
    CREATE INDEX idx_pinned_messages_user_id ON pinned_messages(user_id),
    CREATE INDEX idx_pinned_messages_pin_type ON pinned_messages(pin_type),
    CREATE INDEX idx_pinned_messages_message_id ON pinned_messages(message_id)
);
```

**หมายเหตุ:**
- `user_id`: ผู้ที่ทำการปักหมุด
- `pin_type`: ประเภทการปักหมุด (personal/public)
- Unique constraint: ป้องกันการ pin message เดียวกันซ้ำ โดย user คนเดียวกันในประเภทเดียวกัน

## 2. Domain Layer

### 2.1 Model: `domain/models/pinned_message.go`

```go
package models

import (
    "time"
    "github.com/google/uuid"
)

type PinnedMessage struct {
    ID             uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
    MessageID      uuid.UUID  `json:"message_id" gorm:"type:uuid;not null"`
    ConversationID uuid.UUID  `json:"conversation_id" gorm:"type:uuid;not null"`
    UserID         uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
    PinType        string     `json:"pin_type" gorm:"type:varchar(20);not null"` // personal, public
    PinnedAt       time.Time  `json:"pinned_at" gorm:"type:timestamp with time zone;default:now()"`
    CreatedAt      time.Time  `json:"created_at" gorm:"type:timestamp with time zone;default:now()"`
    UpdatedAt      time.Time  `json:"updated_at" gorm:"type:timestamp with time zone;default:now()"`

    // Associations
    Message      *Message      `json:"message,omitempty" gorm:"foreignkey:MessageID"`
    Conversation *Conversation `json:"conversation,omitempty" gorm:"foreignkey:ConversationID"`
    User         *User         `json:"user,omitempty" gorm:"foreignkey:UserID"`
}

func (PinnedMessage) TableName() string {
    return "pinned_messages"
}

// Constants
const (
    PinTypePersonal = "personal"
    PinTypePublic   = "public"
)
```

### 2.2 DTO: `domain/dto/pinned_message_dto.go`

```go
package dto

import (
    "time"
    "github.com/google/uuid"
)

// ============ Request DTOs ============

// PinMessageRequest สำหรับการปักหมุดข้อความ
type PinMessageRequest struct {
    PinType string `json:"pin_type" validate:"required,oneof=personal public"`
}

// UnpinMessageRequest สำหรับการยกเลิกปักหมุดข้อความ
type UnpinMessageRequest struct {
    PinType string `json:"pin_type" validate:"required,oneof=personal public"`
}

// GetPinnedMessagesRequest สำหรับการดึงข้อมูลข้อความที่ปักหมุด
type GetPinnedMessagesRequest struct {
    PinType string `json:"pin_type,omitempty" validate:"omitempty,oneof=personal public all"`
    Limit   int    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
    Offset  int    `json:"offset,omitempty" validate:"omitempty,min=0"`
}

// ============ Response DTOs ============

// PinnedMessageDTO ข้อมูลข้อความที่ปักหมุด
type PinnedMessageDTO struct {
    ID             uuid.UUID   `json:"id"`
    MessageID      uuid.UUID   `json:"message_id"`
    ConversationID uuid.UUID   `json:"conversation_id"`
    UserID         uuid.UUID   `json:"user_id"`
    PinType        string      `json:"pin_type"`
    PinnedAt       time.Time   `json:"pinned_at"`
    PinnedBy       *UserBasicDTO `json:"pinned_by,omitempty"`
    Message        *MessageDTO `json:"message,omitempty"`
}

// PinnedMessagesListDTO รายการข้อความที่ปักหมุด
type PinnedMessagesListDTO struct {
    ConversationID uuid.UUID          `json:"conversation_id"`
    Total          int                `json:"total"`
    PinnedMessages []PinnedMessageDTO `json:"pinned_messages"`
}

// PinMessageResponse สำหรับผลลัพธ์การปักหมุดข้อความ
type PinMessageResponse struct {
    GenericResponse
    Data PinnedMessageDTO `json:"data"`
}

// PinnedMessagesResponse สำหรับผลลัพธ์การดึงข้อมูลข้อความที่ปักหมุด
type PinnedMessagesResponse struct {
    GenericResponse
    Data PinnedMessagesListDTO `json:"data"`
}

// UnpinMessageResponse สำหรับผลลัพธ์การยกเลิกปักหมุดข้อความ
type UnpinMessageResponse struct {
    GenericResponse
}
```

## 3. Repository Layer

### 3.1 Interface: `domain/repository/pinned_message_repository.go`

```go
package repository

import (
    "context"
    "github.com/google/uuid"
    "github.com/thizplus/gofiber-chat-api/domain/models"
)

type PinnedMessageRepository interface {
    // Create a pinned message
    CreatePinnedMessage(ctx context.Context, pinnedMessage *models.PinnedMessage) error

    // Delete a pinned message
    DeletePinnedMessage(ctx context.Context, messageID, userID uuid.UUID, pinType string) error

    // Get pinned message by ID
    GetPinnedMessageByID(ctx context.Context, id uuid.UUID) (*models.PinnedMessage, error)

    // Check if message is pinned
    IsPinned(ctx context.Context, messageID, userID uuid.UUID, pinType string) (bool, error)

    // Get all pinned messages in a conversation
    GetPinnedMessagesByConversation(ctx context.Context, conversationID, userID uuid.UUID, pinType string, limit, offset int) ([]*models.PinnedMessage, int, error)

    // Get public pinned messages count
    GetPublicPinnedCount(ctx context.Context, conversationID uuid.UUID) (int, error)

    // Delete all pinned messages for a specific message
    DeleteAllPinnedByMessage(ctx context.Context, messageID uuid.UUID) error
}
```

### 3.2 Implementation: `infrastructure/persistence/postgres/pinned_message_repository.go`

```go
package postgres

import (
    "context"
    "github.com/google/uuid"
    "github.com/thizplus/gofiber-chat-api/domain/models"
    "github.com/thizplus/gofiber-chat-api/domain/repository"
    "gorm.io/gorm"
)

type pinnedMessageRepository struct {
    db *gorm.DB
}

func NewPinnedMessageRepository(db *gorm.DB) repository.PinnedMessageRepository {
    return &pinnedMessageRepository{db: db}
}

// Implementation methods here...
```

## 4. Service Layer

### 4.1 Interface: `domain/service/pinned_message_service.go`

```go
package service

import (
    "context"
    "github.com/google/uuid"
    "github.com/thizplus/gofiber-chat-api/domain/dto"
)

type PinnedMessageService interface {
    // Pin a message
    PinMessage(ctx context.Context, conversationID, messageID, userID uuid.UUID, pinType string) (*dto.PinnedMessageDTO, error)

    // Unpin a message
    UnpinMessage(ctx context.Context, conversationID, messageID, userID uuid.UUID, pinType string) error

    // Get pinned messages in a conversation
    GetPinnedMessages(ctx context.Context, conversationID, userID uuid.UUID, pinType string, limit, offset int) (*dto.PinnedMessagesListDTO, error)

    // Toggle pin (pin if not pinned, unpin if pinned)
    TogglePin(ctx context.Context, conversationID, messageID, userID uuid.UUID, pinType string) (*dto.PinnedMessageDTO, error)
}
```

### 4.2 Implementation: `application/serviceimpl/pinned_message_service.go`

```go
package serviceimpl

import (
    "context"
    "errors"
    "github.com/google/uuid"
    "github.com/thizplus/gofiber-chat-api/domain/dto"
    "github.com/thizplus/gofiber-chat-api/domain/repository"
    "github.com/thizplus/gofiber-chat-api/domain/service"
)

type pinnedMessageService struct {
    pinnedRepo       repository.PinnedMessageRepository
    messageRepo      repository.MessageRepository
    conversationRepo repository.ConversationsRepository
}

func NewPinnedMessageService(
    pinnedRepo repository.PinnedMessageRepository,
    messageRepo repository.MessageRepository,
    conversationRepo repository.ConversationsRepository,
) service.PinnedMessageService {
    return &pinnedMessageService{
        pinnedRepo:       pinnedRepo,
        messageRepo:      messageRepo,
        conversationRepo: conversationRepo,
    }
}

// Implementation methods here...
```

## 5. Handler/API Layer

### 5.1 Handler: `interfaces/api/handler/pinned_message_handler.go`

```go
package handler

import (
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "github.com/thizplus/gofiber-chat-api/domain/dto"
    "github.com/thizplus/gofiber-chat-api/domain/service"
    "github.com/thizplus/gofiber-chat-api/pkg/middleware"
)

type PinnedMessageHandler struct {
    service service.PinnedMessageService
}

func NewPinnedMessageHandler(service service.PinnedMessageService) *PinnedMessageHandler {
    return &PinnedMessageHandler{service: service}
}

// PinMessage godoc
// @Summary Pin a message
// @Description Pin a message in a conversation (personal or public)
// @Tags Pinned Messages
// @Accept json
// @Produce json
// @Param conversation_id path string true "Conversation ID"
// @Param message_id path string true "Message ID"
// @Param request body dto.PinMessageRequest true "Pin request"
// @Success 200 {object} dto.PinMessageResponse
// @Router /api/v1/conversations/{conversation_id}/messages/{message_id}/pin [post]
func (h *PinnedMessageHandler) PinMessage(c *fiber.Ctx) error {
    // Implementation...
}

// UnpinMessage godoc
// @Summary Unpin a message
// @Description Unpin a message in a conversation
// @Tags Pinned Messages
// @Accept json
// @Produce json
// @Param conversation_id path string true "Conversation ID"
// @Param message_id path string true "Message ID"
// @Param pin_type query string true "Pin type (personal/public)"
// @Success 200 {object} dto.UnpinMessageResponse
// @Router /api/v1/conversations/{conversation_id}/messages/{message_id}/pin [delete]
func (h *PinnedMessageHandler) UnpinMessage(c *fiber.Ctx) error {
    // Implementation...
}

// GetPinnedMessages godoc
// @Summary Get pinned messages
// @Description Get all pinned messages in a conversation
// @Tags Pinned Messages
// @Accept json
// @Produce json
// @Param conversation_id path string true "Conversation ID"
// @Param pin_type query string false "Pin type (personal/public/all)" default(all)
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} dto.PinnedMessagesResponse
// @Router /api/v1/conversations/{conversation_id}/pinned-messages [get]
func (h *PinnedMessageHandler) GetPinnedMessages(c *fiber.Ctx) error {
    // Implementation...
}
```

### 5.2 Routes: `interfaces/api/routes/pinned_message_routes.go`

```go
package routes

import (
    "github.com/gofiber/fiber/v2"
    "github.com/thizplus/gofiber-chat-api/interfaces/api/handler"
    "github.com/thizplus/gofiber-chat-api/pkg/middleware"
)

func SetupPinnedMessageRoutes(app *fiber.App, handler *handler.PinnedMessageHandler, auth *middleware.AuthMiddleware) {
    api := app.Group("/api/v1")

    // Pinned message routes
    conversations := api.Group("/conversations/:conversation_id")
    conversations.Use(auth.Protected())

    // Pin/Unpin message
    conversations.Post("/messages/:message_id/pin", handler.PinMessage)
    conversations.Delete("/messages/:message_id/pin", handler.UnpinMessage)

    // Get pinned messages
    conversations.Get("/pinned-messages", handler.GetPinnedMessages)
}
```

## 6. WebSocket Events

### 6.1 WebSocket Broadcast Events: `interfaces/websocket/broadcast.go`

เพิ่ม event types ใหม่:

```go
const (
    // Existing events...
    EventMessagePinned   = "message.pinned"
    EventMessageUnpinned = "message.unpinned"
)

// MessagePinnedEvent
type MessagePinnedEvent struct {
    ConversationID uuid.UUID                `json:"conversation_id"`
    MessageID      uuid.UUID                `json:"message_id"`
    PinType        string                   `json:"pin_type"`
    PinnedBy       uuid.UUID                `json:"pinned_by"`
    PinnedMessage  *dto.PinnedMessageDTO    `json:"pinned_message"`
    Timestamp      time.Time                `json:"timestamp"`
}

// MessageUnpinnedEvent
type MessageUnpinnedEvent struct {
    ConversationID uuid.UUID `json:"conversation_id"`
    MessageID      uuid.UUID `json:"message_id"`
    PinType        string    `json:"pin_type"`
    UnpinnedBy     uuid.UUID `json:"unpinned_by"`
    Timestamp      time.Time `json:"timestamp"`
}
```

### 6.2 Broadcasting Logic

```go
// Broadcast when message is pinned
func (h *Hub) BroadcastMessagePinned(conversationID uuid.UUID, pinnedMsg *dto.PinnedMessageDTO, pinType string) {
    event := MessagePinnedEvent{
        ConversationID: conversationID,
        MessageID:      pinnedMsg.MessageID,
        PinType:        pinType,
        PinnedBy:       pinnedMsg.UserID,
        PinnedMessage:  pinnedMsg,
        Timestamp:      time.Now(),
    }

    // For public pins, broadcast to all members
    if pinType == "public" {
        h.BroadcastToConversation(conversationID, EventMessagePinned, event)
    } else {
        // For personal pins, only send to the user who pinned
        h.SendToUser(pinnedMsg.UserID, EventMessagePinned, event)
    }
}
```

## 7. Business Rules & Validations

### 7.1 Permission Rules

**Personal Pin:**
- ✅ ทุกคนสามารถ pin ข้อความสำหรับตัวเองได้
- ✅ มองเห็นได้เฉพาะผู้ที่ pin
- ✅ ไม่จำกัดจำนวน

**Public Pin:**
- ✅ เฉพาะ admin/owner ของ conversation เท่านั้น
- ✅ มองเห็นได้ทุกคนใน conversation
- ⚠️ จำกัดจำนวนไม่เกิน 3-5 ข้อความต่อ conversation (ตั้งค่าได้)
- ⚠️ หาก pin เกินจำนวนสูงสุด ให้ unpin ข้อความเก่าสุดอัตโนมัติ

### 7.2 Validation Checks

```go
// 1. Check if user is member of conversation
// 2. Check if message exists and belongs to conversation
// 3. Check if message is not deleted
// 4. For public pin: Check if user is admin/owner
// 5. For public pin: Check max pinned messages limit
// 6. Check if already pinned (prevent duplicates)
```

### 7.3 Configuration

```go
const (
    MaxPublicPinnedMessages  = 5
    MaxPersonalPinnedMessages = 100 // หรือไม่จำกัด
)
```

## 8. Database Migration

### 8.1 Migration File: `infrastructure/persistence/database/migrations/YYYYMMDD_create_pinned_messages_table.sql`

```sql
-- Create pinned_messages table
CREATE TABLE IF NOT EXISTS pinned_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    message_id UUID NOT NULL,
    conversation_id UUID NOT NULL,
    user_id UUID NOT NULL,
    pin_type VARCHAR(20) NOT NULL CHECK (pin_type IN ('personal', 'public')),
    pinned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Foreign Keys
    CONSTRAINT fk_pinned_messages_message
        FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
    CONSTRAINT fk_pinned_messages_conversation
        FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
    CONSTRAINT fk_pinned_messages_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,

    -- Unique Constraint
    CONSTRAINT unique_pinned_message_user_type
        UNIQUE (message_id, user_id, pin_type)
);

-- Create Indexes
CREATE INDEX idx_pinned_messages_conversation_id ON pinned_messages(conversation_id);
CREATE INDEX idx_pinned_messages_user_id ON pinned_messages(user_id);
CREATE INDEX idx_pinned_messages_pin_type ON pinned_messages(pin_type);
CREATE INDEX idx_pinned_messages_message_id ON pinned_messages(message_id);
CREATE INDEX idx_pinned_messages_pinned_at ON pinned_messages(pinned_at DESC);

-- Add comment
COMMENT ON TABLE pinned_messages IS 'Stores pinned messages for conversations (personal and public pins)';
COMMENT ON COLUMN pinned_messages.pin_type IS 'Type of pin: personal (visible only to user) or public (visible to all members)';
```

### 8.2 Update Migration Runner: `infrastructure/persistence/database/migration.go`

```go
func RunMigrations(db *gorm.DB) error {
    // Existing migrations...

    // Add new model
    err = db.AutoMigrate(
        // Existing models...
        &models.PinnedMessage{},
    )

    return err
}
```

## 9. Dependency Injection

### 9.1 Update DI Container: `pkg/di/container.go`

```go
// Add repository
pinnedMessageRepo := postgres.NewPinnedMessageRepository(db)

// Add service
pinnedMessageService := serviceimpl.NewPinnedMessageService(
    pinnedMessageRepo,
    messageRepo,
    conversationRepo,
)

// Add handler
pinnedMessageHandler := handler.NewPinnedMessageHandler(pinnedMessageService)

// Setup routes
routes.SetupPinnedMessageRoutes(app, pinnedMessageHandler, authMiddleware)
```

## 10. API Endpoints Summary

| Method | Endpoint | Description | Permission |
|--------|----------|-------------|------------|
| POST | `/api/v1/conversations/:id/messages/:messageId/pin` | Pin a message | Member (personal), Admin/Owner (public) |
| DELETE | `/api/v1/conversations/:id/messages/:messageId/pin` | Unpin a message | Pin owner (personal), Admin/Owner (public) |
| GET | `/api/v1/conversations/:id/pinned-messages` | Get pinned messages | Member |

## 11. Example API Usage

### 11.1 Pin Message (Personal)

```bash
POST /api/v1/conversations/123e4567-e89b-12d3-a456-426614174000/messages/456e7890-e89b-12d3-a456-426614174001/pin
Content-Type: application/json
Authorization: Bearer {token}

{
  "pin_type": "personal"
}

Response:
{
  "success": true,
  "message": "Message pinned successfully",
  "data": {
    "id": "789e0123-e89b-12d3-a456-426614174002",
    "message_id": "456e7890-e89b-12d3-a456-426614174001",
    "conversation_id": "123e4567-e89b-12d3-a456-426614174000",
    "user_id": "user-uuid",
    "pin_type": "personal",
    "pinned_at": "2025-01-14T10:30:00Z"
  }
}
```

### 11.2 Pin Message (Public) - Admin Only

```bash
POST /api/v1/conversations/123e4567-e89b-12d3-a456-426614174000/messages/456e7890-e89b-12d3-a456-426614174001/pin
Content-Type: application/json
Authorization: Bearer {admin-token}

{
  "pin_type": "public"
}
```

### 11.3 Get Pinned Messages

```bash
GET /api/v1/conversations/123e4567-e89b-12d3-a456-426614174000/pinned-messages?pin_type=all&limit=50&offset=0
Authorization: Bearer {token}

Response:
{
  "success": true,
  "message": "Pinned messages retrieved successfully",
  "data": {
    "conversation_id": "123e4567-e89b-12d3-a456-426614174000",
    "total": 7,
    "pinned_messages": [
      {
        "id": "pin-uuid-1",
        "message_id": "msg-uuid-1",
        "pin_type": "public",
        "pinned_at": "2025-01-14T10:30:00Z",
        "pinned_by": {...},
        "message": {...}
      },
      {
        "id": "pin-uuid-2",
        "message_id": "msg-uuid-2",
        "pin_type": "personal",
        "pinned_at": "2025-01-14T09:15:00Z",
        "message": {...}
      }
    ]
  }
}
```

### 11.4 Unpin Message

```bash
DELETE /api/v1/conversations/123e4567-e89b-12d3-a456-426614174000/messages/456e7890-e89b-12d3-a456-426614174001/pin?pin_type=personal
Authorization: Bearer {token}

Response:
{
  "success": true,
  "message": "Message unpinned successfully"
}
```

## 12. Implementation Steps (Recommended Order)

1. ✅ **Phase 1: Database & Models**
   - Create migration script
   - Create PinnedMessage model
   - Run migration
   - Test database schema

2. ✅ **Phase 2: DTOs**
   - Create all request/response DTOs
   - Add validation rules

3. ✅ **Phase 3: Repository Layer**
   - Create repository interface
   - Implement repository methods
   - Write unit tests for repository

4. ✅ **Phase 4: Service Layer**
   - Create service interface
   - Implement service methods
   - Add business logic and validations
   - Write unit tests for service

5. ✅ **Phase 5: Handler & Routes**
   - Create handler
   - Implement API endpoints
   - Setup routes
   - Add to DI container

6. ✅ **Phase 6: WebSocket Integration**
   - Add WebSocket events
   - Implement broadcasting logic
   - Test real-time updates

7. ✅ **Phase 7: Testing**
   - Integration tests
   - API endpoint tests
   - WebSocket event tests
   - Permission tests

8. ✅ **Phase 8: Documentation**
   - Update API documentation
   - Add Swagger annotations
   - Update README

## 13. Testing Scenarios

### 13.1 Unit Tests
- ✅ Pin personal message
- ✅ Pin public message (as admin)
- ✅ Pin public message (as non-admin) - Should fail
- ✅ Unpin message
- ✅ Get pinned messages (personal only)
- ✅ Get pinned messages (public only)
- ✅ Get pinned messages (all)
- ✅ Pin already pinned message - Should fail
- ✅ Pin deleted message - Should fail
- ✅ Pin message in conversation where user is not a member - Should fail
- ✅ Exceed max public pins limit

### 13.2 Integration Tests
- ✅ Full pin/unpin flow
- ✅ WebSocket event delivery
- ✅ Concurrent pin operations
- ✅ Permission checks across different user roles

## 14. Future Enhancements (Optional)

1. **Pin Notifications**
   - แจ้งเตือนเมื่อมีการ pin public message

2. **Pin Order Management**
   - สามารถจัดลำดับ pinned messages ได้

3. **Pin with Comments**
   - เพิ่ม note หรือ comment เมื่อ pin message

4. **Pin Analytics**
   - ติดตาม metrics ว่าข้อความไหนถูก pin บ่อยที่สุด

5. **Bulk Pin Operations**
   - Pin/Unpin หลายข้อความพร้อมกัน

6. **Pin Expiration**
   - กำหนดเวลาหมดอายุของ pinned message

## 15. Security Considerations

1. ✅ **Authentication**: ตรวจสอบ JWT token
2. ✅ **Authorization**: ตรวจสอบสิทธิ์ตาม pin type
3. ✅ **Input Validation**: Validate ทุก input parameters
4. ✅ **Rate Limiting**: จำกัดจำนวนครั้งในการ pin/unpin
5. ✅ **SQL Injection Prevention**: ใช้ prepared statements (GORM handles this)
6. ✅ **Data Privacy**: Personal pins ต้องมองเห็นได้เฉพาะเจ้าของ

## 16. Performance Optimization

1. **Database Indexes**: สร้าง indexes ที่เหมาะสม
2. **Caching**: Cache pinned messages ที่เป็น public
3. **Pagination**: ใช้ pagination เมื่อดึง pinned messages
4. **Lazy Loading**: โหลด message details เมื่อจำเป็นเท่านั้น
5. **Query Optimization**: ใช้ joins ที่มีประสิทธิภาพ

---

## Summary

ฟีเจอร์ Pin Message นี้จะช่วยให้ผู้ใช้สามารถ:
- ✅ ปักหมุดข้อความสำคัญสำหรับตัวเอง (Personal Pin)
- ✅ ปักหมุดข้อความสำคัญให้ทุกคนใน conversation เห็น (Public Pin - สำหรับ admin)
- ✅ จัดการ pinned messages ได้ง่าย
- ✅ รับ real-time updates ผ่าน WebSocket

ระบบนี้ออกแบบมาเพื่อรองรับการใช้งานที่ยืดหยุ่น มีความปลอดภัย และ scale ได้ดี
