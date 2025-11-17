# Media Summary & Multiple Upload Analysis

**วันที่**: 2025-11-12
**วัตถุประสงค์**: วิเคราะห์ความเป็นไปได้ในการสรุป media/file/link และการส่งหลายไฟล์พร้อมกัน

---

## 📋 สรุปภาพรวม

| Feature | Status | ความเป็นไปได้ | ความซับซ้อน |
|---------|--------|--------------|------------|
| **Media Summary (Count)** | ❌ ไม่มี | ✅ **ทำได้** | ⭐⭐ กลาง |
| **Media Summary (By Date)** | ❌ ไม่มี | ✅ **ทำได้** | ⭐⭐⭐ สูง |
| **Multiple File Upload** | ❌ ไม่มี | ✅ **ทำได้** | ⭐⭐⭐ สูง |
| **Jump to Media** | ✅ มีแล้ว | ✅ **พร้อมใช้** | ✅ ไม่ต้องทำ |

---

## ✅ 1. Jump to Message - มีอยู่แล้ว

### สิ่งที่มี
```
GET /conversations/:conversationId/messages?target=<message_id>&before_count=10&after_count=10
```

✅ **สามารถ jump ไปยัง media/file/link ได้แล้ว** เพราะ media/file/link ก็คือ message ธรรมดา
แค่ส่ง `message_id` ของ media นั้นๆ ไปก็จะ jump ได้เลย

---

## ❌ 2. Media Summary - ยังไม่มี (แต่ทำได้)

### 2.1 ความต้องการ

**แสดงใน Conversation Overview**:
```
📁 Media:   125 items
📄 Files:   43 items
🔗 Links:   28 items
```

**แสดงแบบละเอียด (By Date)**:
```
Media:
  - Today: 5 photos
  - Yesterday: 12 photos
  - 10 Jan 2024: 8 photos

Files:
  - Today: 2 files
  - 5 Jan 2024: 3 files
```

### 2.2 ปัญหาปัจจุบัน

#### ❌ ConversationDTO ไม่มี Media Summary

**File**: `domain/dto/conversation_dto.go`

```go
type ConversationDTO struct {
    ID              uuid.UUID   `json:"id"`
    Type            string      `json:"type"`
    Title           string      `json:"title"`
    // ... other fields ...

    // ❌ ไม่มีฟิลด์เหล่านี้
    // MediaCount      int         `json:"media_count"`
    // FileCount       int         `json:"file_count"`
    // LinkCount       int         `json:"link_count"`
    // MediaSummary    *MediaSummary `json:"media_summary"`
}
```

#### ❌ ไม่มี Repository Method สำหรับ Count

**File**: `infrastructure/persistence/postgres/message_repository.go`

ไม่มีฟังก์ชันเหล่านี้:
- `CountMessagesByType(conversationID, messageType)`
- `CountMediaByDate(conversationID)`
- `GetMediaSummary(conversationID)`

---

## 🔧 3. วิธีแก้ไข - Media Summary

### 3.1 Option 1: Simple Count (แนะนำสำหรับเริ่มต้น)

#### 3.1.1 เพิ่ม Repository Method

**File**: `infrastructure/persistence/postgres/message_repository.go`

```go
// CountMessagesByType นับข้อความตามประเภท
func (r *messageRepository) CountMessagesByType(
    conversationID uuid.UUID,
    messageType string,
) (int64, error) {
    var count int64

    err := r.db.Model(&models.Message{}).
        Where("conversation_id = ? AND message_type = ? AND is_deleted = false",
              conversationID, messageType).
        Count(&count).Error

    return count, err
}

// GetMessageTypeSummary สรุปจำนวนข้อความแต่ละประเภท
func (r *messageRepository) GetMessageTypeSummary(
    conversationID uuid.UUID,
) (map[string]int64, error) {
    var results []struct {
        MessageType string
        Count       int64
    }

    err := r.db.Model(&models.Message{}).
        Select("message_type, COUNT(*) as count").
        Where("conversation_id = ? AND is_deleted = false", conversationID).
        Group("message_type").
        Find(&results).Error

    if err != nil {
        return nil, err
    }

    // แปลงเป็น map
    summary := make(map[string]int64)
    for _, result := range results {
        summary[result.MessageType] = result.Count
    }

    return summary, nil
}

// CountMessagesWithLinks นับข้อความที่มีลิงก์
func (r *messageRepository) CountMessagesWithLinks(
    conversationID uuid.UUID,
) (int64, error) {
    var count int64

    // นับข้อความ text ที่มี links ใน metadata
    err := r.db.Model(&models.Message{}).
        Where("conversation_id = ? AND message_type = 'text' AND metadata->>'links' IS NOT NULL AND is_deleted = false",
              conversationID).
        Count(&count).Error

    return count, err
}
```

