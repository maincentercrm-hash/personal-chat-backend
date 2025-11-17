# Delete/Leave Conversation Feature - Implementation Plan

## สรุป Requirement

### 1. Direct Conversation (แชทส่วนตัว 1:1)
**ปัญหาปัจจุบัน:**
- ❌ ผู้ใช้ไม่สามารถ "ลบแชท" หรือ "ออกจาก conversation" ได้
- ❌ แชทที่ไม่ต้องการจะติดค้างอยู่ในรายการตลอด

**แนวทางแก้ไข: Soft Delete (Hide Conversation)**
- ✅ ผู้ใช้สามารถ "ลบแชท" (ซ่อน conversation) จากรายการได้
- ✅ อีกฝ่ายยังเห็น conversation อยู่ตามปกติ (ไม่กระทบ)
- ✅ ถ้าได้รับข้อความใหม่ conversation จะกลับมาแสดงอัตโนมัติ (Unhide)
- ✅ เหมือน WhatsApp, Telegram, Line

**ตัวอย่าง Use Case:**
```
User A และ User B มีแชทส่วนตัว
→ User A "ลบแชท" (Hide)
→ แชทหายจากรายการของ User A
→ User B ยังเห็นแชทอยู่ตามปกติ
→ User B ส่งข้อความใหม่
→ แชทกลับมาแสดงในรายการของ User A อีกครั้ง
```

### 2. Group Conversation (แชทกลุ่ม)
**สถานะปัจจุบัน:**
- ✅ สามารถออกจากกลุ่ม (Leave Group) ได้แล้ว
- ✅ Admin สามารถลบสมาชิกคนอื่นได้
- ✅ ป้องกันการลบ admin คนสุดท้าย

**ไม่ต้องแก้ไข - ทำงานได้ดีแล้ว**

### 3. Summary Table

| Conversation Type | Action | ปัจจุบัน | หลังแก้ไข |
|------------------|--------|----------|-----------|
| **Direct** | ลบแชท (Hide) | ❌ ไม่ได้ | ✅ ได้ (Soft Delete) |
| **Direct** | Leave | ❌ ไม่ได้ | ✅ ได้ (เหมือน Hide) |
| **Group** | Leave | ✅ ได้ | ✅ ได้ (เหมือนเดิม) |
| **Group** | Hide | ❌ ไม่ได้ | ✅ ได้ (Optional - เพิ่มใหม่) |

---

## Implementation Plan

### Phase 1: Database Schema Changes

#### 1.1 เพิ่ม Field ใน `conversation_members` table

```sql
-- Migration: Add hidden fields to conversation_members table
ALTER TABLE conversation_members
ADD COLUMN is_hidden BOOLEAN DEFAULT FALSE,
ADD COLUMN hidden_at TIMESTAMP WITH TIME ZONE NULL;

-- Add index for performance
CREATE INDEX idx_conversation_members_is_hidden ON conversation_members(is_hidden);

-- Add comment
COMMENT ON COLUMN conversation_members.is_hidden IS 'User has hidden this conversation from their list';
COMMENT ON COLUMN conversation_members.hidden_at IS 'Timestamp when conversation was hidden';
```

#### 1.2 Update Model: `domain/models/conversation_member.go`

```go
type ConversationMember struct {
    ID                   uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
    ConversationID       uuid.UUID   `json:"conversation_id" gorm:"type:uuid;not null"`
    UserID               uuid.UUID   `json:"user_id" gorm:"type:uuid;not null"`
    Role                 string      `json:"role" gorm:"type:varchar(20);default:'member'"`
    IsAdmin              bool        `json:"is_admin" gorm:"default:false"`
    JoinedAt             time.Time   `json:"joined_at" gorm:"type:timestamp with time zone;default:now()"`
    LastReadAt           *time.Time  `json:"last_read_at,omitempty" gorm:"type:timestamp with time zone"`
    IsMuted              bool        `json:"is_muted" gorm:"default:false"`
    IsPinned             bool        `json:"is_pinned" gorm:"default:false"`
    IsHidden             bool        `json:"is_hidden" gorm:"default:false"`          // NEW
    HiddenAt             *time.Time  `json:"hidden_at,omitempty" gorm:"type:timestamp with time zone"` // NEW
    Nickname             string      `json:"nickname,omitempty" gorm:"type:varchar(100)"`
    NotificationSettings types.JSONB `json:"notification_settings,omitempty" gorm:"type:jsonb;default:'{}'::jsonb"`

    // Associations
    Conversation *Conversation `json:"conversation,omitempty" gorm:"foreignkey:ConversationID"`
    User         *User         `json:"user,omitempty" gorm:"foreignkey:UserID"`
}
```

