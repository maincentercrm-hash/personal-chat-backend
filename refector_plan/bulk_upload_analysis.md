# Bulk Upload & Album Message Analysis

**วันที่**: 2025-11-12
**วัตถุประสงค์**: วิเคราะห์และออกแบบระบบส่งรูปภาพหลายรูปพร้อมข้อความ (Telegram-like)

---

## 📋 สรุปความต้องการ

**ฟีเจอร์ที่ต้องการ**:
- ส่งภาพหลายภาพพร้อมกัน (เช่น 4 รูปพร้อมกัน)
- แนบข้อความ/caption กับภาพแต่ละภาพได้
- จัดกลุ่มภาพเป็น Album (แสดงเป็น grid)
- เหมือน Telegram/WhatsApp

**ตัวอย่าง Telegram**:
```
[User sends 4 photos with caption "Holiday trip 🏖️"]

Frontend Display:
┌─────────────────────┐
│  Photo 1  │ Photo 2 │  <- Grid 2x2
│  Photo 3  │ Photo 4 │
│                     │
│ Holiday trip 🏖️    │  <- Caption ด้านล่าง
└─────────────────────┘
```

---

## 🔍 วิเคราะห์สถาปัตยกรรมปัจจุบัน

### ✅ สิ่งที่มีอยู่และใช้ได้

#### 1. Message Model รองรับ Metadata (JSONB)
```go
// domain/models/message.go
type Message struct {
    MediaURL          string      // URL รูปภาพ/ไฟล์เดี่ยว
    MediaThumbnailURL string      // Thumbnail
    Metadata          types.JSONB // ✅ ใช้เก็บข้อมูลเพิ่มเติมได้
    // ...
}
```

**ข้อดี**:
- มี `Metadata` JSONB field ที่ยืดหยุ่น
- สามารถเพิ่ม `album_id` หรือ `group_id` ใน metadata ได้

#### 2. Repository รองรับ Batch Insert
```go
// infrastructure/persistence/postgres/message_repository.go
func (r *messageRepository) Create(message *models.Message) error {
    return r.db.Create(message).Error  // ใช้ Create ได้
}
```

**GORM รองรับ**:
```go
r.db.CreateInBatches(messages, 100)  // ✅ มีอยู่แล้วใน GORM
```

#### 3. NotificationService รองรับการแจ้งเตือน
```go
// application/serviceimpl/notification_service.go
func (s *notificationService) NotifyNewMessage(conversationID uuid.UUID, message interface{}) error
```

**สามารถ**: ส่ง notification หลายครั้งหรือส่งรวมได้

### ❌ สิ่งที่ยังไม่มี

1. ❌ **Bulk Upload API** - ไม่มี endpoint รับหลาย media พร้อมกัน
2. ❌ **Album/Group ID** - ไม่มี field สำหรับ group messages
3. ❌ **Bulk Insert Service** - ไม่มี method สร้างหลาย messages พร้อมกัน
4. ❌ **Album Query** - ไม่มี method ดึง messages ที่เป็น album เดียวกัน

---

## 💡 แนวทางการแก้ไข (3 Options)

### 📌 Option 1: เพิ่ม Album ID ใน Metadata (แนะนำ ⭐)

**ทำไมดี**:
- ไม่ต้องแก้ Database Schema
- ใช้ Metadata JSONB ที่มีอยู่แล้ว
- Flexible - เพิ่มฟีเจอร์อื่นได้ภายหลัง
- ไม่กระทบโครงสร้างเดิม

**วิธีการ**:
```json
// Metadata structure
{
  "album_id": "uuid-xxx-xxx",     // ใช้ group messages เป็น album
  "album_position": 0,             // ตำแหน่งในอัลบั้ม (0, 1, 2, 3)
  "album_total": 4,                // จำนวนรูปทั้งหมดใน album
  "album_caption": "Holiday trip"  // Caption สำหรับอัลบั้ม (ใส่ที่ตำแหน่งแรก)
}
```