#### 3.1.2 เพิ่ม Service Method

**File**: `application/serviceimpl/conversation_service.go`

```go
// MediaSummary สรุปข้อมูล media
type MediaSummary struct {
    ImageCount int64 `json:"image_count"`
    VideoCount int64 `json:"video_count"`
    FileCount  int64 `json:"file_count"`
    LinkCount  int64 `json:"link_count"`
    TotalMedia int64 `json:"total_media"`
}

// GetConversationMediaSummary ดึงสรุปข้อมูล media ในการสนทนา
func (s *conversationService) GetConversationMediaSummary(
    conversationID, userID uuid.UUID,
) (*MediaSummary, error) {
    // ตรวจสอบสิทธิ์
    isMember, err := s.conversationRepo.IsMember(conversationID, userID)
    if err != nil || !isMember {
        return nil, errors.New("not a member of this conversation")
    }

    // ดึงสรุปตามประเภท
    typeSummary, err := s.messageRepo.GetMessageTypeSummary(conversationID)
    if err != nil {
        return nil, err
    }

    // นับลิงก์
    linkCount, err := s.messageRepo.CountMessagesWithLinks(conversationID)
    if err != nil {
        linkCount = 0 // ไม่ให้ error ถ้านับลิงก์ไม่ได้
    }

    summary := &MediaSummary{
        ImageCount: typeSummary["image"],
        VideoCount: typeSummary["video"],
        FileCount:  typeSummary["file"],
        LinkCount:  linkCount,
    }

    // นับรวม media (รูป + วิดีโอ)
    summary.TotalMedia = summary.ImageCount + summary.VideoCount

    return summary, nil
}
```

#### 3.1.3 เพิ่ม API Endpoint

**File**: `interfaces/api/handler/conversation_handler.go`

```go
// GetConversationMediaSummary ดึงสรุป media
func (h *ConversationHandler) GetConversationMediaSummary(c *fiber.Ctx) error {
    userID, err := middleware.GetUserUUID(c)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "success": false,
            "message": "Unauthorized",
        })
    }

    conversationID, err := utils.ParseUUIDParam(c, "conversationId")
    if err != nil {
        return err
    }

    summary, err := h.conversationService.GetConversationMediaSummary(conversationID, userID)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "success": false,
            "message": err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "success": true,
        "data":    summary,
    })
}
```

#### 3.1.4 Register Route

**File**: `interfaces/api/routes/conversation_routes.go`

```go
// เพิ่มใน conversation routes
conversations.Get("/:conversationId/media/summary", conversationHandler.GetConversationMediaSummary)
```

#### 3.1.5 Response Format

```json
{
  "success": true,
  "data": {
    "image_count": 125,
    "video_count": 15,
    "file_count": 43,
    "link_count": 28,
    "total_media": 140
  }
}
```

---

### 3.2 Option 2: Detailed Summary (By Date)

#### 3.2.1 Repository Method

```go
// GetMediaSummaryByDate สรุป media แยกตามวันที่
func (r *messageRepository) GetMediaSummaryByDate(
    conversationID uuid.UUID,
    messageType string, // "image", "video", "file"
) ([]map[string]interface{}, error) {
    var results []struct {
        Date  string
        Count int64
    }

    // Query โดย group by date
    err := r.db.Model(&models.Message{}).
        Select("DATE(created_at) as date, COUNT(*) as count").
        Where("conversation_id = ? AND message_type = ? AND is_deleted = false",
              conversationID, messageType).
        Group("DATE(created_at)").
        Order("date DESC").
        Limit(30). // จำกัดแค่ 30 วันล่าสุด
        Find(&results).Error

    if err != nil {
        return nil, err
    }

    // แปลงเป็น map
    summary := make([]map[string]interface{}, len(results))
    for i, result := range results {
        summary[i] = map[string]interface{}{
            "date":  formatDate(result.Date), // "Today", "Yesterday", "10 Jan 2024"
            "count": result.Count,
        }
    }

    return summary, nil
}
```

#### 3.2.2 Response Format