---

### Phase 2: DTOs

#### 2.1 Request DTOs: `domain/dto/conversation_dto.go`

```go
// HideConversationRequest สำหรับการซ่อน/แสดง conversation
type HideConversationRequest struct {
    IsHidden bool `json:"is_hidden" validate:"required"`
}

// DeleteConversationRequest สำหรับการลบ conversation (alias for hide)
// สำหรับ Direct: Hide conversation
// สำหรับ Group: Leave conversation
type DeleteConversationRequest struct {
    // ไม่ต้องมี field - ใช้ HTTP DELETE method
}
```

#### 2.2 Response DTOs

```go
// HideConversationResponse
type HideConversationResponse struct {
    GenericResponse
    Data struct {
        IsHidden bool       `json:"is_hidden"`
        HiddenAt *time.Time `json:"hidden_at,omitempty"`
    } `json:"data"`
}

// DeleteConversationResponse
type DeleteConversationResponse struct {
    GenericResponse
    Data struct {
        ConversationID string `json:"conversation_id"`
        Action         string `json:"action"` // "hidden" or "left"
        Message        string `json:"message"`
    } `json:"data"`
}
```

#### 2.3 Update ConversationDTO

```go
type ConversationDTO struct {
    // ... existing fields ...
    IsHidden bool       `json:"is_hidden"` // NEW
    HiddenAt *time.Time `json:"hidden_at,omitempty"` // NEW
}
```

---

### Phase 3: Repository Layer

#### 3.1 Update Interface: `domain/repository/conversations_repository.go`

```go
type ConversationsRepository interface {
    // ... existing methods ...

    // Hide/Unhide conversation
    SetHiddenStatus(conversationID, userID uuid.UUID, isHidden bool) error

    // Check if conversation is hidden
    IsHidden(conversationID, userID uuid.UUID) (bool, error)
}
```

#### 3.2 Implementation: `infrastructure/persistence/postgres/conversation_repository.go`

```go
// SetHiddenStatus ตั้งค่าสถานะการซ่อน conversation
func (r *conversationRepository) SetHiddenStatus(conversationID, userID uuid.UUID, isHidden bool) error {
    updates := map[string]interface{}{
        "is_hidden": isHidden,
    }

    if isHidden {
        now := time.Now()
        updates["hidden_at"] = now
    } else {
        updates["hidden_at"] = nil
    }

    result := r.db.Model(&models.ConversationMember{}).
        Where("conversation_id = ? AND user_id = ?", conversationID, userID).
        Updates(updates)

    if result.Error != nil {
        return result.Error
    }

    if result.RowsAffected == 0 {
        return errors.New("conversation member not found")
    }

    return nil
}

// IsHidden ตรวจสอบว่า conversation ถูกซ่อนหรือไม่
func (r *conversationRepository) IsHidden(conversationID, userID uuid.UUID) (bool, error) {
    var member models.ConversationMember

    err := r.db.Where("conversation_id = ? AND user_id = ?", conversationID, userID).
        First(&member).Error

    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return false, errors.New("conversation member not found")
        }
        return false, err
    }

    return member.IsHidden, nil
}
```

#### 3.3 Update GetUserConversations Query

ต้องแก้ไข query เพื่อ **ไม่แสดง conversations ที่ hidden**