**ตัวอย่าง Messages ใน Database**:
```
Message 1: { media_url: "photo1.jpg", metadata: { album_id: "abc-123", album_position: 0, album_total: 4, album_caption: "Holiday" } }
Message 2: { media_url: "photo2.jpg", metadata: { album_id: "abc-123", album_position: 1, album_total: 4 } }
Message 3: { media_url: "photo3.jpg", metadata: { album_id: "abc-123", album_position: 2, album_total: 4 } }
Message 4: { media_url: "photo4.jpg", metadata: { album_id: "abc-123", album_position: 3, album_total: 4 } }
```

**Pros**:
- ✅ ไม่ต้อง migrate database
- ✅ Backward compatible (messages เดิมยังใช้ได้)
- ✅ Flexible (เพิ่ม field อื่นใน metadata ได้เสมอ)
- ✅ Query ง่าย: `WHERE metadata->>'album_id' = 'xxx'`

**Cons**:
- ❌ Query ช้ากว่า indexed column เล็กน้อย (แก้ได้ด้วย GIN index)
- ❌ ต้อง parse JSON ทุกครั้ง

---

### 📌 Option 2: เพิ่ม Column ใหม่ใน Message Table

**วิธีการ**:
```go
type Message struct {
    // ... existing fields ...
    AlbumID       *uuid.UUID `json:"album_id,omitempty" gorm:"type:uuid;index"`
    AlbumPosition *int       `json:"album_position,omitempty"`
    AlbumCaption  string     `json:"album_caption,omitempty"`
}
```

**Pros**:
- ✅ Query เร็วกว่า (indexed column)
- ✅ Type-safe (ไม่ต้อง parse JSON)
- ✅ Easier to query and filter

**Cons**:
- ❌ ต้อง migrate database (ADD COLUMN)
- ❌ Downtime อาจเกิดขึ้น (ถ้า table ใหญ่)
- ❌ ไม่ flexible (เพิ่ม field ใหม่ต้อง migrate อีก)

---

### 📌 Option 3: สร้าง Album Table แยก

**วิธีการ**:
```go
type Album struct {
    ID             uuid.UUID `gorm:"type:uuid;primary_key"`
    ConversationID uuid.UUID `gorm:"type:uuid;not null"`
    Caption        string
    CreatedAt      time.Time
    Messages       []*Message `gorm:"foreignkey:AlbumID"`
}

type Message struct {
    // ... existing fields ...
    AlbumID *uuid.UUID `gorm:"type:uuid;index"`
}
```

**Pros**:
- ✅ Normalized database structure
- ✅ Easier to manage album metadata
- ✅ Can add album-level features later

**Cons**:
- ❌ ซับซ้อนมาก
- ❌ ต้องสร้าง table ใหม่
- ❌ Query ต้อง JOIN
- ❌ Over-engineering สำหรับ feature นี้

---

## 🎯 แนวทางที่แนะนำ: Option 1 (Metadata)

### เหตุผล:
1. **ไม่ต้องแก้ Database** - ใช้ Metadata JSONB ที่มีอยู่
2. **Flexible** - เพิ่มฟีเจอร์อื่นได้ในอนาคต
3. **Backward Compatible** - ไม่กระทบ messages เดิม
4. **รวดเร็ว** - implement ได้เลยไม่ต้อง migration

---

## 🛠️ Implementation Plan (Option 1)

### Phase 1: Backend API

#### 1. สร้าง DTO สำหรับ Bulk Upload

**File**: `domain/dto/message_dto.go`
```go
// BulkMessageRequest สำหรับส่งข้อความหลายข้อความพร้อมกัน
type BulkMessageRequest struct {
    Messages []*BulkMessageItem `json:"messages"`
}

// BulkMessageItem ข้อมูลข้อความแต่ละข้อความ
type BulkMessageItem struct {
    MessageType       string      `json:"message_type"` // image, video, file
    MediaURL          string      `json:"media_url"`
    MediaThumbnailURL string      `json:"media_thumbnail_url,omitempty"`
    Caption           string      `json:"caption,omitempty"`
    Metadata          types.JSONB `json:"metadata,omitempty"`
}

// BulkMessageResponse ผลลัพธ์การส่งข้อความหลายข้อความ
type BulkMessageResponse struct {
    Messages []*MessageDTO `json:"messages"`
    AlbumID  string        `json:"album_id,omitempty"`
}
```

