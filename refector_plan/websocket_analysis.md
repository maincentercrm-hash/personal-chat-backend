# WebSocket System Analysis

**วันที่สร้าง**: 2025-11-12
**เวอร์ชัน**: 1.0
**ผู้วิเคราะห์**: Claude Code

---

## สารบัญ

1. [ภาพรวมระบบ](#ภาพรวมระบบ)
2. [สถาปัตยกรรม WebSocket](#สถาปัตยกรรม-websocket)
3. [Message Types ที่รองรับ](#message-types-ที่รองรับ)
4. [การทำงานของแต่ละส่วน](#การทำงานของแต่ละส่วน)
5. [ปัญหาที่พบและวิธีแก้ไข](#ปัญหาที่พบและวิธีแก้ไข)
6. [วิธีการใช้งานที่ถูกต้อง](#วิธีการใช้งานที่ถูกต้อง)

---

## ภาพรวมระบบ

ระบบ WebSocket ของ backend นี้ออกแบบมาเพื่อ **รับ-ส่งการแจ้งเตือนแบบ Real-time** ระหว่าง users ในระบบแชท

### หน้าที่หลัก:
- ✅ แจ้งเตือนข้อความใหม่แบบ real-time
- ✅ แจ้งเตือนสถานะการอ่านข้อความ
- ✅ แจ้งเตือนการแก้ไข/ลบข้อความ
- ✅ แจ้งเตือนสถานะออนไลน์/ออฟไลน์ของผู้ใช้
- ✅ แจ้งเตือนคำขอเป็นเพื่อน
- ✅ แจ้งเตือนการสร้าง/อัปเดตการสนทนา
- ⚠️ รับข้อความจาก client (มีปัญหา - ไม่แนะนำให้ใช้)

---

## สถาปัตยกรรม WebSocket

### 1. **Hub** (`interfaces/websocket/hub.go`)
**หน้าที่**: ตัวกลางจัดการ WebSocket connections ทั้งหมด

**โครงสร้างข้อมูล**:
```go
type Hub struct {
    clients              map[uuid.UUID]*Client           // เก็บ clients ทั้งหมด
    userConnections      map[uuid.UUID][]uuid.UUID       // userID -> clientIDs
    conversationSubs     map[uuid.UUID][]uuid.UUID       // conversationID -> clientIDs
    userStatusSubs       map[uuid.UUID][]uuid.UUID       // userID -> subscribers
    handlers             map[string]MessageHandler        // message handlers
    conversationService  service.ConversationService     // ดึงข้อมูลการสนทนา
    notificationService  service.NotificationService     // ส่ง notifications
    register             chan *Client                     // channel สำหรับ register client
    unregister           chan *Client                     // channel สำหรับ unregister client
    broadcast            chan *BroadcastMessage           // channel สำหรับ broadcast
}
```

**การทำงาน**:
1. รัน goroutine `Run()` ตลอดเวลา รอรับ events
2. เมื่อมี client ใหม่ เก็บไว้ใน `clients` และ `userConnections`
3. Subscribe การสนทนาของ user อัตโนมัติ (5 การสนทนาแรก)
4. ส่ง broadcast messages ไปยัง clients ที่เกี่ยวข้อง

---

### 2. **Client** (`interfaces/websocket/client.go`)
**หน้าที่**: แทนการเชื่อมต่อ WebSocket ของ user แต่ละคน

**โครงสร้างข้อมูล**:
```go
type Client struct {
    ID                   uuid.UUID              // Client ID (unique per connection)
    UserID               uuid.UUID              // User ID
    BusinessID           *uuid.UUID             // Business ID (ถ้ามี)
    ActiveConversationID *uuid.UUID             // การสนทนาที่กำลังเปิดอยู่
    Conn                 *websocket.Conn        // WebSocket connection
    Send                 chan []byte            // Channel สำหรับส่งข้อความ
    Hub                  *Hub                   // อ้างอิงไปยัง Hub
    IsAlive              bool                   // สถานะการเชื่อมต่อ
    LastPingTime         time.Time              // เวลา ping ล่าสุด
    RateLimiter          *RateLimiter           // จำกัดอัตราการส่งข้อความ
}
```

**การทำงาน**:
- `ReadPump()`: อ่านข้อความจาก client และส่งไป handler ที่เหมาะสม
- `WritePump()`: เขียนข้อความไปยัง client และส่ง ping เป็นระยะ
- Rate limiting: จำกัด 60 ข้อความต่อนาที

---

### 3. **Message Handlers** (`interfaces/websocket/handlers.go`)
**หน้าที่**: จัดการข้อความแต่ละประเภทที่ได้รับจาก client

**Handlers ที่มี**:
- `MessageSendHandler` - รับข้อความจาก client (⚠️ มีปัญหา)
- `MessageTypingHandler` - แจ้งสถานะกำลังพิมพ์
- `MessageReadHandler` - แจ้งสถานะการอ่านข้อความ
- `MessageEditHandler` - แจ้งการแก้ไขข้อความ (⚠️ ไม่ได้แก้ DB)
- `MessageDeleteHandler` - แจ้งการลบข้อความ (⚠️ ไม่ได้ลบ DB)
- `ConversationJoinHandler` - Join การสนทนา
- `ConversationLeaveHandler` - Leave การสนทนา
- `ConversationActiveHandler` - ตั้งค่าการสนทนาที่กำลังเปิด
- `ConversationCreateHandler` - สร้างการสนทนาใหม่
- `ConversationsLoadHandler` - โหลดรายการการสนทนา
- `SubscribeUserStatusHandler` - Subscribe สถานะผู้ใช้
- `UnsubscribeUserStatusHandler` - Unsubscribe สถานะผู้ใช้
- `PingHandler` - Ping/Pong

---

### 4. **Notification Service** (`application/serviceimpl/notification_service.go`)
**หน้าที่**: ส่ง notifications ผ่าน WebSocket ไปยัง clients

**เมธอดสำคัญ**:

#### Messages
- `NotifyNewMessage(conversationID, message)` - แจ้งข้อความใหม่
- `NotifyMessageRead(conversationID, message)` - แจ้งการอ่านข้อความ
- `NotifyMessageEdited(conversationID, message)` - แจ้งการแก้ไขข้อความ
- `NotifyMessageDeleted(conversationID, messageID)` - แจ้งการลบข้อความ
- `NotifyMessageReaction(conversationID, reaction)` - แจ้งการแสดงความรู้สึก

#### Conversations
- `NotifyConversationCreated(userIDs, conversation)` - แจ้งการสร้างการสนทนา
- `NotifyConversationUpdated(conversationID, update)` - แจ้งการอัปเดตการสนทนา
- `NotifyConversationDeleted(conversationID, memberIDs)` - แจ้งการลบการสนทนา

#### Friends
- `NotifyFriendRequestReceived(request)` - แจ้งคำขอเป็นเพื่อน
- `NotifyFriendRequestAccepted(friendship)` - แจ้งการยอมรับคำขอ
- `NotifyFriendRemoved(userID, friendID)` - แจ้งการลบเพื่อน

---

### 5. **WebSocket Adapter** (`infrastructure/adapter/websocket_adapter.go`)
**หน้าที่**: เชื่อมต่อ Notification Service กับ WebSocket Hub

**ทำหน้าที่เป็น Bridge**:
```
NotificationService -> WebSocketAdapter -> Hub -> Clients
```

---

## Message Types ที่รองรับ

### Connection Management
| Type | Direction | Description |
|------|-----------|-------------|
| `connect` | Server → Client | เชื่อมต่อสำเร็จ |
| `disconnect` | Client → Server | ยกเลิกการเชื่อมต่อ |
| `ping` | Client → Server | Ping เพื่อเช็คการเชื่อมต่อ |
| `pong` | Server → Client | Pong ตอบกลับ |

### Chat Messages
| Type | Direction | Description |
|------|-----------|-------------|
| `message.send` | Client → Server | ⚠️ ส่งข้อความ (มีปัญหา) |
| `message.receive` | Server → Client | ได้รับข้อความใหม่ |
| `message.edit` | Both | แก้ไขข้อความ |
| `message.delete` | Both | ลบข้อความ |
| `message.read` | Both | อ่านข้อความ |
| `message.typing` | Client → Server | กำลังพิมพ์ |

### Conversations
| Type | Direction | Description |
|------|-----------|-------------|
| `conversation.create` | Both | สร้างการสนทนาใหม่ |
| `conversation.update` | Server → Client | อัปเดตการสนทนา |
| `conversation.join` | Client → Server | เข้าร่วมการสนทนา |
| `conversation.leave` | Client → Server | ออกจากการสนทนา |
| `conversation.active` | Client → Server | ตั้งค่าการสนทนาที่เปิดอยู่ |
| `conversation.load` | Client → Server | โหลดรายการการสนทนา |
| `conversation.list` | Server → Client | รายการการสนทนา |

### User Status
| Type | Direction | Description |
|------|-----------|-------------|
| `user.status.subscribe` | Client → Server | Subscribe สถานะผู้ใช้ |
| `user.status.unsubscribe` | Client → Server | Unsubscribe สถานะผู้ใช้ |
| `user.online` | Server → Client | ผู้ใช้ออนไลน์ |
| `user.offline` | Server → Client | ผู้ใช้ออฟไลน์ |
| `user.status` | Server → Client | สถานะผู้ใช้ |

### Friends
| Type | Direction | Description |
|------|-----------|-------------|
| `friend.request` | Server → Client | ได้รับคำขอเป็นเพื่อน |
| `friend.accept` | Server → Client | คำขอถูกยอมรับ |
| `friend.remove` | Server → Client | เพื่อนถูกลบ |

### Notifications
| Type | Direction | Description |
|------|-----------|-------------|
| `notification` | Server → Client | การแจ้งเตือนทั่วไป |
| `alert` | Server → Client | การแจ้งเตือนสำคัญ |
| `error` | Server → Client | ข้อผิดพลาด |

---

## การทำงานของแต่ละส่วน

### 1. การเชื่อมต่อ WebSocket

**Endpoint**: `GET /ws/user?token=<JWT_TOKEN>`

**ขั้นตอน**:
```
1. Client ส่ง request พร้อม JWT token
2. Server validate token → ดึง userID
3. สร้าง Client object ใหม่
4. ส่ง client ไป register channel
5. Hub รับ client และ:
   - เก็บใน clients map
   - เพิ่มใน userConnections
   - Subscribe การสนทนา 5 รายการแรกอัตโนมัติ
   - ส่ง user.online notification ให้ผู้ที่ subscribe
   - ส่ง conversation.list กลับไป
6. เริ่ม ReadPump และ WritePump goroutines
```

**Response ตัวอย่าง**:
```json
{
  "type": "connect",
  "data": {
    "message": "Connected successfully",
    "client_id": "xxxx-xxxx-xxxx"
  },
  "timestamp": "2025-11-12T10:00:00Z",
  "success": true
}
```

---

### 2. การ Subscribe การสนทนา

**การทำงาน**:
- เมื่อ client เชื่อมต่อ Hub จะเรียก `loadUserConversations()`
- ดึงการสนทนาของ user จาก ConversationService
- Subscribe การสนทนา 5 รายการแรกอัตโนมัติ
- ส่ง `conversation.list` กลับไป client

**Message ที่ client ได้รับ**:
```json
{
  "type": "conversation.list",
  "data": [
    {
      "id": "conv-id-1",
      "title": "John Doe",
      "type": "direct",
      "last_message_at": "2025-11-12T09:30:00Z",
      "unread_count": 3,
      "is_subscribed": true
    },
    ...
  ],
  "timestamp": "2025-11-12T10:00:00Z",
  "success": true
}
```

---

### 3. การ Join การสนทนา

**Client ส่ง**:
```json
{
  "type": "conversation.join",
  "data": {
    "conversation_id": "xxxx-xxxx-xxxx"
  },
  "timestamp": "2025-11-12T10:00:00Z"
}
```

**Server ทำ**:
1. ตรวจสอบว่า user เป็นสมาชิกการสนทนาหรือไม่
2. ตั้งค่า `client.ActiveConversationID`
3. Subscribe การสนทนา (ถ้ายัง)
4. Broadcast `conversation.user_active` ไปยังสมาชิกอื่น
5. ส่ง `conversation.joined` กลับมา

**Response**:
```json
{
  "type": "conversation.joined",
  "data": {
    "conversation_id": "xxxx-xxxx-xxxx",
    "success": true
  },
  "timestamp": "2025-11-12T10:00:01Z",
  "success": true
}
```

---

### 4. การส่งข้อความ (⚠️ มีปัญหา)

**⚠️ วิธีนี้ไม่แนะนำ - ข้อความไม่ถูกบันทึกใน Database!**

**Client ส่ง**:
```json
{
  "type": "message.send",
  "data": {
    "conversation_id": "xxxx-xxxx-xxxx",
    "content": "Hello!",
    "message_type": "text"
  },
  "timestamp": "2025-11-12T10:00:00Z"
}
```

**ปัญหา**:
- `MessageSendHandler` ไม่ได้บันทึกข้อความลง database
- เพียงแค่ broadcast ไปยังสมาชิกในการสนทนา
- ข้อความหายเมื่อ reload หน้า
- ไม่มี message ID จริงจาก database

**วิธีที่ถูกต้อง**: ใช้ REST API แทน (ดู [วิธีการใช้งานที่ถูกต้อง](#วิธีการใช้งานที่ถูกต้อง))

---

### 5. การรับข้อความใหม่

**เมื่อมีข้อความใหม่ (จาก API)**:
1. API บันทึกข้อความลง database
2. API เรียก `notificationService.NotifyNewMessage()`
3. NotificationService สร้าง MessageDTO พร้อมข้อมูลผู้ส่ง
4. ส่งผ่าน WebSocketAdapter → Hub → Clients ที่ subscribe

**Message ที่ client ได้รับ**:
```json
{
  "type": "message.receive",
  "data": {
    "id": "msg-id-123",
    "conversation_id": "conv-id-456",
    "sender_id": "user-id-789",
    "sender_name": "John Doe",
    "sender_avatar": "https://...",
    "sender_info": {
      "id": "user-id-789",
      "username": "johndoe",
      "display_name": "John Doe",
      "profile_image_url": "https://..."
    },
    "message_type": "text",
    "content": "Hello!",
    "created_at": "2025-11-12T10:00:00Z",
    "is_read": true,
    "read_count": 1
  },
  "timestamp": "2025-11-12T10:00:00Z",
  "success": true
}
```

---

### 6. การ Subscribe สถานะผู้ใช้

**Client ส่ง**:
```json
{
  "type": "user.status.subscribe",
  "data": {
    "user_id": "friend-id-123"
  }
}
```

**Server ทำ**:
1. เพิ่ม clientID ใน `userStatusSubs[friend-id-123]`
2. ส่งสถานะปัจจุบันของผู้ใช้กลับมาทันที

**Response**:
```json
{
  "type": "user.online",  // หรือ user.offline
  "data": {
    "user_id": "friend-id-123",
    "online": true,
    "timestamp": "2025-11-12T10:00:00Z"
  },
  "timestamp": "2025-11-12T10:00:00Z",
  "success": true
}
```

**เมื่อผู้ใช้ออนไลน์/ออฟไลน์**:
- Hub จะส่ง `user.online` หรือ `user.offline` ไปยังทุกคนที่ subscribe

---

### 7. การรับคำขอเป็นเพื่อน

**Flow**:
1. User A เรียก API: `POST /api/friendships` → ส่งคำขอไปหา User B
2. API บันทึกคำขอลง database
3. API เรียก `notificationService.NotifyFriendRequestReceived()`
4. NotificationService ดึงข้อมูล User A จาก database
5. ส่ง notification ผ่าน WebSocket ไปยัง User B

**Message ที่ User B ได้รับ**:
```json
{
  "type": "friend.request",
  "data": {
    "request_id": "req-id-123",
    "user_id": "user-a-id",
    "friend_id": "user-b-id",
    "status": "pending",
    "requested_at": "2025-11-12T10:00:00Z",
    "sender": {
      "id": "user-a-id",
      "username": "usera",
      "display_name": "User A",
      "profile_image_url": "https://..."
    }
  },
  "timestamp": "2025-11-12T10:00:00Z",
  "success": true
}
```

**⚠️ สำคัญ**: User B ต้องเชื่อมต่อ WebSocket ไว้ตลอดเวลา ถึงจะได้รับ notification แบบ real-time

---

### 8. การยอมรับคำขอเป็นเพื่อน

**Flow**:
1. User B เรียก API: `PATCH /api/friendships/:id/accept`
2. API อัปเดตสถานะเป็น "accepted"
3. API เรียก `notificationService.NotifyFriendRequestAccepted()`
4. NotificationService ดึงข้อมูล User B
5. ส่ง notification ไปยัง User A (ผู้ส่งคำขอเดิม)

**Message ที่ User A ได้รับ**:
```json
{
  "type": "friend.accept",
  "data": {
    "friendship_id": "friendship-id-123",
    "user_id": "user-a-id",
    "friend_id": "user-b-id",
    "status": "accepted",
    "accepted_at": "2025-11-12T10:05:00Z",
    "acceptor": {
      "id": "user-b-id",
      "username": "userb",
      "display_name": "User B",
      "profile_image_url": "https://...",
      "last_active_at": "2025-11-12T10:05:00Z"
    }
  },
  "timestamp": "2025-11-12T10:05:00Z",
  "success": true
}
```

---

## ปัญหาที่พบและวิธีแก้ไข

### ❌ ปัญหาที่ 1: ข้อความไม่แสดงระหว่างกัน (User ใน Room เดียวกัน)

**อาการ**:
- User1 และ User2 join ห้องเดียวกัน
- User1 พิมพ์ข้อความ User2 ไม่เห็น
- User2 พิมพ์ข้อความ User1 ไม่เห็น

**สาเหตุ**:
1. **Frontend ส่งข้อความผ่าน WebSocket โดยตรง** (`message.send`)
2. `MessageSendHandler` ไม่ได้บันทึกข้อความลง database
3. เพียงแค่ broadcast ข้อมูลที่ได้รับมาตรงๆ
4. ข้อความไม่มี ID จริง, timestamp จาก server, ข้อมูลผู้ส่งที่สมบูรณ์

**วิธีแก้ที่ถูกต้อง**:
```
❌ เดิม: Frontend → WebSocket (message.send) → Broadcast → Frontend
✅ ใหม่: Frontend → REST API → Database → NotificationService → WebSocket → Frontend
```

**ตัวอย่าง Code (Frontend)**:
```javascript
// ❌ วิธีเดิม (ผิด)
websocket.send({
  type: "message.send",
  data: {
    conversation_id: "xxx",
    content: "Hello"
  }
})

// ✅ วิธีใหม่ (ถูกต้อง)
await fetch('/api/conversations/xxx/messages', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    content: "Hello",
    message_type: "text"
  })
})
// → Server บันทึกลง DB แล้วส่ง notification ผ่าน WebSocket อัตโนมัติ
```

---

### ❌ ปัญหาที่ 2: การแอดเพื่อนไม่แจ้งเตือนแบบ Real-time (ต้องกด F5)

**อาการ**:
- User1 ส่งคำขอเป็นเพื่อนให้ User2
- User2 ไม่ได้รับการแจ้งเตือนทันที
- ต้อง refresh หน้าถึงจะเห็น

**สาเหตุ**:
1. **User2 ไม่ได้เชื่อมต่อ WebSocket**
2. หรือ **Frontend ไม่ได้ listen event** `friend.request`

**วิธีแก้**:

**Backend**: (อันนี้ใช้งานได้แล้ว)
```go
// ใน friendship handler/service
func AcceptFriendRequest(friendshipID uuid.UUID) {
    // 1. บันทึกลง database
    friendship.Status = "accepted"
    repo.Update(friendship)

    // 2. ส่ง notification ผ่าน WebSocket
    notificationService.NotifyFriendRequestAccepted(friendship)
}
```

**Frontend**: (ต้องแก้)
```javascript
// 1. เชื่อมต่อ WebSocket ตั้งแต่ login
const ws = new WebSocket(`ws://localhost:8080/ws/user?token=${token}`)

// 2. Listen event friend.request
ws.addEventListener('message', (event) => {
  const message = JSON.parse(event.data)

  switch(message.type) {
    case 'friend.request':
      // แสดง notification ว่ามีคำขอเป็นเพื่อนใหม่
      showNotification('คำขอเป็นเพื่อนจาก', message.data.sender.display_name)
      // อัปเดต UI
      addFriendRequestToList(message.data)
      break

    case 'friend.accept':
      // แสดง notification ว่าคำขอถูกยอมรับ
      showNotification('ยอมรับคำขอเป็นเพื่อน', message.data.acceptor.display_name)
      // อัปเดต friends list
      addFriendToList(message.data.acceptor)
      break
  }
})
```

---

### ❌ ปัญหาที่ 3: สถานะออนไลน์/ออฟไลน์ไม่อัปเดต

**อาการ**:
- User online/offline แต่ UI ไม่แสดงสถานะ

**สาเหตุ**:
- Frontend ไม่ได้ **subscribe สถานะผู้ใช้**

**วิธีแก้**:

**Frontend**:
```javascript
// เมื่อเข้าหน้าแชทหรือรายชื่อเพื่อน
// Subscribe สถานะของเพื่อนทุกคน
friends.forEach(friend => {
  ws.send(JSON.stringify({
    type: "user.status.subscribe",
    data: {
      user_id: friend.id
    }
  }))
})

// Listen events
ws.addEventListener('message', (event) => {
  const message = JSON.parse(event.data)

  switch(message.type) {
    case 'user.online':
      updateUserStatus(message.data.user_id, true)
      break

    case 'user.offline':
      updateUserStatus(message.data.user_id, false)
      break
  }
})
```

---

### ❌ ปัญหาที่ 4: Conversation ใหม่ไม่แสดงทันที

**อาการ**:
- User1 สร้าง conversation กับ User2
- User2 ไม่เห็น conversation ใหม่ทันที

**สาเหตุ**:
- Frontend ไม่ได้ listen event `conversation.create`

**วิธีแก้**:

**Frontend**:
```javascript
ws.addEventListener('message', (event) => {
  const message = JSON.parse(event.data)

  switch(message.type) {
    case 'conversation.create':
      // เพิ่ม conversation ใหม่ใน list
      addConversationToList(message.data)
      break

    case 'message.receive':
      // แสดงข้อความใหม่
      if (currentConversationId === message.data.conversation_id) {
        appendMessageToChat(message.data)
      }
      // อัปเดต unread count
      updateUnreadCount(message.data.conversation_id)
      break
  }
})
```

---

## วิธีการใช้งานที่ถูกต้อง

### 1. การส่งข้อความ

**✅ ขั้นตอนที่ถูกต้อง**:

**Frontend**:
```javascript
// 1. เชื่อมต่อ WebSocket (ครั้งเดียวตอน login)
const ws = new WebSocket(`ws://localhost:8080/ws/user?token=${jwtToken}`)

// 2. Join conversation ก่อนส่งข้อความ
ws.send(JSON.stringify({
  type: "conversation.join",
  data: {
    conversation_id: conversationId
  }
}))

// 3. ส่งข้อความผ่าน REST API (ไม่ใช่ WebSocket!)
async function sendMessage(conversationId, content) {
  const response = await fetch(`/api/conversations/${conversationId}/messages`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      content: content,
      message_type: "text"
    })
  })

  // ไม่ต้องทำอะไรเพิ่ม - WebSocket จะส่ง notification มาเอง
}

// 4. รับข้อความผ่าน WebSocket
ws.addEventListener('message', (event) => {
  const message = JSON.parse(event.data)

  if (message.type === 'message.receive') {
    // แสดงข้อความใหม่ใน UI
    appendMessage(message.data)
  }
})
```

**Backend Flow**:
```
1. API Endpoint: POST /api/conversations/:id/messages
2. Handler → MessageService.SendTextMessage()
3. MessageService บันทึกข้อความลง database
4. MessageService เรียก NotificationService.NotifyNewMessage()
5. NotificationService สร้าง MessageDTO
6. ส่งผ่าน WebSocketAdapter → Hub
7. Hub broadcast ไปยังทุกคนที่ subscribe conversation นั้น
```

**Backend API Endpoint** (ตัวอย่าง):
```go
// interfaces/api/handler/message_handler.go
func (h *MessageHandler) SendTextMessage(c *fiber.Ctx) error {
    // 1. ดึง userID จาก JWT
    userID := c.Locals("userID").(uuid.UUID)

    // 2. Parse request
    var req struct {
        Content     string `json:"content"`
        MessageType string `json:"message_type"`
    }
    if err := c.BodyParser(&req); err != nil {
        return err
    }

    conversationID, _ := uuid.Parse(c.Params("id"))

    // 3. บันทึกข้อความ (MessageService จะ notify อัตโนมัติ)
    message, err := h.messageService.SendTextMessage(
        conversationID, userID, req.Content, nil,
    )
    if err != nil {
        return err
    }

    // 4. Return ข้อความที่บันทึกแล้ว
    return c.JSON(message)
}
```

---

### 2. การแอดเพื่อน

**✅ ขั้นตอนที่ถูกต้อง**:

**Frontend (User A)**:
```javascript
// 1. ส่งคำขอเป็นเพื่อน
async function sendFriendRequest(friendId) {
  await fetch('/api/friendships', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      friend_id: friendId
    })
  })

  // User B จะได้รับ notification ทาง WebSocket อัตโนมัติ
}
```

**Frontend (User B)**:
```javascript
// รับ notification คำขอเป็นเพื่อน
ws.addEventListener('message', (event) => {
  const message = JSON.parse(event.data)

  if (message.type === 'friend.request') {
    // แสดง notification
    showFriendRequestNotification(message.data)

    // เพิ่มใน pending requests list
    addPendingFriendRequest(message.data)
  }
})

// ยอมรับคำขอ
async function acceptFriendRequest(requestId) {
  await fetch(`/api/friendships/${requestId}/accept`, {
    method: 'PATCH',
    headers: {
      'Authorization': `Bearer ${token}`
    }
  })

  // User A จะได้รับ notification 'friend.accept' ทาง WebSocket
}
```

**Backend API**:
```go
// POST /api/friendships
func (h *FriendshipHandler) CreateFriendRequest(c *fiber.Ctx) error {
    userID := c.Locals("userID").(uuid.UUID)

    var req struct {
        FriendID uuid.UUID `json:"friend_id"`
    }
    c.BodyParser(&req)

    // สร้างคำขอ
    friendship, err := h.friendshipService.SendFriendRequest(userID, req.FriendID)
    if err != nil {
        return err
    }

    // ✅ Service จะเรียก NotifyFriendRequestReceived อัตโนมัติ

    return c.JSON(friendship)
}

// PATCH /api/friendships/:id/accept
func (h *FriendshipHandler) AcceptFriendRequest(c *fiber.Ctx) error {
    friendshipID, _ := uuid.Parse(c.Params("id"))

    // ยอมรับคำขอ
    friendship, err := h.friendshipService.AcceptFriendRequest(friendshipID)
    if err != nil {
        return err
    }

    // ✅ Service จะเรียก NotifyFriendRequestAccepted อัตโนมัติ

    return c.JSON(friendship)
}
```

---

### 3. สถานะออนไลน์/ออฟไลน์

**✅ ขั้นตอนที่ถูกต้อง**:

**Frontend**:
```javascript
// 1. เมื่อโหลดหน้าแชทหรือรายชื่อเพื่อน
// Subscribe สถานะของเพื่อนทุกคน
function subscribeToFriendsStatus(friends) {
  friends.forEach(friend => {
    ws.send(JSON.stringify({
      type: "user.status.subscribe",
      data: {
        user_id: friend.id
      }
    }))
  })
}

// 2. รับการอัปเดตสถานะ
ws.addEventListener('message', (event) => {
  const message = JSON.parse(event.data)

  switch(message.type) {
    case 'user.online':
      updateFriendStatus(message.data.user_id, 'online')
      break

    case 'user.offline':
      updateFriendStatus(message.data.user_id, 'offline')
      break
  }
})

// 3. Unsubscribe เมื่อออกจากหน้า
function unsubscribeFromFriendsStatus(friends) {
  friends.forEach(friend => {
    ws.send(JSON.stringify({
      type: "user.status.unsubscribe",
      data: {
        user_id: friend.id
      }
    }))
  })
}
```

**Backend**: ทำงานอัตโนมัติแล้ว
- เมื่อ user เชื่อมต่อ WebSocket → ส่ง `user.online` ไปยังผู้ที่ subscribe
- เมื่อ user ตัดการเชื่อมต่อ → ส่ง `user.offline` ไปยังผู้ที่ subscribe

---

### 4. สรุป Best Practices

| การทำงาน | ✅ ทำผ่าน | ❌ ไม่ควรทำผ่าน | เหตุผล |
|---------|-----------|-----------------|--------|
| ส่งข้อความ | REST API | WebSocket | ต้องบันทึก DB |
| แก้ไขข้อความ | REST API | WebSocket | ต้องอัปเดต DB |
| ลบข้อความ | REST API | WebSocket | ต้องลบใน DB |
| ส่งคำขอเป็นเพื่อน | REST API | WebSocket | ต้องบันทึก DB |
| ยอมรับคำขอ | REST API | WebSocket | ต้องอัปเดต DB |
| สร้าง conversation | REST API | WebSocket | ต้องบันทึก DB |
| รับ notification | WebSocket | ❌ | Real-time |
| Subscribe status | WebSocket | ❌ | Real-time |
| Typing indicator | WebSocket | ✅ | ไม่ต้องบันทึก |
| Ping/Pong | WebSocket | ✅ | Keep-alive |

---

## สรุป

### ✅ สิ่งที่ระบบทำได้ดี:
1. ส่ง notifications แบบ real-time
2. จัดการ connections หลายๆ คนพร้อมกัน
3. Subscribe/Unsubscribe conversations อัตโนมัติ
4. ติดตามสถานะออนไลน์/ออฟไลน์
5. Rate limiting

### ⚠️ สิ่งที่ต้องแก้ไข:
1. **อย่าใช้ WebSocket สำหรับส่งข้อความ** - ใช้ REST API แทน
2. **Frontend ต้อง listen events ทั้งหมด** - friend.request, friend.accept, conversation.create, message.receive
3. **Frontend ต้อง subscribe user status** - เพื่อดูสถานะออนไลน์/ออฟไลน์

### 🔧 การแก้ไขที่แนะนำ:

#### Backend (ถ้าต้องการ):
1. ปิดการใช้งาน `MessageSendHandler` หรือเปลี่ยนให้บันทึก DB
2. เพิ่ม logging เพิ่มเติม
3. เพิ่ม error handling

#### Frontend (จำเป็น):
1. **เปลี่ยนวิธีส่งข้อความ**: จาก WebSocket → REST API
2. **Listen events ทั้งหมด**: message.receive, friend.request, friend.accept, conversation.create, user.online, user.offline
3. **Subscribe user status**: เมื่อเปิดหน้าแชทหรือรายชื่อเพื่อน
4. **Join conversation**: ก่อนส่งข้อความ

---

## ตัวอย่าง Frontend Implementation (สมบูรณ์)

```javascript
class ChatWebSocket {
  constructor(token) {
    this.token = token
    this.ws = null
    this.handlers = {}
    this.reconnectAttempts = 0
    this.maxReconnectAttempts = 5
  }

  connect() {
    this.ws = new WebSocket(`ws://localhost:8080/ws/user?token=${this.token}`)

    this.ws.onopen = () => {
      console.log('WebSocket connected')
      this.reconnectAttempts = 0
      this.onConnect()
    }

    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data)
      this.handleMessage(message)
    }

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error)
    }

    this.ws.onclose = () => {
      console.log('WebSocket closed')
      this.reconnect()
    }
  }

  reconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++
      setTimeout(() => {
        console.log(`Reconnecting... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`)
        this.connect()
      }, 1000 * this.reconnectAttempts)
    }
  }

  onConnect() {
    // เรียก callback เมื่อเชื่อมต่อสำเร็จ
    if (this.handlers.onConnect) {
      this.handlers.onConnect()
    }
  }

  handleMessage(message) {
    switch(message.type) {
      case 'connect':
        console.log('Connected:', message.data)
        break

      case 'conversation.list':
        if (this.handlers.onConversationList) {
          this.handlers.onConversationList(message.data)
        }
        break

      case 'message.receive':
        if (this.handlers.onNewMessage) {
          this.handlers.onNewMessage(message.data)
        }
        break

      case 'conversation.create':
        if (this.handlers.onConversationCreated) {
          this.handlers.onConversationCreated(message.data)
        }
        break

      case 'friend.request':
        if (this.handlers.onFriendRequest) {
          this.handlers.onFriendRequest(message.data)
        }
        break

      case 'friend.accept':
        if (this.handlers.onFriendAccepted) {
          this.handlers.onFriendAccepted(message.data)
        }
        break

      case 'user.online':
        if (this.handlers.onUserOnline) {
          this.handlers.onUserOnline(message.data.user_id)
        }
        break

      case 'user.offline':
        if (this.handlers.onUserOffline) {
          this.handlers.onUserOffline(message.data.user_id)
        }
        break

      case 'error':
        console.error('WebSocket error:', message.error)
        break
    }
  }

  // API methods
  joinConversation(conversationId) {
    this.send({
      type: 'conversation.join',
      data: { conversation_id: conversationId }
    })
  }

  leaveConversation(conversationId) {
    this.send({
      type: 'conversation.leave',
      data: { conversation_id: conversationId }
    })
  }

  subscribeUserStatus(userId) {
    this.send({
      type: 'user.status.subscribe',
      data: { user_id: userId }
    })
  }

  unsubscribeUserStatus(userId) {
    this.send({
      type: 'user.status.unsubscribe',
      data: { user_id: userId }
    })
  }

  setTyping(conversationId, isTyping) {
    this.send({
      type: 'message.typing',
      data: {
        conversation_id: conversationId,
        is_typing: isTyping
      }
    })
  }

  send(message) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      message.timestamp = new Date().toISOString()
      this.ws.send(JSON.stringify(message))
    }
  }

  on(eventName, callback) {
    this.handlers[eventName] = callback
  }

  disconnect() {
    if (this.ws) {
      this.ws.close()
    }
  }
}

