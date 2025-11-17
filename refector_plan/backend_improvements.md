# Backend Improvements - สถานะปัจจุบัน

**วันที่อัพเดท**: 2025-11-12
**สถานะ**: ✅ **แก้ไขเสร็จสมบูรณ์แล้ว**

---

## ✅ สรุปการแก้ไขที่เสร็จสิ้น

จากการตรวจสอบ Backend พบว่า **API มีครบแล้ว** แต่ **ขาดการส่ง WebSocket Notifications** หลายจุด

**ผลการแก้ไข**: แก้ไขครบทั้งหมด **9 notifications** แบ่งเป็น:
- ✅ Message Notifications: 1 รายการ
- ✅ Friend Notifications: 4 รายการ
- ✅ Group Chat Notifications: 4 รายการ

---

## 1. ✅ Message Notifications (เสร็จสิ้น)

### ✅ DeleteMessage - เพิ่ม Notification แล้ว

**ไฟล์ที่แก้ไข**:
1. `application/serviceimpl/message_service.go` - เพิ่ม `notificationService` field
2. `application/serviceimpl/message_delete_service.go` - เรียก notification หลังลบข้อความ
3. `pkg/di/container.go` - อัพเดท DI container

**โค้ดที่เพิ่ม** (`message_delete_service.go` line 122-125):
```go
// ส่ง WebSocket notification แจ้งว่าข้อความถูกลบ
if s.notificationService != nil {
    s.notificationService.NotifyMessageDeleted(message.ConversationID, messageID)
}
```

**Frontend Event**: `message.delete`

**Payload**:
```json
{
  "message_id": "uuid",
  "deleted_at": "2025-11-12T10:30:00Z"
}
```

---

## 2. ✅ Friend Notifications (เสร็จสิ้น)

### ✅ NotifyFriendRequestRejected - สร้างครบแล้ว

**ไฟล์ที่แก้ไข**:
1. `domain/service/notification_service.go` - เพิ่ม interface method
2. `application/serviceimpl/notification_service.go` - implement เต็มรูปแบบ
3. `domain/port/websocket_port.go` - เพิ่ม interface method
4. `infrastructure/adapter/websocket_adapter.go` - implement
5. `interfaces/api/handler/user_friendship_handler.go` - เรียกใช้ใน handler

**Frontend Event**: `friend.reject`

**Payload**:
```json
{
  "friendship_id": "uuid",
  "user_id": "uuid",
  "friend_id": "uuid",
  "status": "rejected",
  "rejected_at": "2025-11-12T10:30:00Z",
  "rejector": {
    "id": "uuid",
    "username": "user123",
    "display_name": "User Name",
    "profile_image_url": "https://..."
  }
}
```

---

### ✅ NotifyFriendRemoved - เพิ่มการเรียกใช้แล้ว

**ไฟล์ที่แก้ไข**:
- `interfaces/api/handler/user_friendship_handler.go` - เพิ่มบรรทัด 403-404

**โค้ดที่เพิ่ม**:
```go
// ส่ง WebSocket notification แจ้งทั้งสองฝ่าย
h.notificationService.NotifyFriendRemoved(userID, friendID)
```

**Frontend Event**: `friend.remove`

**Payload**:
```json
{
  "friend_id": "uuid",
  "removed_by": "uuid",
  "timestamp": "2025-11-12T10:30:00Z"
}
```

---

### ✅ NotifyUserBlocked - เพิ่มการเรียกใช้แล้ว

**ไฟล์ที่แก้ไข**:
- `interfaces/api/handler/user_friendship_handler.go` - เพิ่มบรรทัด 438-439

**โค้ดที่เพิ่ม**:
```go
// ส่ง WebSocket notification แจ้งผู้ถูกบล็อก
h.notificationService.NotifyUserBlocked(userID, targetID)
```

**Frontend Event**: `user.blocked`

**Payload**:
```json
{
  "blocker_id": "uuid",
  "blocked_at": "2025-11-12T10:30:00Z"
}
```

---

### ✅ NotifyUserUnblocked - เพิ่มการเรียกใช้แล้ว

