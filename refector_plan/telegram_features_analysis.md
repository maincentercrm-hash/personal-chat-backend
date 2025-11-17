# Telegram-like Features Analysis

**วันที่**: 2025-11-12 (Updated)
**วัตถุประสงค์**: ตรวจสอบความพร้อมของระบบในการรองรับฟีเจอร์แบบ Telegram

---

## 📋 สรุปภาพรวม

| Feature | Status | Backend | Frontend |
|---------|--------|---------|----------|
| **Jump to Message** | ✅ 100% | ✅ Complete | ❌ Need UI |
| **Media Gallery** | ✅ 100% Backend | ✅ Complete | ❌ Need UI |
| **File Gallery** | ✅ 100% Backend | ✅ Complete | ❌ Need UI |
| **Link Summary** | ✅ 100% Backend | ✅ Complete | ❌ Need UI |

**สรุป**: Backend เสร็จสมบูรณ์ 100% แล้ว ⚡️ เหลือแค่ Frontend UI

---

## ✅ 1. Jump to Message - เสร็จแล้ว 100%

### 1.1 API Endpoint
```
GET /conversations/:conversationId/messages/context?targetId=xxx&before=10&after=10
```

### 1.2 Implementation

**Service**: `application/serviceimpl/conversations_service.go` (line 603-693)
```go
func (s *conversationService) GetMessageContext(
    conversationID, userID uuid.UUID,
    targetID string,
    beforeCount, afterCount int
) ([]*dto.MessageDTO, bool, bool, error)
```

**Handler**: `interfaces/api/handler/conversation_handler.go` (line 764-806)
```go
func (h *ConversationHandler) GetMessageContext(c *fiber.Ctx) error
```

**Route**: `interfaces/api/routes/conversation_routes.go` (line 33)
```go
conversations.Get("/:conversationId/messages/context", conversationHandler.GetMessageContext)
```

### 1.3 Features
✅ **ดึงข้อความรอบๆ target message** - ทำได้แล้ว
✅ **ตรวจสอบสิทธิ์** - มีการเช็คว่าเป็น member หรือไม่
✅ **Validate target** - เช็คว่า message อยู่ใน conversation นี้
✅ **Has More indicators** - บอกว่ามีข้อความเพิ่มเติมก่อน/หลังหรือไม่
✅ **Sorted by time** - เรียงตามเวลาอัตโนมัติ
✅ **Handler & Route** - ลงทะเบียนเรียบร้อยแล้ว

### 1.4 Response Format
```json
{
  "success": true,
  "data": [
    // 10 messages before target
    // target message
    // 10 messages after target
  ],
  "has_before": true,
  "has_after": false
}
```

### 1.5 สิ่งที่ต้องเพิ่ม (Frontend เท่านั้น)

❌ **Highlight target message** - ไม่มี UI effect
❌ **Scroll to position** - ไม่มีการ auto-scroll
❌ **Visual indicator** - ไม่มี badge หรือ animation

**แนะนำ Frontend Implementation**:
```typescript
function jumpToMessage(messageId: string) {
  // 1. Fetch context
  const response = await fetch(
    `/api/conversations/${conversationId}/messages/context?targetId=${messageId}&before=20&after=20`
  )

  // 2. Replace messages in view
  setMessages(response.data)

  // 3. Scroll to target
  const targetElement = document.getElementById(`message-${messageId}`)
  targetElement?.scrollIntoView({ behavior: 'smooth', block: 'center' })

  // 4. Highlight target
  targetElement?.classList.add('highlighted')
  setTimeout(() => targetElement?.classList.remove('highlighted'), 2000)
}
```

---

## ✅ 2. Media Gallery - เสร็จแล้ว 100% (Backend)

### 2.1 API Endpoints

#### 2.1.1 Media Summary (นับจำนวน)
```
GET /conversations/:conversationId/media/summary
```

**Response**:
```json
{
  "success": true,
  "data": {
    "image_count": 125,
    "video_count": 15,
    "file_count": 43,
    "link_count": 28,
    "total_media": 183
  }
}
```

#### 2.1.2 Media List by Type (รายละเอียด + pagination)
```
GET /conversations/:conversationId/media?type=image&limit=20&offset=0
GET /conversations/:conversationId/media?type=video&limit=20&offset=0
GET /conversations/:conversationId/media?type=file&limit=20&offset=0
GET /conversations/:conversationId/media?type=link&limit=20&offset=0
```

**Query Parameters**:
- `type`: image, video, file, link
- `limit`: จำนวนรายการต่อหน้า (default: 20)
- `offset`: เริ่มจากรายการที่ (default: 0)