```json
{
  "success": true,
  "data": {
    "images": [
      { "date": "Today", "count": 5 },
      { "date": "Yesterday", "count": 12 },
      { "date": "10 Jan 2024", "count": 8 }
    ],
    "files": [
      { "date": "Today", "count": 2 },
      { "date": "5 Jan 2024", "count": 3 }
    ],
    "links": [
      { "date": "Today", "count": 1 },
      { "date": "Yesterday", "count": 5 }
    ]
  }
}
```

---

### 3.3 Option 3: Include in ConversationDTO (Auto-load)

#### 3.3.1 แก้ไข DTO

**File**: `domain/dto/conversation_dto.go`

```go
type ConversationDTO struct {
    ID              uuid.UUID   `json:"id"`
    Type            string      `json:"type"`
    Title           string      `json:"title"`
    // ... existing fields ...

    // ✅ เพิ่มฟิลด์ใหม่
    MediaSummary    *MediaSummary `json:"media_summary,omitempty"`
}

type MediaSummary struct {
    ImageCount int64 `json:"image_count"`
    VideoCount int64 `json:"video_count"`
    FileCount  int64 `json:"file_count"`
    LinkCount  int64 `json:"link_count"`
    TotalMedia int64 `json:"total_media"`
}
```

#### 3.3.2 แก้ไข convertToConversationDTO

**File**: `application/serviceimpl/conversations_service.go`

```go
func (s *conversationService) convertToConversationDTO(
    conversation *models.Conversation,
    userID uuid.UUID,
) (*dto.ConversationDTO, error) {
    // ... existing code ...

    // ✅ เพิ่มการดึง media summary
    mediaSummary, err := s.GetConversationMediaSummary(conversation.ID, userID)
    if err == nil {
        convDTO.MediaSummary = mediaSummary
    }
    // ถ้า error ก็ไม่เป็นไร ไม่ใส่ summary

    return convDTO, nil
}
```

#### 3.3.3 Response Format (Auto-included)

```json
{
  "success": true,
  "data": {
    "conversations": [
      {
        "id": "uuid",
        "title": "Chat Group",
        "type": "group",
        "media_summary": {
          "image_count": 125,
          "video_count": 15,
          "file_count": 43,
          "link_count": 28,
          "total_media": 140
        }
      }
    ]
  }
}
```

⚠️ **คำเตือน**: Option นี้จะทำให้ API `GET /conversations` ช้าลง เพราะต้อง query หา summary ทุก conversation

**แนะนำ**: ใช้ Option 1 (Separate Endpoint) แทน

---

## ❌ 4. Multiple File Upload - ยังไม่มี (แต่ทำได้)

### 4.1 ปัญหาปัจจุบัน

#### ❌ ส่งได้ทีละไฟล์เท่านั้น

**Current API**:
```
POST /conversations/:conversationId/messages/image
POST /conversations/:conversationId/messages/file
```

**Request Body** (ทีละไฟล์):
```json
{
  "media_url": "https://example.com/photo1.jpg",
  "media_thumbnail_url": "https://example.com/thumb1.jpg",
  "caption": "Photo 1"
}
```

ถ้าต้องการส่ง 10 รูป = ต้องเรียก API 10 ครั้ง ❌

---

### 4.2 วิธีแก้ไข - Bulk Upload

#### 4.2.1 Option A: Bulk Upload Endpoint (แนะนำ)

**New API**:
```
POST /conversations/:conversationId/messages/bulk
```

**Request Body**:
```json
{
  "messages": [
    {
      "message_type": "image",
      "media_url": "https://example.com/photo1.jpg",
      "media_thumbnail_url": "https://example.com/thumb1.jpg",
      "caption": "Photo 1"
    },
    {
      "message_type": "image",
      "media_url": "https://example.com/photo2.jpg",
      "media_thumbnail_url": "https://example.com/thumb2.jpg",
      "caption": "Photo 2"
    },
    {
      "message_type": "file",
      "media_url": "https://example.com/document.pdf",
      "file_name": "document.pdf",
      "file_size": 1024000,
      "file_type": "application/pdf"
    }
  ]
}
```

**Response**:
```json
{
  "success": true,
  "message": "3 messages sent successfully",
  "data": {
    "messages": [
      { "id": "uuid1", "message_type": "image", ... },
      { "id": "uuid2", "message_type": "image", ... },
      { "id": "uuid3", "message_type": "file", ... }
    ],
    "success_count": 3,
    "failed_count": 0
  }
}
```

#### 4.2.2 Implementation - Repository

**File**: `infrastructure/persistence/postgres/message_repository.go`