---

#### 2. เพิ่ม Repository Method

**File**: `domain/repository/message_repository.go`
```go
type MessageRepository interface {
    // ... existing methods ...

    // BulkCreate สร้างข้อความหลายข้อความพร้อมกัน
    BulkCreate(messages []*models.Message) error

    // GetMessagesByAlbumID ดึงข้อความทั้งหมดใน album
    GetMessagesByAlbumID(albumID string) ([]*models.Message, error)
}
```

**File**: `infrastructure/persistence/postgres/message_repository.go`
```go
// BulkCreate สร้างข้อความหลายข้อความพร้อมกัน
func (r *messageRepository) BulkCreate(messages []*models.Message) error {
    if len(messages) == 0 {
        return nil
    }

    // ใช้ CreateInBatches เพื่อป้องกัน memory overflow
    return r.db.CreateInBatches(messages, 100).Error
}

// GetMessagesByAlbumID ดึงข้อความทั้งหมดใน album
func (r *messageRepository) GetMessagesByAlbumID(albumID string) ([]*models.Message, error) {
    var messages []*models.Message

    err := r.db.
        Where("metadata->>'album_id' = ?", albumID).
        Order("(metadata->>'album_position')::int ASC").
        Find(&messages).Error

    return messages, err
}
```

---

#### 3. เพิ่ม Service Method

**File**: `domain/service/message_service.go`
```go
type MessageService interface {
    // ... existing methods ...

    // SendBulkMessages ส่งข้อความหลายข้อความพร้อมกัน (Album)
    SendBulkMessages(conversationID, userID uuid.UUID, request *dto.BulkMessageRequest) (*dto.BulkMessageResponse, error)
}
```