```go
// ใน conversation_repository.go - method GetUserConversations
func (r *conversationRepository) GetUserConversations(...) {
    query := r.db.
        Table("conversations").
        Select("conversations.*, cm.last_read_at, cm.is_muted, cm.is_pinned, cm.is_hidden").
        Joins("INNER JOIN conversation_members cm ON conversations.id = cm.conversation_id").
        Where("cm.user_id = ?", userID).
        Where("cm.is_hidden = ?", false) // NEW: ไม่แสดง conversations ที่ซ่อน

    // ... rest of the query
}
```

---

### Phase 4: Service Layer

#### 4.1 Update Interface: `domain/service/conversations_service.go`

```go
type ConversationService interface {
    // ... existing methods ...

    // Hide/Unhide conversation (for Direct conversations)
    SetHiddenStatus(conversationID, userID uuid.UUID, isHidden bool) error

    // Delete conversation (smart delete - hide for direct, leave for group)
    DeleteConversation(conversationID, userID uuid.UUID) (string, error)
}
```

#### 4.2 Implementation: `application/serviceimpl/conversations_service.go`

```go
// SetHiddenStatus ตั้งค่าสถานะการซ่อน conversation
func (s *conversationService) SetHiddenStatus(conversationID, userID uuid.UUID, isHidden bool) error {
    // 1. ตรวจสอบว่าเป็นสมาชิก
    isMember, err := s.conversationRepo.IsMember(conversationID, userID)
    if err != nil {
        return err
    }
    if !isMember {
        return errors.New("you are not a member of this conversation")
    }

    // 2. ตั้งค่า hidden status
    return s.conversationRepo.SetHiddenStatus(conversationID, userID, isHidden)
}

// DeleteConversation ลบ conversation (smart delete)
// - Direct conversation: Hide
// - Group conversation: Leave (Remove member)
func (s *conversationService) DeleteConversation(conversationID, userID uuid.UUID) (string, error) {
    // 1. ตรวจสอบว่าเป็นสมาชิก
    isMember, err := s.conversationRepo.IsMember(conversationID, userID)
    if err != nil {
        return "", err
    }
    if !isMember {
        return "", errors.New("you are not a member of this conversation")
    }

    // 2. ดึงข้อมูล conversation
    conversation, err := s.conversationRepo.GetByID(conversationID)
    if err != nil {
        return "", err
    }

    // 3. จัดการตามประเภท
    if conversation.Type == "direct" {
        // Direct: Hide conversation
        err = s.conversationRepo.SetHiddenStatus(conversationID, userID, true)
        if err != nil {
            return "", err
        }
        return "hidden", nil
    } else {
        // Group: Remove member (leave group)
        err = s.memberService.RemoveMember(userID, conversationID, userID)
        if err != nil {
            return "", err
        }
        return "left", nil
    }
}
```

---

### Phase 5: Message Service - Auto Unhide

ต้องเพิ่ม logic เพื่อ **unhide conversation อัตโนมัติ** เมื่อมีข้อความใหม่

#### 5.1 Update: `application/serviceimpl/message_send_standard_service.go`

```go
func (s *messageSendStandardService) SendTextMessage(...) {
    // ... existing code ...

    // NEW: Auto unhide conversation for all members when new message arrives
    members, err := s.conversationRepo.GetMembers(conversationID)
    if err == nil {
        for _, member := range members {
            if member.IsHidden {
                // Unhide conversation for this member
                s.conversationRepo.SetHiddenStatus(conversationID, member.UserID, false)
            }
        }
    }

    // ... rest of the code ...
}
```

**หมายเหตุ:** ต้องเพิ่ม logic นี้ในทุก methods ที่ส่งข้อความ:
- `SendTextMessage`
- `SendImageMessage`
- `SendFileMessage`
- `SendStickerMessage`
- `ReplyToMessage`

---

### Phase 6: Handler & Routes

#### 6.1 Update Handler: `interfaces/api/handler/conversation_handler.go`