**ไฟล์ที่แก้ไข**:
- `interfaces/api/handler/user_friendship_handler.go` - เพิ่มบรรทัด 473-474

**โค้ดที่เพิ่ม**:
```go
// ส่ง WebSocket notification แจ้งผู้ถูกปลดบล็อก
h.notificationService.NotifyUserUnblocked(userID, targetID)
```

**Frontend Event**: `user.unblocked`

**Payload**:
```json
{
  "unblocker_id": "uuid",
  "unblocked_at": "2025-11-12T10:30:00Z"
}
```

---

## 3. ✅ Group Chat Notifications (เสร็จสิ้น - เพิ่งค้นพบและแก้ไข)

### ✅ CreateGroupConversation - เพิ่ม Notification แล้ว

**ปัญหาเดิม**: เวลาสร้างกลุ่มใหม่ สมาชิกที่ถูกเชิญไม่ได้รับ notification real-time

**ไฟล์ที่แก้ไข**:
- `interfaces/api/handler/conversation_handler.go` - เพิ่มบรรทัด 172-180

**โค้ดที่เพิ่ม**:
```go
// รวมรายการผู้ใช้ทั้งหมดที่เกี่ยวข้อง (creator + members)
allMembers := append([]uuid.UUID{userID}, memberIDs...)

// ส่ง WebSocket notification แจ้งสมาชิกทุกคนในกลุ่ม
err = h.notificationService.NotifyConversationCreated(allMembers, conversation)
if err != nil {
    log.Printf("Failed to send group conversation created notification: %v", err)
}
```

**Frontend Event**: `conversation.create`

**Payload**: (conversation DTO เต็มรูปแบบ)
```json
{
  "id": "uuid",
  "type": "group",
  "title": "ชื่อกลุ่ม",
  "icon_url": "https://...",
  "members": [...],
  "created_at": "2025-11-12T10:30:00Z"
}
```

---

### ✅ UpdateConversation - เพิ่ม Notification แล้ว

**ปัญหาเดิม**: เวลาแก้ไขชื่อกลุ่มหรือรูปกลุ่ม สมาชิกคนอื่นไม่เห็นการเปลี่ยนแปลง real-time

**ไฟล์ที่แก้ไข**:
- `interfaces/api/handler/conversation_handler.go` - เพิ่มบรรทัด 671-681

**โค้ดที่เพิ่ม**:
```go
// ส่ง WebSocket notification แจ้งสมาชิกทุกคนในกลุ่ม
notificationData := types.JSONB{
    "conversation_id": conversationID.String(),
}
if title, ok := updateData["title"]; ok {
    notificationData["title"] = title
}
if iconURL, ok := updateData["icon_url"]; ok {
    notificationData["icon_url"] = iconURL
}
h.notificationService.NotifyConversationUpdated(conversationID, notificationData)
```

**Frontend Event**: `conversation.update`

**Payload**:
```json
{
  "conversation_id": "uuid",
  "title": "ชื่อกลุ่มใหม่",
  "icon_url": "https://..."
}
```

---

### ✅ AddConversationMember - เพิ่ม Notification แล้ว

**ปัญหาเดิม**: เวลาเชิญเพื่อนเข้ากลุ่ม เพื่อนที่ถูกเชิญและสมาชิกในกลุ่มไม่ได้รับ notification

**ไฟล์ที่แก้ไข**:
1. `interfaces/api/handler/conversation_member_handler.go` - เพิ่ม `notificationService` field
2. `interfaces/api/handler/conversation_member_handler.go` - เพิ่มการเรียก notification (บรรทัด 103-104)
3. `pkg/di/container.go` - อัพเดท DI container

**โค้ดที่เพิ่ม**:
```go
// ส่ง WebSocket notification แจ้งว่ามีสมาชิกใหม่ถูกเพิ่มเข้ากลุ่ม
h.notificationService.NotifyUserAddedToConversation(conversationID, newMemberID)
```

**Frontend Event**:
- `conversation.user_added` - ส่งให้สมาชิกเดิมในกลุ่ม
- `conversation.create` - ส่งให้สมาชิกใหม่