**File**: `application/serviceimpl/message_service.go`
```go
// SendBulkMessages ส่งข้อความหลายข้อความพร้อมกัน (Album)
func (s *messageService) SendBulkMessages(
    conversationID, userID uuid.UUID,
    request *dto.BulkMessageRequest,
) (*dto.BulkMessageResponse, error) {

    // 1. ตรวจสอบว่า user เป็นสมาชิกของ conversation
    isMember, err := s.conversationRepo.IsMember(conversationID, userID)
    if err != nil {
        return nil, err
    }
    if !isMember {
        return nil, errors.New("user is not a member of this conversation")
    }

    // 2. Validate
    if len(request.Messages) == 0 {
        return nil, errors.New("messages cannot be empty")
    }
    if len(request.Messages) > 10 {
        return nil, errors.New("maximum 10 messages per bulk upload")
    }

    // 3. สร้าง album_id
    albumID := uuid.New().String()
    totalMessages := len(request.Messages)

    // 4. แยก caption (ถ้ามี) - caption จะอยู่ที่ message แรก
    var albumCaption string
    if totalMessages > 0 && request.Messages[0].Caption != "" {
        albumCaption = request.Messages[0].Caption
    }

    // 5. สร้าง messages
    messages := make([]*models.Message, 0, totalMessages)
    messageDTOs := make([]*dto.MessageDTO, 0, totalMessages)

    for i, item := range request.Messages {
        // Validate message type
        if item.MessageType != "image" && item.MessageType != "video" && item.MessageType != "file" {
            return nil, fmt.Errorf("invalid message type: %s", item.MessageType)
        }

        if item.MediaURL == "" {
            return nil, errors.New("media_url is required")
        }

        // สร้าง metadata
        metadata := item.Metadata
        if metadata == nil {
            metadata = make(types.JSONB)
        }

        // เพิ่มข้อมูล album
        metadata["album_id"] = albumID
        metadata["album_position"] = i
        metadata["album_total"] = totalMessages

        // เพิ่ม caption ที่ message แรก
        if i == 0 && albumCaption != "" {
            metadata["album_caption"] = albumCaption
        }

        // สร้าง Message
        message := &models.Message{
            ID:                uuid.New(),
            ConversationID:    conversationID,
            SenderID:          &userID,
            SenderType:        "user",
            MessageType:       item.MessageType,
            Content:           "", // ไม่ใช้ content field สำหรับ media
            MediaURL:          item.MediaURL,
            MediaThumbnailURL: item.MediaThumbnailURL,
            Metadata:          metadata,
            CreatedAt:         time.Now(),
            UpdatedAt:         time.Now(),
        }

        messages = append(messages, message)
    }

    // 6. บันทึกลง database (bulk insert)
    if err := s.messageRepo.BulkCreate(messages); err != nil {
        return nil, fmt.Errorf("failed to create messages: %w", err)
    }

    // 7. อัพเดต last_message ของ conversation (ใช้ message แรก)
    if len(messages) > 0 {
        lastMessageText := fmt.Sprintf("📷 Album (%d photos)", totalMessages)
        if albumCaption != "" {
            lastMessageText = albumCaption
        }

        s.messageRepo.UpdateConversationLastMessage(
            conversationID,
            lastMessageText,
            messages[0].CreatedAt,
        )
    }

    // 8. แปลงเป็น DTO
    for _, msg := range messages {
        messageDTO := s.convertToDTO(msg, userID)
        messageDTOs = append(messageDTOs, messageDTO)
    }

    // 9. ส่ง WebSocket notification
    if s.notificationService != nil {
        // ส่ง notification แค่ครั้งเดียวสำหรับ album
        s.notificationService.NotifyNewMessage(conversationID, map[string]interface{}{
            "type":       "album",
            "album_id":   albumID,
            "messages":   messageDTOs,
            "total":      totalMessages,
            "caption":    albumCaption,
        })
    }

    return &dto.BulkMessageResponse{
        Messages: messageDTOs,
        AlbumID:  albumID,
    }, nil
}

// convertToDTO แปลง Message เป็น DTO (helper)
func (s *messageService) convertToDTO(message *models.Message, userID uuid.UUID) *dto.MessageDTO {
    // ดึงข้อมูล sender
    var senderDTO *dto.UserDTO
    if message.SenderID != nil {
        sender, _ := s.userRepo.FindByID(*message.SenderID)
        if sender != nil {
            senderDTO = &dto.UserDTO{
                ID:              sender.ID.String(),
                Username:        sender.Username,
                DisplayName:     sender.DisplayName,
                ProfileImageURL: sender.ProfileImageURL,
            }
        }
    }

    // Check if message is read
    isRead := false
    if message.SenderID != nil && *message.SenderID != userID {
        isRead, _ = s.messageReadRepo.IsMessageRead(message.ID, userID)
    }

    return &dto.MessageDTO{
        ID:                message.ID.String(),
        ConversationID:    message.ConversationID.String(),
        Sender:            senderDTO,
        SenderType:        message.SenderType,
        MessageType:       message.MessageType,
        Content:           message.Content,
        MediaURL:          message.MediaURL,
        MediaThumbnailURL: message.MediaThumbnailURL,
        Metadata:          message.Metadata,
        CreatedAt:         message.CreatedAt,
        UpdatedAt:         message.UpdatedAt,
        IsDeleted:         message.IsDeleted,
        IsEdited:          message.IsEdited,
        IsRead:            isRead,
    }
}
```

---

#### 4. เพิ่ม Handler