```go
// HideConversation ซ่อน/แสดง conversation
func (h *ConversationHandler) HideConversation(c *fiber.Ctx) error {
    userID, err := middleware.GetUserUUID(c)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "success": false,
            "message": "Unauthorized: " + err.Error(),
        })
    }

    conversationID, err := utils.ParseUUIDParam(c, "conversationId")
    if err != nil {
        return err
    }

    var input dto.HideConversationRequest
    if err := c.BodyParser(&input); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "Invalid request data: " + err.Error(),
        })
    }

    err = h.conversationService.SetHiddenStatus(conversationID, userID, input.IsHidden)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "success": false,
            "message": err.Error(),
        })
    }

    var hiddenAt *time.Time
    if input.IsHidden {
        now := time.Now()
        hiddenAt = &now
    }

    return c.JSON(fiber.Map{
        "success": true,
        "message": "Conversation hidden status updated successfully",
        "data": fiber.Map{
            "is_hidden": input.IsHidden,
            "hidden_at": hiddenAt,
        },
    })
}

// DeleteConversation ลบ conversation (smart delete)
func (h *ConversationHandler) DeleteConversation(c *fiber.Ctx) error {
    userID, err := middleware.GetUserUUID(c)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "success": false,
            "message": "Unauthorized: " + err.Error(),
        })
    }

    conversationID, err := utils.ParseUUIDParam(c, "conversationId")
    if err != nil {
        return err
    }

    action, err := h.conversationService.DeleteConversation(conversationID, userID)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "success": false,
            "message": err.Error(),
        })
    }

    var message string
    if action == "hidden" {
        message = "Conversation hidden successfully"
    } else {
        message = "Left conversation successfully"
    }

    return c.JSON(fiber.Map{
        "success": true,
        "message": message,
        "data": fiber.Map{
            "conversation_id": conversationID.String(),
            "action":          action,
            "message":         message,
        },
    })
}
```

#### 6.2 Update Routes: `interfaces/api/routes/conversation_routes.go`

```go
func SetupConversationRoutes(
    router fiber.Router,
    conversationHandler *handler.ConversationHandler,
    conversationMemberHandler *handler.ConversationMemberHandler,
) {
    conversations := router.Group("/conversations")
    conversations.Use(middleware.Protected())

    // ... existing routes ...

    // NEW: Hide/Unhide conversation
    conversations.Patch("/:conversationId/hide", conversationHandler.HideConversation)

    // NEW: Delete conversation (smart delete - hide for direct, leave for group)
    conversations.Delete("/:conversationId", conversationHandler.DeleteConversation)
}
```

---

### Phase 7: WebSocket Integration

#### 7.1 Add Events: `interfaces/websocket/broadcast.go`

```go
const (
    // ... existing events ...
    EventConversationHidden   = "conversation.hidden"
    EventConversationUnhidden = "conversation.unhidden"
)

// ConversationHiddenEvent
type ConversationHiddenEvent struct {
    ConversationID uuid.UUID `json:"conversation_id"`
    UserID         uuid.UUID `json:"user_id"`
    IsHidden       bool      `json:"is_hidden"`
    Timestamp      time.Time `json:"timestamp"`
}
```

#### 7.2 Broadcast Logic

```go
// ใน notification_service.go
func (s *notificationService) NotifyConversationHidden(userID, conversationID uuid.UUID, isHidden bool) {
    event := ConversationHiddenEvent{
        ConversationID: conversationID,
        UserID:         userID,
        IsHidden:       isHidden,
        Timestamp:      time.Now(),
    }

    // ส่งเฉพาะผู้ใช้คนนั้นๆ (ไม่ส่งให้สมาชิคนอื่น)
    s.wsAdapter.SendToUser(userID, EventConversationHidden, event)
}
```

---

## API Specification for Frontend

### 1. Hide/Unhide Conversation

**Endpoint:** `PATCH /api/v1/conversations/:conversationId/hide`

**Description:** ซ่อนหรือแสดง conversation (ใช้กับทั้ง Direct และ Group)

**Request:**
```json
{
  "is_hidden": true
}
```

**Response (Success - 200):**
```json
{
  "success": true,
  "message": "Conversation hidden status updated successfully",
  "data": {
    "is_hidden": true,
    "hidden_at": "2025-01-14T10:30:00Z"
  }
}
```

**Response (Error - 403):**
```json
{
  "success": false,
  "message": "you are not a member of this conversation"
}
```

---

### 2. Delete Conversation (Smart Delete)

