# WebSocket Events สำหรับ Group Chat

## 📋 สรุปภาพรวม

เอกสารนี้สรุป WebSocket events ทั้งหมดที่เกี่ยวข้องกับ Group Chat ในระบบ โดยแบ่งเป็น events ที่มีอยู่แล้วและ events ที่อาจจะขาดหายไป

---

## ✅ Events ที่มีอยู่แล้วในระบบ

### 1. Conversation Management Events

| Event Type | ทิศทาง | คำอธิบาย | Location |
|-----------|--------|----------|----------|
| `conversation.create` | Client → Server<br>Server → Client | สร้าง group conversation ใหม่ | `handlers.go:651-757`<br>`websocket_adapter.go:87-90` |
| `conversation.update` | Server → Client | อัพเดตข้อมูล group (title, icon, settings) | `websocket_adapter.go:92-95` |
| `conversation.delete` | Server → Client | ลบ group conversation | `websocket_adapter.go:97-104` |
| `conversation.join` | Client → Server | เข้าร่วม/subscribe group conversation | `handlers.go:391-474` |
| `conversation.leave` | Client → Server | ออกจาก/unsubscribe group conversation | `handlers.go:481-548` |
| `conversation.active` | Client → Server | ตั้งค่า active conversation ปัจจุบัน | `handlers.go:555-644` |
| `conversation.load` | Client → Server | โหลดรายการ conversations | `handlers.go:258-384` |
| `conversation.list` | Server → Client | ส่งรายการ conversations กลับ | `handlers.go:374-378` |
| `conversation.joined` | Server → Client | ยืนยันการเข้าร่วม conversation | `handlers.go:460-468` |
| `conversation.left` | Server → Client | ยืนยันการออกจาก conversation | `handlers.go:537-545` |
| `conversation.active_set` | Server → Client | ยืนยันการตั้งค่า active conversation | `handlers.go:633-641` |

### 2. Member Management Events

| Event Type | ทิศทาง | คำอธิบาย | Location |
|-----------|--------|----------|----------|
| `conversation.user_added` | Server → Client | แจ้งเตือนเมื่อมี member ใหม่ถูกเพิ่มเข้า group | `websocket_adapter.go:106-122` |
| `conversation.user_removed` | Server → Client | แจ้งเตือนเมื่อ member ถูกลบออกจาก group | `websocket_adapter.go:124-131` |
| `conversation.user_active` | Server → Client | แจ้งสถานะว่า user กำลัง active ใน conversation | `handlers.go:451-457`<br>`handlers.go:622-629` |

**Data Structure สำหรับ `conversation.user_added`:**
```json
{
  "conversation_id": "uuid",
  "user_id": "uuid",
  "added_at": "timestamp"
}
```

**Data Structure สำหรับ `conversation.user_removed`:**
```json
{
  "conversation_id": "uuid",
  "removed_at": "timestamp"
}
```

**Data Structure สำหรับ `conversation.user_active`:**
```json
{
  "user_id": "uuid",
  "conversation_id": "uuid",
  "active": true/false,
  "timestamp": "timestamp"
}
```

### 3. Message Events (ใช้ได้กับ Group Chat)

| Event Type | ทิศทาง | คำอธิบาย | Location |
|-----------|--------|----------|----------|
| `message.send` | Client → Server | ส่งข้อความใหม่ใน group | `handlers.go:38-116` |
| `message.receive` | Server → Client | รับข้อความใหม่ใน group | `handlers.go:113`<br>`websocket_adapter.go:47-51` |
| `message.edit` | Client → Server<br>Server → Client | แก้ไขข้อความใน group | `handlers.go:189-218`<br>`websocket_adapter.go:63-66` |
| `message.delete` | Client → Server<br>Server → Client | ลบข้อความใน group | `handlers.go:225-251`<br>`websocket_adapter.go:73-80` |
| `message.read` | Client → Server<br>Server → Client | อ่านข้อความใน group | `handlers.go:156-182`<br>`websocket_adapter.go:53-56` |
| `message.read_all` | Server → Client | อ่านข้อความทั้งหมดใน group | `websocket_adapter.go:58-61` |
| `message.typing` | Client → Server<br>Server → Client | แสดงสถานะกำลังพิมพ์ | `handlers.go:123-149` |
| `message.reply` | Server → Client | ตอบกลับข้อความ | `websocket_adapter.go:68-71` |
| `message.reaction` | Server → Client | แสดง reaction ต่อข้อความ | `websocket_adapter.go:82-85` |