**File**: `interfaces/api/handler/message_handler.go`
```go
// SendBulkMessages จัดการคำขอส่งข้อความหลายข้อความพร้อมกัน
func (h *MessageHandler) SendBulkMessages(c *fiber.Ctx) error {
    // ดึง User ID จาก context
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

    // รับข้อมูลจาก request body
    var request dto.BulkMessageRequest
    if err := c.BodyParser(&request); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "Invalid request body: " + err.Error(),
        })
    }

    // เรียกใช้ service
    response, err := h.messageService.SendBulkMessages(conversationID, userID, &request)
    if err != nil {
        statusCode := fiber.StatusInternalServerError

        // ตรวจสอบประเภทข้อผิดพลาด
        if err.Error() == "user is not a member of this conversation" {
            statusCode = fiber.StatusForbidden
        } else if strings.Contains(err.Error(), "maximum") ||
                  strings.Contains(err.Error(), "required") ||
                  strings.Contains(err.Error(), "invalid") {
            statusCode = fiber.StatusBadRequest
        }

        return c.Status(statusCode).JSON(fiber.Map{
            "success": false,
            "message": err.Error(),
        })
    }

    return c.Status(fiber.StatusCreated).JSON(fiber.Map{
        "success": true,
        "message": "Messages sent successfully",
        "data":    response,
    })
}
```

---

#### 5. ลงทะเบียน Route

**File**: `interfaces/api/routes/message_routes.go`
```go
func SetupMessageRoutes(router fiber.Router, messageHandler *handler.MessageHandler, ...) {
    messages := router.Group("/messages")
    messages.Use(middleware.Protected())

    // ... existing routes ...

    // Bulk upload
    messages.Post("/:conversationId/bulk", messageHandler.SendBulkMessages)
}
```

---

### Phase 2: Database Optimization (Optional)

#### เพิ่ม GIN Index สำหรับ album_id

**Migration File**: `migrations/xxx_add_album_id_index.sql`
```sql
-- สร้าง GIN index สำหรับ metadata->album_id เพื่อเพิ่มความเร็วใน query
CREATE INDEX idx_messages_album_id ON messages USING GIN ((metadata->'album_id'));

-- Index สำหรับ album_position (ถ้าต้องการ sort เร็วขึ้น)
CREATE INDEX idx_messages_album_position ON messages ((CAST(metadata->>'album_position' AS INTEGER)));
```

**ประโยชน์**:
- Query `WHERE metadata->>'album_id' = 'xxx'` เร็วขึ้น 10-100 เท่า
- Support full-text search ใน JSONB

---

### Phase 3: Frontend Integration

#### API Usage

**1. Upload หลายไฟล์ไปที่ Storage ก่อน**
```typescript
// 1. Upload files to storage (S3, CloudFlare, etc.)
const uploadFile = async (file: File) => {
  const formData = new FormData()
  formData.append('file', file)

  const response = await fetch('/api/upload', {
    method: 'POST',
    body: formData
  })

  const data = await response.json()
  return {
    media_url: data.url,
    media_thumbnail_url: data.thumbnail_url
  }
}

// Upload all files
const files = [file1, file2, file3, file4]
const uploadedFiles = await Promise.all(
  files.map(file => uploadFile(file))
)
```

**2. ส่ง Bulk Message API**
```typescript
// 2. Send bulk messages
const sendAlbum = async (conversationId: string, files: any[], caption: string) => {
  const response = await fetch(`/api/messages/${conversationId}/bulk`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      messages: files.map((file, index) => ({
        message_type: 'image',
        media_url: file.media_url,
        media_thumbnail_url: file.media_thumbnail_url,
        caption: index === 0 ? caption : '' // caption แค่ที่ message แรก
      }))
    })
  })

  return await response.json()
}

// Usage
await sendAlbum(conversationId, uploadedFiles, "Holiday trip 🏖️")
```

**3. แสดงผลเป็น Album Grid**
```typescript
// Component for displaying album
interface AlbumProps {
  messages: Message[]
  albumId: string
}

function AlbumView({ messages, albumId }: AlbumProps) {
  // Group messages by album_id
  const albumMessages = messages.filter(
    msg => msg.metadata?.album_id === albumId
  ).sort((a, b) =>
    (a.metadata?.album_position || 0) - (b.metadata?.album_position || 0)
  )

  const caption = albumMessages[0]?.metadata?.album_caption || ''

  return (
    <div className="album">
      {/* Grid layout */}
      <div className={`album-grid grid-${albumMessages.length}`}>
        {albumMessages.map(msg => (
          <img
            key={msg.id}
            src={msg.media_thumbnail_url || msg.media_url}
            alt=""
            onClick={() => openLightbox(msg.id)}
          />
        ))}
      </div>

      {/* Caption */}
      {caption && <p className="album-caption">{caption}</p>}
    </div>
  )
}
```