**Payload** (user_added):
```json
{
  "conversation_id": "uuid",
  "user_id": "uuid",
  "added_at": "2025-11-12T10:30:00Z"
}
```

**Payload** (สำหรับสมาชิกใหม่):
```json
{
  "conversation_id": "uuid",
  "message": "คุณถูกเพิ่มในบทสนทนา"
}
```

---

### ✅ RemoveConversationMember - เพิ่ม Notification แล้ว

**ปัญหาเดิม**: เวลาเตะเพื่อนออกจากกลุ่ม เพื่อนที่ถูกเตะไม่รู้ว่าถูกเตะออก

**ไฟล์ที่แก้ไข**:
- `interfaces/api/handler/conversation_member_handler.go` - เพิ่มบรรทัด 240-241

**โค้ดที่เพิ่ม**:
```go
// ส่ง WebSocket notification แจ้งว่าสมาชิกถูกลบออกจากกลุ่ม
h.notificationService.NotifyUserRemovedFromConversation(targetUserID, conversationID)
```

**Frontend Event**: `conversation.user_removed`

**Payload**:
```json
{
  "conversation_id": "uuid",
  "removed_at": "2025-11-12T10:30:00Z"
}
```

---

## 4. สรุปการเปลี่ยนแปลงทั้งหมด

### ไฟล์ที่ถูกแก้ไข (รวม 8 ไฟล์):

1. ✅ `application/serviceimpl/message_service.go` - เพิ่ม notificationService dependency
2. ✅ `application/serviceimpl/message_delete_service.go` - เพิ่ม notification call
3. ✅ `domain/service/notification_service.go` - เพิ่ม NotifyFriendRequestRejected interface
4. ✅ `application/serviceimpl/notification_service.go` - implement NotifyFriendRequestRejected
5. ✅ `domain/port/websocket_port.go` - เพิ่ม BroadcastFriendRequestRejected interface
6. ✅ `infrastructure/adapter/websocket_adapter.go` - implement BroadcastFriendRequestRejected
7. ✅ `interfaces/api/handler/user_friendship_handler.go` - เพิ่ม notifications 4 จุด
8. ✅ `interfaces/api/handler/conversation_handler.go` - เพิ่ม notifications 2 จุด
9. ✅ `interfaces/api/handler/conversation_member_handler.go` - เพิ่ม notificationService + notifications 2 จุด
10. ✅ `pkg/di/container.go` - อัพเดท DI 2 จุด

### Compilation Status:
✅ **ผ่านการ compile สำเร็จ** (go build ./cmd/api)

---

## 5. Frontend: WebSocket Events ที่ต้องรองรับ

Frontend ต้อง **เพิ่มการ listen events ใหม่** ดังนี้:

```javascript
ws.addEventListener('message', (event) => {
  const message = JSON.parse(event.data)

  switch(message.type) {
    // ============= Message Events =============
    case 'message.receive':      // ✅ มีอยู่แล้ว
    case 'message.edit':         // ✅ มีอยู่แล้ว
    case 'message.read':         // ✅ มีอยู่แล้ว
    case 'message.read_all':     // ✅ มีอยู่แล้ว
    case 'message.delete':       // ⚠️ ต้องเพิ่มใหม่
      handleMessageDeleted(message.data)
      break

    // ============= Friend Events =============
    case 'friend.request':       // ✅ มีอยู่แล้ว
    case 'friend.accept':        // ✅ มีอยู่แล้ว
    case 'friend.reject':        // ⚠️ ต้องเพิ่มใหม่
      handleFriendRequestRejected(message.data)
      break

    case 'friend.remove':        // ⚠️ ต้องเพิ่มใหม่
      handleFriendRemoved(message.data)
      break

    // ============= User Events =============
    case 'user.blocked':         // ⚠️ ต้องเพิ่มใหม่
      handleUserBlocked(message.data)
      break

    case 'user.unblocked':       // ⚠️ ต้องเพิ่มใหม่
      handleUserUnblocked(message.data)
      break

    // ============= Group Chat Events =============
    case 'conversation.create':  // ✅ มีอยู่แล้ว (แต่ต้องเช็คว่า handle group ด้วยไหม)
      handleNewConversation(message.data)
      break

    case 'conversation.update':  // ⚠️ ต้องเพิ่มใหม่ หรือตรวจสอบว่ามีแล้วหรือยัง
      handleConversationUpdated(message.data)
      break

    case 'conversation.user_added':    // ⚠️ ต้องเพิ่มใหม่
      handleUserAddedToConversation(message.data)
      break

    case 'conversation.user_removed':  // ⚠️ ต้องเพิ่มใหม่
      handleUserRemovedFromConversation(message.data)
      break
  }
})
```