**Endpoint:** `DELETE /api/v1/conversations/:conversationId`

**Description:** ลบ conversation
- **Direct conversation:** ซ่อน (hide) conversation
- **Group conversation:** ออกจากกลุ่ม (leave)

**Request:** ไม่ต้องมี body

**Response (Success - Direct - 200):**
```json
{
  "success": true,
  "message": "Conversation hidden successfully",
  "data": {
    "conversation_id": "123e4567-e89b-12d3-a456-426614174000",
    "action": "hidden",
    "message": "Conversation hidden successfully"
  }
}
```

**Response (Success - Group - 200):**
```json
{
  "success": true,
  "message": "Left conversation successfully",
  "data": {
    "conversation_id": "123e4567-e89b-12d3-a456-426614174000",
    "action": "left",
    "message": "Left conversation successfully"
  }
}
```

**Response (Error - 403):**
```json
{
  "success": false,
  "message": "you are not a member of this conversation"
}
```

---

### 3. Leave Group (Existing - No Changes)

**Endpoint:** `DELETE /api/v1/conversations/:conversationId/members/:userId`

**Description:** ลบสมาชิกออกจากกลุ่ม (ใช้ userId ของตัวเองเพื่อออกจากกลุ่ม)

**Response (Success - 200):**
```json
{
  "success": true,
  "message": "Member removed successfully"
}
```

**Response (Error - Group - 400):**
```json
{
  "success": false,
  "message": "cannot remove members from direct conversation"
}
```

---

### 4. Get User Conversations (Updated)

**Endpoint:** `GET /api/v1/conversations`

**Changes:**
- ✅ จะไม่แสดง conversations ที่ `is_hidden = true` อีกต่อไป
- ✅ เพิ่ม field `is_hidden` และ `hidden_at` ใน response

**Query Parameters:**
```
limit: 20
offset: 0
type: direct|group|business (optional)
pinned: true|false (optional)
show_hidden: true|false (optional - default: false)
```

**Response (Success - 200):**
```json
{
  "success": true,
  "message": "Conversations retrieved successfully",
  "data": {
    "conversations": [
      {
        "id": "123e4567-e89b-12d3-a456-426614174000",
        "type": "direct",
        "title": "John Doe",
        "is_hidden": false,
        "hidden_at": null,
        "is_pinned": true,
        "is_muted": false,
        "last_message": "Hello!",
        "last_message_at": "2025-01-14T10:30:00Z",
        "unread_count": 3
      }
    ],
    "has_more": false,
    "pagination": {
      "total": 25,
      "limit": 20,
      "offset": 0
    }
  }
}
```

---

## WebSocket Events for Frontend

### 1. Conversation Hidden Event

**Event:** `conversation.hidden`

**Payload:**
```json
{
  "conversation_id": "123e4567-e89b-12d3-a456-426614174000",
  "user_id": "user-uuid",
  "is_hidden": true,
  "timestamp": "2025-01-14T10:30:00Z"
}
```

**Frontend Action:**
- ลบ conversation ออกจากรายการ (ถ้า `is_hidden = true`)
- เพิ่ม conversation กลับเข้ารายการ (ถ้า `is_hidden = false`)

### 2. New Message Event (Updated)

**Event:** `message.new`

**Behavior Change:**
- เมื่อได้รับข้อความใหม่ใน conversation ที่ hidden
- Conversation จะ unhide อัตโนมัติ
- Frontend ต้องโหลด conversation นั้นกลับมาแสดง

---

## Frontend Implementation Guide

### Use Cases

#### 1. ลบแชทส่วนตัว (Direct Conversation)