**4. CSS Grid Layout**
```css
/* 2 photos - 1x2 */
.album-grid.grid-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 4px;
}

/* 3 photos - 1+2 */
.album-grid.grid-3 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  grid-template-rows: auto auto;
  gap: 4px;
}
.album-grid.grid-3 img:first-child {
  grid-column: 1 / -1; /* Full width */
}

/* 4 photos - 2x2 */
.album-grid.grid-4 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  grid-template-rows: 1fr 1fr;
  gap: 4px;
}

/* 5+ photos - 2x3 or custom */
.album-grid.grid-5,
.album-grid.grid-6 {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 4px;
}
```

---

## 📊 สรุปไฟล์ที่ต้องสร้าง/แก้ไข

### Backend

#### ไฟล์ที่ต้องแก้ไข:
1. ✏️ `domain/dto/message_dto.go` - เพิ่ม BulkMessageRequest, BulkMessageResponse DTOs
2. ✏️ `domain/repository/message_repository.go` - เพิ่ม interface methods (BulkCreate, GetMessagesByAlbumID)
3. ✏️ `infrastructure/persistence/postgres/message_repository.go` - implement 2 methods
4. ✏️ `domain/service/message_service.go` - เพิ่ม interface method (SendBulkMessages)
5. ✏️ `application/serviceimpl/message_service.go` - implement SendBulkMessages (ใหญ่สุด ~150 lines)
6. ✏️ `interfaces/api/handler/message_handler.go` - เพิ่ม SendBulkMessages handler
7. ✏️ `interfaces/api/routes/message_routes.go` - ลงทะเบียน route ใหม่

#### ไฟล์ที่ต้องสร้างใหม่ (Optional):
8. 🆕 `migrations/xxx_add_album_id_index.sql` - GIN index สำหรับ performance (optional)

### Frontend

#### Components ใหม่:
1. 🆕 `AlbumUploader.tsx` - Component สำหรับเลือกและอัพโหลดหลายไฟล์
2. 🆕 `AlbumView.tsx` - Component สำหรับแสดง album เป็น grid
3. 🆕 `AlbumLightbox.tsx` - Lightbox สำหรับดูรูปเต็ม

---

## ⚡ ประมาณการเวลาการพัฒนา

### Backend Implementation
- **Phase 1**: Repository & Service (2-3 ชั่วโมง)
- **Phase 2**: Handler & Routes (1 ชั่วโมง)
- **Phase 3**: Testing & Bug Fix (1-2 ชั่วโมง)

**รวม Backend**: 4-6 ชั่วโมง (ครึ่งวัน)

### Frontend Implementation
- **Phase 1**: File Upload UI (2-3 ชั่วโมง)
- **Phase 2**: Album Display Grid (2-3 ชั่วโมง)
- **Phase 3**: Lightbox & Interactions (2-3 ชั่วโมง)

**รวม Frontend**: 6-9 ชั่วโมง (1 วัน)

### Total: 1-1.5 วัน

---

## 🎯 API Specification

### Endpoint
```
POST /messages/:conversationId/bulk
```

### Request Body
```json
{
  "messages": [
    {
      "message_type": "image",
      "media_url": "https://storage.com/photo1.jpg",
      "media_thumbnail_url": "https://storage.com/thumb1.jpg",
      "caption": "Holiday trip 🏖️",
      "metadata": {}
    },
    {
      "message_type": "image",
      "media_url": "https://storage.com/photo2.jpg",
      "media_thumbnail_url": "https://storage.com/thumb2.jpg"
    },
    {
      "message_type": "image",
      "media_url": "https://storage.com/photo3.jpg",
      "media_thumbnail_url": "https://storage.com/thumb3.jpg"
    },
    {
      "message_type": "image",
      "media_url": "https://storage.com/photo4.jpg",
      "media_thumbnail_url": "https://storage.com/thumb4.jpg"
    }
  ]
}
```