**Response**:
```json
{
  "success": true,
  "data": [
    {
      "message_id": "abc-123",
      "message_type": "image",
      "media_url": "https://storage.com/image.jpg",
      "thumbnail_url": "https://storage.com/thumb.jpg",
      "created_at": "2025-01-15T10:30:00Z"
    },
    {
      "message_id": "def-456",
      "message_type": "image",
      "media_url": "https://storage.com/image2.jpg",
      "thumbnail_url": "https://storage.com/thumb2.jpg",
      "created_at": "2025-01-15T09:15:00Z"
    }
  ],
  "pagination": {
    "total": 125,
    "limit": 20,
    "offset": 0,
    "has_more": true
  }
}
```

### 2.2 Implementation

**Repository**: `infrastructure/persistence/postgres/message_repository.go`
- `GetMessageTypeSummary()` - นับจำนวนแต่ละประเภท (line 353-381)
- `CountMessagesWithLinks()` - นับข้อความที่มีลิงก์ (line 383-393)
- `GetMediaByType()` - ดึงรายละเอียด media พร้อม pagination (line 395-425)

**Service**: `application/serviceimpl/conversations_service.go`
- `GetConversationMediaSummary()` - สรุปจำนวน (line 906-938)
- `GetConversationMediaByType()` - รายละเอียดแยกตามประเภท (line 940-1012)

**Handler**: `interfaces/api/handler/conversation_handler.go`
- `GetMediaSummary()` - handler สำหรับ summary (line 690-722)
- `GetMediaByType()` - handler สำหรับรายละเอียด (line 724-762)

**Routes**: `interfaces/api/routes/conversation_routes.go` (line 30-33)
```go
conversations.Get("/:conversationId/media/summary", conversationHandler.GetMediaSummary)
conversations.Get("/:conversationId/media", conversationHandler.GetMediaByType)
```

### 2.3 Features
✅ **นับจำนวน media แต่ละประเภท** - image, video, file
✅ **นับจำนวน link** - ข้อความที่มี URL
✅ **ดึงรายละเอียด media** - พร้อม pagination
✅ **รองรับ 4 ประเภท** - image, video, file, link
✅ **ตรวจสอบสิทธิ์** - เช็ค membership
✅ **Metadata support** - file_name, file_size สำหรับ file type
✅ **เรียงตามเวลา** - DESC (ใหม่สุดก่อน)
✅ **Pagination** - limit, offset, has_more

### 2.4 สิ่งที่ต้องเพิ่ม (Frontend เท่านั้น)

**ขั้นตอนที่ 1: แสดง Summary ใน Conversation Info**
```typescript
// ConversationInfo.tsx
const [summary, setSummary] = useState(null)

useEffect(() => {
  fetch(`/api/conversations/${conversationId}/media/summary`)
    .then(res => res.json())
    .then(data => setSummary(data.data))
}, [conversationId])

return (
  <div className="media-summary">
    <div onClick={() => openGallery('image')}>
      📷 Photos: {summary?.image_count || 0}
    </div>
    <div onClick={() => openGallery('video')}>
      🎥 Videos: {summary?.video_count || 0}
    </div>
    <div onClick={() => openGallery('file')}>
      📁 Files: {summary?.file_count || 0}
    </div>
    <div onClick={() => openGallery('link')}>
      🔗 Links: {summary?.link_count || 0}
    </div>
  </div>
)
```