```go
// BulkCreate สร้างข้อความหลายรายการพร้อมกัน
func (r *messageRepository) BulkCreate(messages []*models.Message) error {
    // GORM รองรับ bulk insert
    return r.db.Create(messages).Error
}
```

#### 4.2.3 Implementation - Service

**File**: `application/serviceimpl/message_service.go`

```go
type BulkMessageRequest struct {
    MessageType       string            `json:"message_type"`
    Content           string            `json:"content,omitempty"`
    MediaURL          string            `json:"media_url,omitempty"`
    MediaThumbnailURL string            `json:"media_thumbnail_url,omitempty"`
    Caption           string            `json:"caption,omitempty"`
    FileName          string            `json:"file_name,omitempty"`
    FileSize          int64             `json:"file_size,omitempty"`
    FileType          string            `json:"file_type,omitempty"`
    Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

type BulkMessageResult struct {
    Messages     []*dto.MessageDTO `json:"messages"`
    SuccessCount int               `json:"success_count"`
    FailedCount  int               `json:"failed_count"`
    Errors       []string          `json:"errors,omitempty"`
}

// SendBulkMessages ส่งข้อความหลายรายการพร้อมกัน
func (s *messageService) SendBulkMessages(
    conversationID, userID uuid.UUID,
    requests []BulkMessageRequest,
) (*BulkMessageResult, error) {
    // 1. ตรวจสอบสิทธิ์
    isMember, err := s.conversationRepo.IsMember(conversationID, userID)
    if err != nil || !isMember {
        return nil, errors.New("not a member of this conversation")
    }

    // 2. จำกัดจำนวน (ไม่ให้ส่งเกิน 10 รายการต่อครั้ง)
    if len(requests) > 10 {
        return nil, errors.New("cannot send more than 10 messages at once")
    }

    // 3. สร้าง messages
    now := time.Now()
    messages := make([]*models.Message, 0, len(requests))
    result := &BulkMessageResult{
        Messages: make([]*dto.MessageDTO, 0),
        Errors:   make([]string, 0),
    }

    for i, req := range requests {
        message := &models.Message{
            ID:             uuid.New(),
            ConversationID: conversationID,
            SenderID:       &userID,
            SenderType:     "user",
            MessageType:    req.MessageType,
            CreatedAt:      now.Add(time.Duration(i) * time.Millisecond), // เพิ่มทีละ 1ms เพื่อให้เรียงตามลำดับ
            UpdatedAt:      now,
            IsDeleted:      false,
        }

        // ตั้งค่าตาม message type
        switch req.MessageType {
        case "text":
            message.Content = req.Content
        case "image":
            message.MediaURL = req.MediaURL
            message.MediaThumbnailURL = req.MediaThumbnailURL
            if req.Caption != "" {
                message.Content = req.Caption
            }
        case "file":
            message.MediaURL = req.MediaURL
            if req.Metadata == nil {
                req.Metadata = make(map[string]interface{})
            }
            req.Metadata["file_name"] = req.FileName
            req.Metadata["file_size"] = req.FileSize
            req.Metadata["file_type"] = req.FileType
        default:
            result.Errors = append(result.Errors, fmt.Sprintf("message %d: invalid message type", i))
            result.FailedCount++
            continue
        }

        if req.Metadata != nil {
            message.Metadata = req.Metadata
        }

        messages = append(messages, message)
    }

    // 4. Bulk insert
    if len(messages) > 0 {
        err = s.messageRepo.BulkCreate(messages)
        if err != nil {
            return nil, err
        }

        // 5. แปลงเป็น DTOs
        for _, msg := range messages {
            dto, err := s.ConvertToMessageDTO(msg, userID)
            if err == nil {
                result.Messages = append(result.Messages, dto)
                result.SuccessCount++
            }
        }

        // 6. ส่ง notifications
        go func() {
            for _, msg := range messages {
                s.notificationService.NotifyNewMessage(conversationID, msg)
            }
        }()

        // 7. Update conversation last message
        if len(messages) > 0 {
            lastMessage := messages[len(messages)-1]
            s.conversationRepo.UpdateLastMessage(
                conversationID,
                lastMessage.Content,
                lastMessage.CreatedAt,
            )
        }
    }

    return result, nil
}
```

#### 4.2.4 Implementation - Handler

**File**: `interfaces/api/handler/message_handler.go`