### Response (Success)
```json
{
  "success": true,
  "message": "Messages sent successfully",
  "data": {
    "album_id": "abc-123-def-456",
    "messages": [
      {
        "id": "msg-1",
        "conversation_id": "conv-id",
        "sender": { /* user info */ },
        "message_type": "image",
        "media_url": "https://storage.com/photo1.jpg",
        "media_thumbnail_url": "https://storage.com/thumb1.jpg",
        "metadata": {
          "album_id": "abc-123-def-456",
          "album_position": 0,
          "album_total": 4,
          "album_caption": "Holiday trip 🏖️"
        },
        "created_at": "2025-01-15T10:30:00Z"
      },
      {
        "id": "msg-2",
        "message_type": "image",
        "media_url": "https://storage.com/photo2.jpg",
        "metadata": {
          "album_id": "abc-123-def-456",
          "album_position": 1,
          "album_total": 4
        },
        "created_at": "2025-01-15T10:30:01Z"
      },
      // ... msg-3, msg-4
    ]
  }
}
```

### Response (Error - Not Member)
```json
{
  "success": false,
  "message": "user is not a member of this conversation"
}
```

### Response (Error - Too Many)
```json
{
  "success": false,
  "message": "maximum 10 messages per bulk upload"
}
```

### Response (Error - Invalid Type)
```json
{
  "success": false,
  "message": "invalid message type: document"
}
```

---

## 🔒 Validation & Constraints

### Backend Validation
1. ✅ **Maximum items**: สูงสุด 10 ข้อความต่อ request
2. ✅ **Message type**: เฉพาะ `image`, `video`, `file`
3. ✅ **Media URL required**: ต้องมี `media_url`
4. ✅ **Membership check**: ต้องเป็นสมาชิกของ conversation
5. ✅ **Empty check**: ต้องมีอย่างน้อย 1 message

### Metadata Structure
```typescript
interface AlbumMetadata {
  album_id: string         // UUID of album
  album_position: number   // 0-based index
  album_total: number      // Total messages in album
  album_caption?: string   // Caption (only in first message)
}
```

---

## 🚀 Benefits

### ข้อดีของแนวทางนี้:

1. **ไม่ต้อง Migrate Database** ✅
   - ใช้ Metadata JSONB ที่มีอยู่
   - Backward compatible 100%

2. **Performance ดี** ⚡
   - Bulk Insert (1 query แทน N queries)
   - GIN Index รองรับ (optional)
   - Notification ส่งครั้งเดียวต่อ album

3. **Flexible** 🎨
   - เพิ่มฟีเจอร์อื่นใน metadata ได้ตลอด
   - รองรับ album ขนาดต่างๆ (2-10 items)

4. **User Experience เหมือน Telegram** 📱
   - แสดงเป็น grid
   - มี caption
   - Jump to message ได้

5. **Developer-Friendly** 💻
   - Code ไม่ซับซ้อน
   - Easy to test
   - Easy to maintain

---

## ✅ สรุปท้ายเอกสาร

### ความเป็นไปได้: ✅ เป็นไปได้ 100%

**เหตุผล**:
- ✅ Message Model มี Metadata JSONB อยู่แล้ว
- ✅ GORM รองรับ CreateInBatches
- ✅ NotificationService พร้อมใช้งาน
- ✅ ไม่ต้องแก้ Database Schema

**แนวทางที่แนะนำ**: **Option 1 - ใช้ Metadata (album_id)**

**เวลาที่ใช้**:
- Backend: 4-6 ชั่วโมง
- Frontend: 6-9 ชั่วโมง
- **รวม: 1-1.5 วัน**

**ไฟล์ที่ต้องแก้**: 7 ไฟล์ (Backend)
**ไฟล์ที่ต้องสร้าง**: 3 components (Frontend)

---

**พร้อมเริ่มทำได้เลย!** 🚀