```typescript
// Frontend code example
async function deleteDirectChat(conversationId: string) {
  try {
    const response = await fetch(`/api/v1/conversations/${conversationId}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });

    const data = await response.json();

    if (data.success) {
      if (data.data.action === 'hidden') {
        // ลบ conversation ออกจาก UI
        removeConversationFromList(conversationId);
        showToast('แชทถูกลบแล้ว');
      }
    }
  } catch (error) {
    console.error('Error deleting chat:', error);
  }
}
```

#### 2. ออกจากกลุ่ม (Group Conversation)

```typescript
async function leaveGroup(conversationId: string) {
  try {
    const response = await fetch(`/api/v1/conversations/${conversationId}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });

    const data = await response.json();

    if (data.success) {
      if (data.data.action === 'left') {
        // ลบ conversation ออกจาก UI
        removeConversationFromList(conversationId);
        showToast('คุณออกจากกลุ่มแล้ว');
      }
    }
  } catch (error) {
    console.error('Error leaving group:', error);
  }
}
```

#### 3. Auto Unhide เมื่อได้รับข้อความใหม่

```typescript
// WebSocket listener
socket.on('message.new', (message) => {
  const conversationId = message.conversation_id;

  // ตรวจสอบว่า conversation นี้ถูก hide หรือไม่
  const conversation = findConversation(conversationId);

  if (!conversation) {
    // Conversation ไม่อยู่ในรายการ (อาจถูก hide)
    // โหลด conversation กลับมา
    loadConversation(conversationId);
  }

  // แสดงข้อความใหม่
  displayNewMessage(message);
});
```

#### 4. ซ่อน/แสดง Conversation (Optional Feature)

```typescript
async function toggleHideConversation(conversationId: string, isHidden: boolean) {
  try {
    const response = await fetch(`/api/v1/conversations/${conversationId}/hide`, {
      method: 'PATCH',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ is_hidden: isHidden })
    });

    const data = await response.json();

    if (data.success) {
      if (isHidden) {
        removeConversationFromList(conversationId);
      } else {
        loadConversation(conversationId);
      }
    }
  } catch (error) {
    console.error('Error toggling hide:', error);
  }
}
```

---

## Implementation Steps (Recommended Order)

### Backend Development

1. ✅ **Phase 1: Database Migration** (15 min)
   - เพิ่ม columns `is_hidden`, `hidden_at`
   - Run migration
   - Test database schema

2. ✅ **Phase 2: Models & DTOs** (20 min)
   - Update `ConversationMember` model
   - Create request/response DTOs
   - Add validation

3. ✅ **Phase 3: Repository Layer** (30 min)
   - Add `SetHiddenStatus` method
   - Add `IsHidden` method
   - Update `GetUserConversations` query to exclude hidden
   - Write repository tests

4. ✅ **Phase 4: Service Layer** (45 min)
   - Implement `SetHiddenStatus` service
   - Implement `DeleteConversation` service (smart delete)
   - Add auto-unhide logic in message services
   - Write service tests

5. ✅ **Phase 5: Handler & Routes** (30 min)
   - Create `HideConversation` handler
   - Create `DeleteConversation` handler
   - Add routes
   - Test API endpoints

6. ✅ **Phase 6: WebSocket Events** (20 min)
   - Add `conversation.hidden` event
   - Implement broadcasting
   - Test WebSocket events

7. ✅ **Phase 7: Testing** (1 hour)
   - Unit tests
   - Integration tests
   - API endpoint tests
   - WebSocket event tests

8. ✅ **Phase 8: Documentation** (15 min)
   - Update API documentation
   - Add Swagger annotations
   - Document for frontend team

**Total Estimated Time: 3-4 hours**

### Frontend Development (คำแนะนำ)

1. ✅ **Update API Client** (30 min)
   - เพิ่ม `DELETE /conversations/:id` endpoint
   - เพิ่ม `PATCH /conversations/:id/hide` endpoint
   - Update conversation model to include `is_hidden`, `hidden_at`

2. ✅ **Update UI** (1 hour)
   - เพิ่มปุ่ม "ลบแชท" สำหรับ Direct conversation
   - เพิ่มปุ่ม "ออกจากกลุ่ม" สำหรับ Group conversation
   - แสดง confirmation dialog ก่อนลบ/ออก

3. ✅ **WebSocket Handler** (30 min)
   - รับ `conversation.hidden` event
   - Auto unhide เมื่อได้รับข้อความใหม่
   - Update conversation list

4. ✅ **Testing** (1 hour)
   - Test ลบ Direct conversation
   - Test ออกจาก Group conversation
   - Test auto unhide
   - Test WebSocket events

**Total Estimated Time: 3 hours**

---

## Testing Scenarios

### Backend Tests

1. ✅ **Hide Direct Conversation**
   - User A hides conversation with User B
   - Conversation disappears from User A's list
   - Conversation still visible for User B
   - User B sends new message
   - Conversation reappears for User A

2. ✅ **Delete Direct Conversation**
   - User calls DELETE endpoint
   - Verify action = "hidden"
   - Verify conversation is hidden
   - Verify other user not affected

3. ✅ **Leave Group Conversation**
   - User calls DELETE endpoint
   - Verify action = "left"
   - Verify user removed from members
   - Verify system message created

4. ✅ **Auto Unhide**
   - Hide conversation
   - Receive new message
   - Verify conversation unhidden

5. ✅ **Error Cases**
   - Hide non-existent conversation → 404
   - Hide conversation user is not member of → 403
   - Leave group as last admin → 400

### Frontend Tests

1. ✅ Direct conversation delete flow
2. ✅ Group conversation leave flow
3. ✅ WebSocket event handling
4. ✅ UI updates correctly
5. ✅ Toast notifications display

---

## Summary for Frontend Team

### 📋 Quick Reference

| Action | Conversation Type | API Endpoint | HTTP Method | Result |
|--------|------------------|--------------|-------------|--------|
| ลบแชท | Direct | `/api/v1/conversations/:id` | DELETE | Hide (ซ่อน) |
| ลบแชท | Group | `/api/v1/conversations/:id` | DELETE | Leave (ออก) |
| ซ่อนแชท | Any | `/api/v1/conversations/:id/hide` | PATCH | Hide/Unhide |
| ออกจากกลุ่ม | Group | `/api/v1/conversations/:id/members/:userId` | DELETE | Leave |

### 🔔 Important Notes

1. **Auto Unhide:** เมื่อ conversation ที่ hidden ได้รับข้อความใหม่ จะ unhide อัตโนมัติ
2. **WebSocket Events:** ฟัง `conversation.hidden` event เพื่ออัพเดท UI
3. **Query Parameter:** ใช้ `show_hidden=true` ใน GET conversations เพื่อดู hidden conversations (optional)
4. **Response Field:** ทุก conversation จะมี `is_hidden` และ `hidden_at` fields

### 🚀 Migration Path

**ปัจจุบัน → ใหม่**

```
Direct: ไม่มีปุ่มลบ → มีปุ่ม "ลบแชท" (Hide)
Group: ปุ่ม "ออกจากกลุ่ม" → ปุ่ม "ลบแชท" (Leave) หรือเก็บชื่อเดิม
```

**Recommendation:**
- ใช้ปุ่มเดียว "ลบแชท" สำหรับทั้ง Direct และ Group
- Backend จะจัดการ logic ให้อัตโนมัติ (hide vs leave)
- แสดง confirmation dialog ต่างกันตามประเภท:
  - Direct: "ลบแชทนี้หรือไม่? (สามารถกลับมาแสดงได้ถ้าได้รับข้อความใหม่)"
  - Group: "ออกจากกลุ่มนี้หรือไม่? (คุณจะไม่สามารถอ่านข้อความได้อีก)"

---

## Security & Privacy

1. ✅ **Data Privacy:**
   - Hidden status เป็น personal setting (ไม่แชร์กับสมาชิกคนอื่น)
   - Hidden conversations ยังคงมีอยู่ในฐานข้อมูล (ไม่ถูกลบ)

2. ✅ **Authorization:**
   - ตรวจสอบ membership ก่อนทุก action
   - เฉพาะเจ้าของเท่านั้นที่ซ่อน/แสดง conversation ของตัวเอง

3. ✅ **Data Retention:**
   - Messages ไม่ถูกลบเมื่อ hide conversation
   - สามารถ unhide และอ่าน history ได้

---

**สรุป:** ฟีเจอร์นี้จะทำให้ผู้ใช้สามารถ "ลบแชท" ได้อย่างสมบูรณ์ โดยไม่ทำลายข้อมูล และยังคง UX ที่ดีเหมือนแชทแอปชั้นนำ