**ขั้นตอนที่ 2: สร้าง Media Gallery UI**
```typescript
// MediaGallery.tsx
interface MediaGalleryProps {
  conversationId: string
  mediaType: 'image' | 'video' | 'file' | 'link'
}

function MediaGallery({ conversationId, mediaType }: MediaGalleryProps) {
  const [items, setItems] = useState([])
  const [pagination, setPagination] = useState(null)
  const [offset, setOffset] = useState(0)

  const loadMedia = async (newOffset = 0) => {
    const response = await fetch(
      `/api/conversations/${conversationId}/media?type=${mediaType}&limit=20&offset=${newOffset}`
    )
    const data = await response.json()

    if (newOffset === 0) {
      setItems(data.data)
    } else {
      setItems([...items, ...data.data])
    }
    setPagination(data.pagination)
    setOffset(newOffset)
  }

  useEffect(() => {
    loadMedia(0)
  }, [conversationId, mediaType])

  const handleLoadMore = () => {
    if (pagination?.has_more) {
      loadMedia(offset + 20)
    }
  }

  const handleItemClick = (messageId: string) => {
    jumpToMessage(messageId)
    closeGallery()
  }

  return (
    <div className="media-gallery">
      <h2>
        {mediaType === 'image' && '📷 Photos'}
        {mediaType === 'video' && '🎥 Videos'}
        {mediaType === 'file' && '📁 Files'}
        {mediaType === 'link' && '🔗 Links'}
        ({pagination?.total || 0})
      </h2>

      {/* Grid for images/videos */}
      {(mediaType === 'image' || mediaType === 'video') && (
        <div className="media-grid">
          {items.map(item => (
            <div
              key={item.message_id}
              className="media-item"
              onClick={() => handleItemClick(item.message_id)}
            >
              <img
                src={item.thumbnail_url || item.media_url}
                alt=""
              />
              {mediaType === 'video' && <div className="play-icon">▶</div>}
            </div>
          ))}
        </div>
      )}

      {/* List for files */}
      {mediaType === 'file' && (
        <div className="file-list">
          {items.map(item => (
            <div
              key={item.message_id}
              className="file-item"
              onClick={() => handleItemClick(item.message_id)}
            >
              <div className="file-icon">📄</div>
              <div className="file-info">
                <div className="file-name">{item.file_name}</div>
                <div className="file-size">{formatFileSize(item.file_size)}</div>
                <div className="file-date">{formatDate(item.created_at)}</div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* List for links */}
      {mediaType === 'link' && (
        <div className="link-list">
          {items.map(item => (
            <div
              key={item.message_id}
              className="link-item"
              onClick={() => handleItemClick(item.message_id)}
            >
              <div className="link-content">{item.content}</div>
              <div className="link-date">{formatDate(item.created_at)}</div>
            </div>
          ))}
        </div>
      )}

      {/* Load More Button */}
      {pagination?.has_more && (
        <button onClick={handleLoadMore}>
          Load More
        </button>
      )}
    </div>
  )
}
```

---

## ✅ 3. File Gallery - เสร็จแล้ว 100% (Backend)

### 3.1 API Endpoint
```
GET /conversations/:conversationId/media?type=file&limit=20&offset=0
```

**ใช้ API เดียวกับ Media Gallery** แต่ส่ง `type=file`

### 3.2 Response Format
```json
{
  "success": true,
  "data": [
    {
      "message_id": "uuid",
      "message_type": "file",
      "file_name": "document.pdf",
      "file_size": 1024000,
      "media_url": "https://storage.com/files/document.pdf",
      "created_at": "2025-11-12T10:30:00Z"
    }
  ],
  "pagination": {
    "total": 43,
    "limit": 20,
    "offset": 0,
    "has_more": true
  }
}
```

### 3.3 Frontend Implementation
ใช้ `MediaGallery` component เดียวกัน แต่ส่ง `mediaType="file"`

---

## ✅ 4. Link Summary - เสร็จแล้ว 100% (Backend)

### 4.1 API Endpoint
```
GET /conversations/:conversationId/media?type=link&limit=20&offset=0
```

**ใช้ API เดียวกับ Media Gallery** แต่ส่ง `type=link`

### 4.2 Response Format
```json
{
  "success": true,
  "data": [
    {
      "message_id": "uuid",
      "message_type": "text",
      "content": "Check this out: https://example.com",
      "metadata": {
        "links": ["https://example.com"]
      },
      "created_at": "2025-11-12T10:30:00Z"
    }
  ],
  "pagination": {
    "total": 28,
    "limit": 20,
    "offset": 0,
    "has_more": true
  }
}
```

### 4.3 Link Detection
Links จะถูกจับและเก็บใน `metadata.links` โดยอัตโนมัติเมื่อส่งข้อความ

### 4.4 Frontend Implementation
ใช้ `MediaGallery` component เดียวกัน แต่ส่ง `mediaType="link"`

---

## 🎯 5. สรุปไฟล์ที่ถูกสร้าง/แก้ไข

### ✅ Backend Files (เสร็จแล้ว 100%)

#### Repository Layer
- ✅ `domain/repository/message_repository.go` - เพิ่ม 3 interface methods
- ✅ `infrastructure/persistence/postgres/message_repository.go` - implement 3 methods

#### Service Layer
- ✅ `domain/service/conversations_service.go` - เพิ่ม 2 interface methods
- ✅ `application/serviceimpl/conversations_service.go` - implement 2 methods

#### DTO
- ✅ `domain/dto/media_dto.go` - สร้างไฟล์ใหม่ (MediaSummaryDTO, MediaItemDTO, MediaListDTO)