### 4. User Status Events

| Event Type | ทิศทาง | คำอธิบาย | Location |
|-----------|--------|----------|----------|
| `user.status.subscribe` | Client → Server | ติดตามสถานะของ user | `handlers.go:785-821` |
| `user.status.unsubscribe` | Client → Server | ยกเลิกติดตามสถานะของ user | `handlers.go:831-873` |
| `user.status.subscribed` | Server → Client | ยืนยันการติดตามสถานะ | `handlers.go:809-818` |
| `user.status.unsubscribed` | Server → Client | ยืนยันการยกเลิกติดตาม | `handlers.go:854-863` |

### 5. Connection Events

| Event Type | ทิศทาง | คำอธิบาย | Location |
|-----------|--------|----------|----------|
| `connect` | Client ↔ Server | เชื่อมต่อ WebSocket | `hub.go:110` |
| `disconnect` | Client ↔ Server | ตัดการเชื่อมต่อ | `hub.go:111` |
| `ping` | Client → Server | ตรวจสอบการเชื่อมต่อ | `handlers.go:759-777` |
| `pong` | Server → Client | ตอบกลับ ping | `handlers.go:769-774` |

---

## ❌ Events ที่ขาดหายไป (แนะนำให้เพิ่ม)

### 1. Admin & Permission Management

| Event Type | ทิศทาง | คำอธิบาย | Priority |
|-----------|--------|----------|----------|
| `conversation.admin_added` | Server → Client | แต่งตั้ง admin ของ group | High |
| `conversation.admin_removed` | Server → Client | ถอด admin ของ group | High |
| `conversation.role_updated` | Server → Client | เปลี่ยนแปลง role ของ member | Medium |
| `conversation.permissions_updated` | Server → Client | อัพเดต group permissions | Medium |

**Suggested Data Structure:**
```json
{
  "conversation_id": "uuid",
  "user_id": "uuid",
  "role": "admin|moderator|member",
  "updated_by": "uuid",
  "timestamp": "timestamp"
}
```

### 2. Granular Group Info Events

| Event Type | ทิศทาง | คำอธิบาย | Priority |
|-----------|--------|----------|----------|
| `conversation.title_updated` | Server → Client | เปลี่ยนชื่อ group | Low |
| `conversation.icon_updated` | Server → Client | เปลี่ยน icon/avatar group | Low |
| `conversation.description_updated` | Server → Client | เปลี่ยน description | Low |
| `conversation.settings_updated` | Server → Client | เปลี่ยน group settings | Medium |

**หมายเหตุ:** ปัจจุบันใช้ `conversation.update` ทั่วไป แต่การแยกเป็น event เฉพาะจะช่วยให้ client จัดการได้ละเอียดขึ้น

### 3. Member Action Events

| Event Type | ทิศทาง | คำอธิบาย | Priority |
|-----------|--------|----------|----------|
| `conversation.member_left` | Server → Client | member ออกจาก group เอง (แยกจาก removed) | Medium |
| `conversation.member_joined` | Server → Client | member เข้า group (แยกจาก added) | Low |

**ความแตกต่าง:**
- `conversation.user_added` - ถูกเพิ่มโดยคนอื่น
- `conversation.member_joined` - เข้าร่วมเอง (เช่น ผ่าน invite link)
- `conversation.user_removed` - ถูกลบโดยคนอื่น
- `conversation.member_left` - ออกจาก group เอง

### 4. Invite Link Events