---

## 6. ตัวอย่าง Handler Functions สำหรับ Frontend

### Message Handlers

```javascript
function handleMessageDeleted(data) {
  // data: { message_id: "uuid", deleted_at: "timestamp" }

  // 1. หา message element ใน DOM
  const messageElement = document.querySelector(`[data-message-id="${data.message_id}"]`)

  // 2. แสดง UI ว่าข้อความถูกลบ (อาจจะแสดง "Message deleted" หรือลบออกจาก UI)
  if (messageElement) {
    messageElement.classList.add('deleted')
    messageElement.innerHTML = '<em>This message has been deleted</em>'
  }

  // 3. อัพเดท state/store ถ้ามี
  // store.dispatch('deleteMessage', data.message_id)
}
```

### Friend Handlers

```javascript
function handleFriendRequestRejected(data) {
  // data: { friendship_id, user_id, friend_id, status, rejected_at, rejector: {...} }

  // 1. แสดงการแจ้งเตือน
  showNotification(
    `${data.rejector.display_name} rejected your friend request`,
    'warning'
  )

  // 2. อัพเดท UI (ลบจาก pending requests)
  removePendingRequest(data.friendship_id)

  // 3. อัพเดท state
  // store.dispatch('friendRequestRejected', data)
}

function handleFriendRemoved(data) {
  // data: { friend_id, removed_by, timestamp }

  // 1. แสดงการแจ้งเตือน
  showNotification('A friend has been removed', 'info')

  // 2. ลบเพื่อนออกจาก friend list
  removeFriendFromList(data.friend_id)

  // 3. อัพเดท conversation list (ถ้ามี direct conversation)
  // updateConversationStatus(conversationId, 'inactive')
}
```

### User Block Handlers

```javascript
function handleUserBlocked(data) {
  // data: { blocker_id, blocked_at }

  // 1. แสดงการแจ้งเตือน
  showNotification('You have been blocked by a user', 'error')

  // 2. ลบการสนทนากับผู้ใช้นั้นออก หรือทำให้ส่งข้อความไม่ได้
  // disableConversationWith(data.blocker_id)
}

function handleUserUnblocked(data) {
  // data: { unblocker_id, unblocked_at }

  // 1. แสดงการแจ้งเตือน
  showNotification('You have been unblocked', 'success')

  // 2. เปิดใช้งานการสนทนาอีกครั้ง (ถ้ามี)
  // enableConversationWith(data.unblocker_id)
}
```

### Group Chat Handlers

```javascript
function handleConversationUpdated(data) {
  // data: { conversation_id, title?, icon_url? }

  // 1. อัพเดท conversation ใน list
  updateConversationInList(data.conversation_id, {
    title: data.title,
    icon_url: data.icon_url
  })

  // 2. ถ้ากำลังเปิดการสนทนานี้อยู่ ให้อัพเดท header ด้วย
  if (currentConversationId === data.conversation_id) {
    updateConversationHeader({
      title: data.title,
      iconUrl: data.icon_url
    })
  }
}

function handleUserAddedToConversation(data) {
  // data: { conversation_id, user_id, added_at }

  // 1. ถ้ากำลังเปิดการสนทนานี้อยู่ ให้โหลด member list ใหม่
  if (currentConversationId === data.conversation_id) {
    refreshMemberList(data.conversation_id)
  }

  // 2. แสดง system message ว่ามีคนถูกเพิ่มเข้ามา
  // addSystemMessage(`User ${data.user_id} joined the group`)
}

function handleUserRemovedFromConversation(data) {
  // data: { conversation_id, removed_at }

  // 1. ลบการสนทนานี้ออกจาก list
  removeConversationFromList(data.conversation_id)

  // 2. ถ้ากำลังเปิดการสนทนานี้อยู่ ให้ปิดและแสดงข้อความ
  if (currentConversationId === data.conversation_id) {
    closeConversation()
    showNotification('You have been removed from this group', 'warning')
  }
}
```