// การใช้งาน
const chatWS = new ChatWebSocket(jwtToken)

// ตั้งค่า handlers
chatWS.on('onConversationList', (conversations) => {
  console.log('Conversations:', conversations)
  renderConversations(conversations)
})

chatWS.on('onNewMessage', (message) => {
  console.log('New message:', message)
  appendMessageToChat(message)
})

chatWS.on('onFriendRequest', (request) => {
  console.log('Friend request:', request)
  showFriendRequestNotification(request)
})

chatWS.on('onUserOnline', (userId) => {
  updateUserStatus(userId, true)
})

chatWS.on('onUserOffline', (userId) => {
  updateUserStatus(userId, false)
})

// เชื่อมต่อ
chatWS.connect()

// ส่งข้อความ (ผ่าน REST API)
async function sendMessage(conversationId, content) {
  const response = await fetch(`/api/conversations/${conversationId}/messages`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${jwtToken}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      content: content,
      message_type: 'text'
    })
  })

  // ข้อความจะถูกส่งกลับมาทาง WebSocket event 'message.receive'
}
```

---

## เอกสารอ้างอิง

- **Hub**: `interfaces/websocket/hub.go`
- **Client**: `interfaces/websocket/client.go`
- **Handlers**: `interfaces/websocket/handlers.go`
- **Routes**: `interfaces/websocket/routes.go`
- **Broadcast**: `interfaces/websocket/broadcast.go`
- **Notification Service**: `application/serviceimpl/notification_service.go`
- **WebSocket Adapter**: `infrastructure/adapter/websocket_adapter.go`
- **Notification Interface**: `domain/service/notification_service.go`

---

**หมายเหตุ**: เอกสารนี้วิเคราะห์จากโค้ดปัจจุบัน หากมีการแก้ไขโค้ด ควรอัปเดตเอกสารนี้ด้วย