```go
// SendBulkMessages ส่งข้อความหลายรายการ
func (h *MessageHandler) SendBulkMessages(c *fiber.Ctx) error {
    userID, err := middleware.GetUserUUID(c)
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "success": false,
            "message": "Unauthorized",
        })
    }

    conversationID, err := utils.ParseUUIDParam(c, "conversationId")
    if err != nil {
        return err
    }

    var input struct {
        Messages []serviceimpl.BulkMessageRequest `json:"messages"`
    }

    if err := c.BodyParser(&input); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "Invalid request body",
        })
    }

    if len(input.Messages) == 0 {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "message": "No messages provided",
        })
    }

    result, err := h.messageService.SendBulkMessages(conversationID, userID, input.Messages)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "success": false,
            "message": err.Error(),
        })
    }

    return c.Status(fiber.StatusCreated).JSON(fiber.Map{
        "success": true,
        "message": fmt.Sprintf("%d messages sent successfully", result.SuccessCount),
        "data":    result,
    })
}
```

#### 4.2.5 Register Route

```go
// routes/message_routes.go
conversations.Post("/:conversationId/messages/bulk", messageHandler.SendBulkMessages)
```

---

#### 4.2.6 Option B: Album/Group Message (Advanced)

สร้าง "album" message ที่มีหลายรูปใน 1 message

**Concept**:
- สร้าง 1 message แม่
- เก็บ URL หลายรูปใน metadata

**Metadata Structure**:
```json
{
  "album": true,
  "items": [
    {
      "url": "https://example.com/photo1.jpg",
      "thumbnail": "https://example.com/thumb1.jpg",
      "caption": "Photo 1"
    },
    {
      "url": "https://example.com/photo2.jpg",
      "thumbnail": "https://example.com/thumb2.jpg",
      "caption": "Photo 2"
    }
  ],
  "item_count": 2
}
```

**API**:
```
POST /conversations/:conversationId/messages/album
```

**Request**:
```json
{
  "caption": "Trip to Thailand",
  "items": [
    {
      "media_url": "https://example.com/photo1.jpg",
      "media_thumbnail_url": "https://example.com/thumb1.jpg"
    },
    {
      "media_url": "https://example.com/photo2.jpg",
      "media_thumbnail_url": "https://example.com/thumb2.jpg"
    }
  ]
}
```

**ข้อดี**:
- UI แสดงเป็น album (เหมือน Telegram/WhatsApp)
- ประหยัด database (1 message แทน 10 messages)
- Query เร็วกว่า

**ข้อเสีย**:
- ซับซ้อนกว่า
- Frontend ต้องรองรับ album UI

---

## 📊 5. สรุปและเปรียบเทียบ

### 5.1 Media Summary

| Option | ความเร็ว | ความซับซ้อน | แนะนำ |
|--------|---------|------------|-------|
| **Option 1: Separate Endpoint** | ⚡⚡⚡ เร็ว | ⭐⭐ กลาง | ✅ **แนะนำ** |
| **Option 2: By Date** | ⚡⚡ ปานกลาง | ⭐⭐⭐ สูง | ⚠️ ทำทีหลัง |
| **Option 3: Auto-include** | ⚡ ช้า | ⭐⭐ กลาง | ❌ ไม่แนะนำ |

**แนะนำ**: เริ่มจาก **Option 1** เพราะ:
- ✅ ทำง่ายที่สุด
- ✅ ไม่กระทบ API เดิม
- ✅ เรียกเมื่อต้องการเท่านั้น (ไม่ช้า)

---

### 5.2 Multiple File Upload

| Option | ใช้งาน | Performance | แนะนำ |
|--------|--------|-------------|-------|
| **Option A: Bulk Upload** | ✅ ง่าย | ⚡⚡⚡ เร็ว | ✅ **แนะนำ** |
| **Option B: Album** | ⭐⭐ ซับซ้อน | ⚡⚡⚡⚡ เร็วมาก | ⚠️ Advanced |
| **Current: Loop** | ✅ ใช้ได้ | ⚡ ช้า | ❌ ไม่แนะนำ |

**แนะนำ**: **Option A (Bulk Upload)** เพราะ:
- ✅ ส่งได้หลายไฟล์ในคราวเดียว
- ✅ Performance ดีกว่า loop
- ✅ Database transaction ปลอดภัยกว่า
- ✅ Notification รวมกัน (ไม่ spam)

---

## 🎯 6. แผนการพัฒนาแนะนำ

### Phase 1: Media Summary (1-2 วัน)