---

## 7. การทดสอบที่แนะนำ

### 7.1 Message Tests

- [ ] ลบข้อความ → ฝั่งอื่นควรเห็นข้อความหายไป real-time
- [ ] แก้ไขข้อความ → ฝั่งอื่นควรเห็นข้อความที่แก้แล้ว

### 7.2 Friend Tests

- [ ] ส่งคำขอเป็นเพื่อน → อีกฝ่ายได้รับคำขอ real-time
- [ ] ยอมรับคำขอ → ผู้ส่งคำขอได้รับการแจ้งเตือน
- [ ] ปฏิเสธคำขอ → ผู้ส่งคำขอได้รับการแจ้งเตือน
- [ ] ลบเพื่อน → ทั้งสองฝ่ายได้รับการแจ้งเตือน
- [ ] บล็อกผู้ใช้ → ผู้ถูกบล็อกได้รับการแจ้งเตือน
- [ ] ปลดบล็อก → ผู้ถูกปลดบล็อกได้รับการแจ้งเตือน

### 7.3 Group Chat Tests

- [ ] สร้างกลุ่มใหม่ → สมาชิกทุกคนได้รับการแจ้งเตือน
- [ ] แก้ไขชื่อกลุ่ม → สมาชิกทุกคนเห็นชื่อใหม่ real-time
- [ ] แก้ไขรูปกลุ่ม → สมาชิกทุกคนเห็นรูปใหม่ real-time
- [ ] เชิญสมาชิกใหม่ → สมาชิกใหม่และสมาชิกเดิมได้รับการแจ้งเตือน
- [ ] เตะสมาชิกออก → คนที่ถูกเตะได้รับการแจ้งเตือนและถูกลบออกจากกลุ่ม

---

## 8. สรุป

### ✅ สิ่งที่แก้ไขเสร็จแล้ว (100%)

| Category | Feature | Status | Frontend Event |
|----------|---------|--------|----------------|
| **Messages** | DeleteMessage | ✅ แก้แล้ว | `message.delete` |
| **Friends** | RejectFriendRequest | ✅ แก้แล้ว (สร้างใหม่ทั้งหมด) | `friend.reject` |
| **Friends** | RemoveFriend | ✅ แก้แล้ว | `friend.remove` |
| **Friends** | BlockUser | ✅ แก้แล้ว | `user.blocked` |
| **Friends** | UnblockUser | ✅ แก้แล้ว | `user.unblocked` |
| **Group Chat** | CreateGroupConversation | ✅ แก้แล้ว | `conversation.create` |
| **Group Chat** | UpdateConversation | ✅ แก้แล้ว | `conversation.update` |
| **Group Chat** | AddConversationMember | ✅ แก้แล้ว | `conversation.user_added` |
| **Group Chat** | RemoveConversationMember | ✅ แก้แล้ว | `conversation.user_removed` |

### 📊 สถิติการแก้ไข

- **จำนวน Notifications ที่เพิ่ม**: 9
- **จำนวนไฟล์ที่แก้**: 10
- **Compilation Status**: ✅ Pass
- **เวลาที่ใช้**: ~2 ชั่วโมง

### 📝 สิ่งที่ Frontend ต้องทำต่อ

1. ✅ เพิ่ม Event Listeners ใหม่ 8 events (ดูรายละเอียดใน Section 5)
2. ✅ Implement Handler Functions สำหรับแต่ละ event (มีตัวอย่างใน Section 6)
3. ✅ ทดสอบทุก scenario (ดูรายการใน Section 7)
4. ✅ อัพเดท UI/UX ให้รองรับการแจ้งเตือนแบบ real-time

---

**หมายเหตุ**: Backend พร้อมใช้งานแล้ว ✅ Frontend สามารถเริ่มพัฒนาการรับ WebSocket events ได้เลย