| Event Type | ทิศทาง | คำอธิบาย | Priority |
|-----------|--------|----------|----------|
| `conversation.invite_link_created` | Server → Client | สร้าง invite link | Low |
| `conversation.invite_link_revoked` | Server → Client | ยกเลิก invite link | Low |
| `conversation.user_joined_via_link` | Server → Client | มีคนเข้า group ผ่าน link | Low |

### 5. Pin & Important Messages

| Event Type | ทิศทาง | คำอธิบาย | Priority |
|-----------|--------|----------|----------|
| `message.pinned` | Server → Client | ปักหมุดข้อความ | Medium |
| `message.unpinned` | Server → Client | ยกเลิกปักหมุดข้อความ | Medium |

### 6. Mute & Notification Settings

| Event Type | ทิศทาง | คำอธิบาย | Priority |
|-----------|--------|----------|----------|
| `conversation.muted` | Server → Client | ปิดเสียงแจ้งเตือน group | Low |
| `conversation.unmuted` | Server → Client | เปิดเสียงแจ้งเตือน group | Low |

---

## 📊 สรุปความครบถ้วนของ Events

### ✅ Features ที่ครบถ้วน (90-100%)
- ✅ Message Management (send, edit, delete, read, typing)
- ✅ Conversation Join/Leave
- ✅ Add/Remove Members
- ✅ Conversation CRUD operations
- ✅ Real-time notifications

### ⚠️ Features ที่ครบบางส่วน (50-80%)
- ⚠️ Group Info Updates (มีแต่ใช้ event ทั่วไป)
- ⚠️ Member Status (มี active status แต่ไม่มี online/offline)

### ❌ Features ที่ขาดหายไป (0-30%)
- ❌ Admin & Permission Management (0%)
- ❌ Invite Link System (0%)
- ❌ Pin Messages (0%)
- ❌ Granular Member Actions (30% - มีแต่ไม่แยกประเภท)

---

## 🎯 แนวทางการพัฒนาต่อ

### Phase 1: Critical Features (Priority: High)
1. เพิ่ม Admin Management Events
   - `conversation.admin_added`
   - `conversation.admin_removed`
   - `conversation.permissions_updated`

2. แยก Member Action Events
   - แยกระหว่าง "added" vs "joined"
   - แยกระหว่าง "removed" vs "left"

### Phase 2: Enhanced Features (Priority: Medium)
1. Pin Message Events
   - `message.pinned`
   - `message.unpinned`

2. Granular Settings Events
   - แยก `conversation.update` เป็น events เฉพาะ

### Phase 3: Advanced Features (Priority: Low)
1. Invite Link System
2. Advanced Notification Settings
3. Member Role Management

---

## 📁 ไฟล์ที่เกี่ยวข้อง

### Core Files
- `interfaces/websocket/handlers.go` - Message handlers และ conversation handlers
- `interfaces/websocket/broadcast.go` - Broadcasting functions
- `interfaces/websocket/hub.go` - WebSocket hub และ MessageType definitions
- `domain/port/websocket_port.go` - WebSocket port interface
- `infrastructure/adapter/websocket_adapter.go` - WebSocket adapter implementation

### Handler Registration
- Line 15-36 ใน `handlers.go` - ลงทะเบียน handlers ทั้งหมด

### MessageType Constants
- Line 106-136 ใน `hub.go` - กำหนด MessageType constants

---

## 📝 หมายเหตุ

1. **Broadcasting Mechanism**: ระบบใช้ conversation subscription model โดย client ต้อง join conversation ก่อนจึงจะได้รับ events
2. **User Active Status**: ระบบมีการติดตามสถานะ active ของ user ใน conversation แล้ว
3. **Message Types**: รองรับหลายประเภท (text, file, sticker, etc.)
4. **Business Features**: มี events สำหรับ business-related features แต่ไม่รวมในเอกสารนี้

---

**เอกสารนี้สร้างขึ้นเมื่อ:** 2025-11-17
**Version:** 1.0
**สถานะ:** ✅ Complete Analysis