**Priority 1**:
1. ✅ เพิ่ม `GetMessageTypeSummary()` ใน repository
2. ✅ เพิ่ม `GetConversationMediaSummary()` ใน service
3. ✅ เพิ่ม API `GET /conversations/:id/media/summary`
4. ✅ Frontend: แสดง badge ใน conversation list

**Result**:
```
📱 Chat Group
   📁 125 media  📄 43 files  🔗 28 links
```

---

### Phase 2: Bulk Upload (2-3 วัน)

**Priority 2**:
1. ✅ เพิ่ม `BulkCreate()` ใน repository
2. ✅ เพิ่ม `SendBulkMessages()` ใน service
3. ✅ เพิ่ม API `POST /conversations/:id/messages/bulk`
4. ✅ Frontend: Multiple file selection
5. ✅ Frontend: Progress indicator

**Result**:
- ส่ง 10 รูปพร้อมกัน ใน 1 request แทนที่ 10 requests

---

### Phase 3: Detailed Summary (Optional)

**Priority 3** (ถ้ามีเวลา):
1. ✅ เพิ่ม `GetMediaSummaryByDate()`
2. ✅ API `GET /conversations/:id/media/summary/detailed`
3. ✅ Frontend: Date grouping UI

---

## 📝 7. ไฟล์ที่ต้องแก้ไข

### Phase 1: Media Summary

1. `domain/repository/message_repository.go` - เพิ่ม interface
2. `infrastructure/persistence/postgres/message_repository.go` - implement
3. `domain/service/conversation_service.go` - เพิ่ม interface
4. `application/serviceimpl/conversation_service.go` - implement
5. `interfaces/api/handler/conversation_handler.go` - add handler
6. `interfaces/api/routes/conversation_routes.go` - register route

### Phase 2: Bulk Upload

1. `domain/repository/message_repository.go` - เพิ่ม interface
2. `infrastructure/persistence/postgres/message_repository.go` - implement
3. `domain/service/message_service.go` - เพิ่ม interface
4. `application/serviceimpl/message_service.go` - implement
5. `interfaces/api/handler/message_handler.go` - add handler
6. `interfaces/api/routes/message_routes.go` - register route

---

## ✅ 8. สรุปคำตอบ

### คำถาม 1: สรุปว่ามี image/video/file/link กี่อัน วันไหนบ้าง

**คำตอบ**: ✅ **ทำได้**

- ✅ Database มี field `message_type` อยู่แล้ว
- ✅ Query COUNT ตาม message_type ได้
- ✅ GROUP BY date ได้
- ❌ แต่ยังไม่มี API สำเร็จรูป (ต้องสร้างใหม่)

**วิธีทำ**: ตาม Phase 1 ข้างบน

---

### คำถาม 2: สามารถ go to ข้อความเหล่านั้นได้เลย

**คำตอบ**: ✅ **ได้แล้ว**

- ✅ มี API `GetMessageContext` อยู่แล้ว
- ✅ ส่ง message_id ของ media/file/link ไปก็ jump ได้เลย
- ✅ ไม่ต้องทำอะไรเพิ่ม

**API**:
```
GET /conversations/:conversationId/messages?target=<media_message_id>&before_count=20&after_count=20
```

---

### คำถาม 3: ส่ง media หลายไฟล์พร้อมกัน

**คำตอบ**: ❌ **ยังไม่ได้** แต่ ✅ **ทำได้**

- ❌ API ปัจจุบันส่งได้ทีละไฟล์
- ✅ สามารถสร้าง Bulk Upload API ได้
- ✅ Database รองรับ bulk insert

**วิธีทำ**: ตาม Phase 2 ข้างบน

---

## 🚀 9. ขั้นตอนถัดไป

1. **ตัดสินใจว่าจะทำ Phase ไหนก่อน**
   - แนะนำ: Phase 1 (Media Summary) → Phase 2 (Bulk Upload)

2. **เริ่มพัฒนา Phase 1**
   - เวลา: 1-2 วัน
   - ความซับซ้อน: ⭐⭐ กลาง

3. **ทดสอบ**
   - Unit tests
   - Integration tests
   - Frontend integration

4. **Deploy**
   - Update API documentation
   - Notify frontend team

---

**สรุปสุดท้าย**:
- ✅ **ทำได้ทั้งหมด**
- ✅ **Database พร้อมแล้ว**
- ⭐⭐ **ความซับซ้อนปานกลาง**
- ⏱️ **ใช้เวลาประมาณ 3-5 วัน**