#### Handler & Routes
- ✅ `interfaces/api/handler/conversation_handler.go` - เพิ่ม 3 handlers
- ✅ `interfaces/api/routes/conversation_routes.go` - ลงทะเบียน 3 routes

### ❌ Frontend Files (ต้องสร้างใหม่)

#### Components
- ❌ `ConversationInfo.tsx` - เพิ่มส่วนแสดง media summary
- ❌ `MediaGallery.tsx` - Gallery component สำหรับแสดง media/file/link
- ❌ `MessageHighlight.tsx` - Highlight effect สำหรับ jump to message

#### Utils
- ❌ `useMediaGallery.ts` - Hook สำหรับจัดการ gallery state
- ❌ `formatters.ts` - Format file size, date, etc.

---

## 📊 6. สถิติการทำงาน

### Backend Implementation
- ✅ **Repository**: 3 methods เพิ่มแล้ว
- ✅ **Service**: 2 methods เพิ่มแล้ว
- ✅ **Handler**: 3 handlers เพิ่มแล้ว
- ✅ **Routes**: 3 routes ลงทะเบียนแล้ว
- ✅ **DTO**: 1 ไฟล์สร้างใหม่
- ✅ **Compilation**: ผ่านแล้ว ไม่มี errors

### API Endpoints ที่พร้อมใช้งาน
1. ✅ `GET /conversations/:conversationId/messages/context` - Jump to Message
2. ✅ `GET /conversations/:conversationId/media/summary` - Media Summary
3. ✅ `GET /conversations/:conversationId/media?type=image` - Image Gallery
4. ✅ `GET /conversations/:conversationId/media?type=video` - Video Gallery
5. ✅ `GET /conversations/:conversationId/media?type=file` - File Gallery
6. ✅ `GET /conversations/:conversationId/media?type=link` - Link Gallery

---

## 🚀 7. แผนการพัฒนา Frontend

### Priority 1: Media Summary (2-4 ชั่วโมง)
**งาน**:
- [ ] เพิ่ม media summary ใน Conversation Info
- [ ] แสดงจำนวน images, videos, files, links
- [ ] คลิกเพื่อเปิด gallery

### Priority 2: Media Gallery UI (1-2 วัน)
**งาน**:
- [ ] สร้าง MediaGallery component
- [ ] Image/Video grid view
- [ ] File list view
- [ ] Link list view
- [ ] Pagination (load more)
- [ ] คลิกเพื่อ jump to message

### Priority 3: Jump to Message UI (2-4 ชั่วโมง)
**งาน**:
- [ ] Scroll to message
- [ ] Highlight effect
- [ ] Visual indicator

---

## 📖 8. ตัวอย่าง API Calls

### 8.1 Get Media Summary
```javascript
const response = await fetch(
  `/api/conversations/${conversationId}/media/summary`
)
// Response: { image_count: 125, video_count: 15, file_count: 43, link_count: 28 }
```

### 8.2 Get Image Gallery
```javascript
const response = await fetch(
  `/api/conversations/${conversationId}/media?type=image&limit=20&offset=0`
)
```

### 8.3 Get File Gallery
```javascript
const response = await fetch(
  `/api/conversations/${conversationId}/media?type=file&limit=20&offset=0`
)
```

### 8.4 Get Link Gallery
```javascript
const response = await fetch(
  `/api/conversations/${conversationId}/media?type=link&limit=20&offset=0`
)
```

### 8.5 Jump to Message
```javascript
const response = await fetch(
  `/api/conversations/${conversationId}/messages/context?targetId=${messageId}&before=20&after=20`
)
```

---

## ✅ 9. สรุปสถานะปัจจุบัน

### Backend: 100% เสร็จสมบูรณ์ ✅
- ✅ Jump to Message API พร้อมใช้งาน
- ✅ Media Summary API พร้อมใช้งาน
- ✅ Media Gallery API พร้อมใช้งาน (4 ประเภท: image, video, file, link)
- ✅ Pagination support
- ✅ Permission check
- ✅ Compilation success

### Frontend: ต้องพัฒนา 100% ❌
- ❌ Media Summary UI
- ❌ Gallery Components
- ❌ Jump to Message UI
- ❌ Highlight Effects

**เวลาโดยประมาณสำหรับ Frontend**: 3-5 วัน

---

**สรุปท้ายเอกสาร**:
🎉 **Backend เสร็จสมบูรณ์ 100%** - พร้อมส่งต่อให้ Frontend Team ทำ UI ได้เลย!
