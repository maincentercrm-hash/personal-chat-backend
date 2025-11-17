# 🔍 การตรวจสอบปัญหา Remove Member - Backend Analysis

**วันที่:** 2025-11-17
**ปัญหา:** คนที่ถูก remove จากกลุ่มยังเห็น conversation อยู่
**สถานะ:** ✅ Backend ทำงานถูกต้องแล้ว

---

## 📊 สรุปผลการตรวจสอบ

### ✅ 1. API GET /conversations - **ผ่าน**

**Location:** `infrastructure/persistence/postgres/conversation_repository.go:722-740`

**การทำงาน:**
```go
func (r *conversationRepository) GetUserConversationsWithFilter(userID uuid.UUID, ...) {
    // ดึงรายการการสนทนาที่ผู้ใช้เป็นสมาชิก
    var memberIDs []uuid.UUID
    err := r.db.Model(&models.ConversationMember{}).
        Select("conversation_id").
        Where("user_id = ? AND is_hidden = ?", userID, false).
        Find(&memberIDs).Error

    // จากนั้นดึงเฉพาะ conversation ที่อยู่ใน memberIDs
    baseQuery := r.db.Model(&models.Conversation{}).
        Where("conversations.id IN (?) AND conversations.is_active = ?", memberIDs, true)
}
```

**ผลการทดสอบ:**
- ✅ กรองเฉพาะ conversations ที่ user **เป็นสมาชิก**อยู่
- ✅ เมื่อ member ถูก **DELETE** จาก `conversation_members` table จะ**ไม่**ปรากฏใน memberIDs
- ✅ Conversation **จะหายไป**จาก API response ทันที

**สรุป:** Backend กรองถูกต้องแล้ว ❌ ไม่ใช่ปัญหา

---

### ✅ 2. API DELETE Remove Member - **ผ่าน**

**Location:** `interfaces/api/handler/conversation_member_handler.go:285-358`

**Flow การทำงาน:**

#### 2.1 Handler (Line 285-358)
```go
func (h *ConversationMemberHandler) RemoveConversationMember(c *fiber.Ctx) error {
    // 1. ตรวจสอบ permissions
    // 2. ดึง targetUserID ที่จะลบ
    // 3. เรียก service.RemoveMember()
    err = h.memberService.RemoveMember(userID, conversationID, targetUserID)

    // 4. ส่ง WebSocket notification
    h.notificationService.NotifyUserRemovedFromConversation(targetUserID, conversationID)

    // 5. Return success
}
```

#### 2.2 Repository (Line 284-294)
```go
func (r *conversationRepository) RemoveMember(conversationID, userID uuid.UUID) error {
    // HARD DELETE จาก database
    result := r.db.Delete(&models.ConversationMember{},
        "conversation_id = ? AND user_id = ?", conversationID, userID)

    if result.RowsAffected == 0 {
        return errors.New("conversation member not found")
    }
    return nil
}
```

**ผลการทดสอบ:**
- ✅ ทำการ **HARD DELETE** record จาก `conversation_members` table
- ✅ Member ที่ถูก remove จะ**หายไป**จากระบบทันที
- ✅ การ query ครั้งถัดไปจะ**ไม่เจอ** conversation นี้

**สรุป:** Remove member ทำงานถูกต้อง ❌ ไม่ใช่ปัญหา

---

### ✅ 3. WebSocket Event `conversation.user_removed` - **ผ่าน**

**Location:** `infrastructure/adapter/websocket_adapter.go:124-131`

**Flow การส่ง Event:**

#### 3.1 Notification Service
```go
// application/serviceimpl/notification_service.go:317-320
func (s *notificationService) NotifyUserRemovedFromConversation(
    userID uuid.UUID, conversationID uuid.UUID) {
    s.wsPort.BroadcastUserRemovedFromConversation(userID, conversationID)
}
```

#### 3.2 WebSocket Adapter
```go
// infrastructure/adapter/websocket_adapter.go:124-131
func (a *WebSocketAdapter) BroadcastUserRemovedFromConversation(
    userID uuid.UUID, conversationID uuid.UUID) {

    data := map[string]interface{}{
        "conversation_id": conversationID,
        "removed_at":      utils.Now(),
    }
    a.BroadcastToUser(userID, "conversation.user_removed", data)
}
```

**Event Details:**

| Property | Value |
|----------|-------|
| **Event Type** | `conversation.user_removed` |
| **Target** | `userID` (คนที่ถูก remove) |
| **Payload** | `{ conversation_id, removed_at }` |

**ตัวอย่าง Payload:**
```json
{
  "type": "conversation.user_removed",
  "data": {
    "conversation_id": "123e4567-e89b-12d3-a456-426614174000",
    "removed_at": "2025-11-17T10:30:00Z"
  },
  "timestamp": "2025-11-17T10:30:00Z",
  "success": true
}
```

**ผลการทดสอบ:**
- ✅ Event **ถูกส่ง**ให้คนที่ถูก remove
- ✅ Event type คือ `conversation.user_removed` (**ไม่มี** prefix `message:`)
- ✅ มี `conversation_id` ให้ frontend ใช้ลบ conversation

**สรุป:** WebSocket event ส่งถูกต้อง ❌ ไม่ใช่ปัญหา

---

## 🐛 จุดที่อาจเป็นสาเหตุของปัญหา

### ⚠️ 1. **Event Type ไม่ตรงกัน** (โอกาสสูง ⭐⭐⭐)

**ปัญหา:**
- Backend ส่ง: `conversation.user_removed`
- Frontend อาจฟัง: `message:conversation.user_removed` (มี prefix `"message:"`)

**วิธีตรวจสอบ:**
```javascript
// ✅ ถูกต้อง - ควรฟังแบบนี้
socket.on('conversation.user_removed', (data) => { ... })

// ❌ ผิด - ถ้าฟังแบบนี้จะไม่ได้รับ event
socket.on('message:conversation.user_removed', (data) => { ... })
```

**แนวทางแก้:**
→ แก้ที่ Frontend: เปลี่ยนจาก `message:conversation.user_removed` เป็น `conversation.user_removed`

---

### ⚠️ 2. **Data Structure ไม่มี `user_id`** (โอกาสปานกลาง ⭐⭐)

**ปัญหา:**
- Backend ส่ง payload: `{ conversation_id, removed_at }`
- Frontend อาจต้องการ: `{ conversation_id, user_id, removed_at }`

**สาเหตุ:**
- WebSocket ส่งไปให้ **เฉพาะคนที่ถูก remove** อยู่แล้ว
- Frontend ควรรู้ว่าตัวเองคือ `current_user` จาก auth context

**แนวทางแก้:**
```javascript
// Frontend ควร handle แบบนี้
socket.on('conversation.user_removed', (data) => {
  const { conversation_id } = data;
  const current_user_id = getCurrentUserId(); // ดึงจาก auth

  // ลบ conversation ทันที
  removeConversation(conversation_id);
});
```

**หรือถ้าต้องการให้ Backend เพิ่ม `user_id`:**

```go
// แก้ที่ websocket_adapter.go:124-131
func (a *WebSocketAdapter) BroadcastUserRemovedFromConversation(
    userID uuid.UUID, conversationID uuid.UUID) {

    data := map[string]interface{}{
        "conversation_id": conversationID,
        "user_id":         userID,  // เพิ่มบรรทัดนี้
        "removed_at":      utils.Now(),
    }
    a.BroadcastToUser(userID, "conversation.user_removed", data)
}
```

---

### ⚠️ 3. **Frontend Refetch ทับ** (โอกาสปานกลาง ⭐⭐)

**ปัญหา:**
- Frontend ได้รับ event และลบ conversation แล้ว
- แต่มีการ **refetch conversations** ทันทีหลังจากนั้น
- Backend ยัง **filter ถูกต้อง** แต่ frontend มี race condition

**ตัวอย่างสถานการณ์:**
```
1. Admin remove member → Backend DELETE จาก DB
2. WebSocket event ส่งมา → Frontend ลบ conversation
3. Component remount → เรียก fetchConversations() อีกครั้ง
4. Backend ส่ง conversation list (ไม่มี conversation ที่ถูกลบ) ← ถูกต้อง
5. แต่ถ้ามี cache/state merge ผิด → conversation อาจกลับมา
```

**วิธีตรวจสอบ:**
```javascript
// ดูว่า fetchConversations() ถูกเรียกบ่อยหรือไม่
useEffect(() => {
  console.log('[DEBUG] fetchConversations called');
  fetchConversations();
}, []); // dependencies ว่างหรือมีค่าที่เปลี่ยนบ่อย?
```

**แนวทางแก้:**
- ตรวจสอบ dependencies ของ `useEffect` ที่เรียก `fetchConversations()`
- ตรวจสอบว่า state merge ทำถูกต้องหรือไม่ (ไม่ merge กับ old state)

---

### ⚠️ 4. **Frontend ไม่ได้ Handle Event** (โอกาสต่ำ ⭐)

**ปัญหา:**
- Event listener ไม่ได้ register
- หรือ register แล้วแต่ logic ไม่ทำงาน

**วิธีตรวจสอบ:**
```javascript
// เพิ่ม debug log
socket.on('conversation.user_removed', (data) => {
  console.log('[DEBUG] conversation.user_removed event received:', data);
  console.log('[DEBUG] current_user_id:', getCurrentUserId());
  console.log('[DEBUG] is_current_user:', data.user_id === getCurrentUserId());

  // ลบ conversation
  removeConversation(data.conversation_id);
});
```

---

## 🔧 แนวทางแก้ไขแบบต่างๆ

### 🎯 วิธีที่ 1: แก้ที่ Frontend (แนะนำ ⭐⭐⭐)

**เหตุผล:** Backend ทำงานถูกต้องแล้ว ไม่ควรแก้

**ขั้นตอน:**

1. **ตรวจสอบ Event Listener**
```javascript
// src/hooks/useConversation.ts หรือที่เกี่ยวข้อง
useEffect(() => {
  if (!socket) return;

  // ✅ ต้องเป็น 'conversation.user_removed' ไม่ใช่ 'message:conversation.user_removed'
  socket.on('conversation.user_removed', (data) => {
    console.log('[DEBUG] conversation.user_removed event received:', {
      conversation_id: data.conversation_id,
      current_user_id: userStore.user?.id,
      removed_at: data.removed_at
    });

    // ลบ conversation ออกจาก store
    conversationStore.removeConversation(data.conversation_id);
  });

  return () => {
    socket.off('conversation.user_removed');
  };
}, [socket]);
```

2. **ตรวจสอบ removeConversation Function**
```javascript
// src/stores/conversationStore.ts
removeConversation: (conversationId: string) => {
  console.log('[DEBUG] removeConversation called:', conversationId);

  set((state) => ({
    conversations: state.conversations.filter(
      (conv) => conv.id !== conversationId
    ),
  }));

  console.log('[DEBUG] Conversation removed successfully');
},
```

3. **ป้องกัน Refetch ทับ**
```javascript
// ตรวจสอบว่า dependencies ไม่ trigger refetch บ่อยเกินไป
useEffect(() => {
  fetchConversations();
}, []); // dependencies ว่าง = เรียกครั้งเดียวตอน mount
```

---

### 🎯 วิธีที่ 2: เพิ่ม `user_id` ใน Event (ถ้า Frontend ต้องการ)

**Location:** `infrastructure/adapter/websocket_adapter.go:124-131`

```go
func (a *WebSocketAdapter) BroadcastUserRemovedFromConversation(
    userID uuid.UUID, conversationID uuid.UUID) {

    data := map[string]interface{}{
        "conversation_id": conversationID,
        "user_id":         userID,  // ← เพิ่มบรรทัดนี้
        "removed_at":      utils.Now(),
    }
    a.BroadcastToUser(userID, "conversation.user_removed", data)
}
```

**ผลลัพธ์:**
```json
{
  "type": "conversation.user_removed",
  "data": {
    "conversation_id": "uuid",
    "user_id": "uuid",  // ← เพิ่มฟิลด์นี้
    "removed_at": "timestamp"
  }
}
```

---

### 🎯 วิธีที่ 3: เพิ่ม `removed_by` (ถ้าต้องการแสดงว่าใครลบ)

**ต้องแก้หลายที่:**

1. **Handler** - ส่ง `userID` (คนที่ลบ) เข้าไปใน notification
```go
// conversation_member_handler.go:351
h.notificationService.NotifyUserRemovedFromConversation(
    targetUserID,    // คนที่ถูกลบ
    conversationID,
    userID,          // ← เพิ่ม: คนที่ลบ
)
```

2. **Service Interface** - เพิ่ม parameter
```go
// domain/service/notification_service.go:24
NotifyUserRemovedFromConversation(
    userID, conversationID, removedBy uuid.UUID  // ← เพิ่ม removedBy
)
```

3. **Service Implementation**
```go
// application/serviceimpl/notification_service.go:317-320
func (s *notificationService) NotifyUserRemovedFromConversation(
    userID uuid.UUID, conversationID uuid.UUID, removedBy uuid.UUID) {
    s.wsPort.BroadcastUserRemovedFromConversation(userID, conversationID, removedBy)
}
```

4. **WebSocket Adapter**
```go
// infrastructure/adapter/websocket_adapter.go:124-131
func (a *WebSocketAdapter) BroadcastUserRemovedFromConversation(
    userID uuid.UUID, conversationID uuid.UUID, removedBy uuid.UUID) {

    data := map[string]interface{}{
        "conversation_id": conversationID,
        "user_id":         userID,
        "removed_by":      removedBy,  // ← เพิ่มบรรทัดนี้
        "removed_at":      utils.Now(),
    }
    a.BroadcastToUser(userID, "conversation.user_removed", data)
}
```

---

## 📋 Checklist การทดสอบ

### Frontend Developer ควรตรวจสอบ:

- [ ] Event listener ฟัง `conversation.user_removed` (**ไม่ใช่** `message:conversation.user_removed`)
- [ ] `removeConversation()` ถูกเรียกเมื่อได้รับ event
- [ ] ไม่มีการ refetch ทันทีหลังจากลบ conversation
- [ ] State management ไม่ merge กับ old state
- [ ] Debug logs แสดงว่า event ถูกรับและ process ถูกต้อง

### การทดสอบแบบ Manual:

1. เปิด **Developer Console (F12)** ทั้ง 2 ฝ่าย:
   - User A (Admin ที่จะ remove)
   - User B (Member ที่จะถูก remove)

2. **User B** เพิ่ม console.log:
```javascript
socket.on('conversation.user_removed', (data) => {
  console.log('[DEBUG] Event received:', {
    event_type: 'conversation.user_removed',
    data: data,
    current_user: getCurrentUserId(),
    will_remove: true
  });
  removeConversation(data.conversation_id);
});
```

3. **User A** remove **User B** จาก group

4. ดู console ของ **User B**:
   - ✅ ควรเห็น `[DEBUG] Event received` พร้อม data
   - ✅ Conversation ควรหายจากรายการ
   - ❌ ถ้าไม่เห็น log = event ไม่ถูกส่งมา หรือ listener ผิด

5. Refresh page ของ **User B**:
   - ✅ Conversation **ไม่ควร**กลับมาปรากฏ (เพราะ API กรองแล้ว)
   - ❌ ถ้ากลับมา = มีปัญหาที่ refetch หรือ cache

---

## 🎯 สรุปและคำแนะนำ

### ✅ Backend Status: **ทำงานถูกต้องแล้ว**

| Component | Status | Details |
|-----------|--------|---------|
| API GET /conversations | ✅ ผ่าน | กรองเฉพาะ member ที่เป็นสมาชิกอยู่ |
| API DELETE member | ✅ ผ่าน | Hard delete จาก DB ทันที |
| WebSocket Event | ✅ ผ่าน | ส่ง `conversation.user_removed` ถูกต้อง |

### 🎯 แนวทางแก้ไข: **แก้ที่ Frontend**

**โอกาสสูงสุด (เรียงลำดับ):**

1. ⭐⭐⭐ Event type ไม่ตรงกัน (`message:` prefix)
2. ⭐⭐ Event listener ไม่ทำงาน หรือไม่ได้ register
3. ⭐⭐ มี refetch ทับหลังจากลบ conversation
4. ⭐ Data structure ไม่มี `user_id` (แต่ไม่จำเป็น)

**ขั้นตอนถัดไป:**

1. ตรวจสอบ Frontend event listener ว่าฟัง `conversation.user_removed` หรือไม่
2. เพิ่ม debug logs ตามตัวอย่างข้างต้น
3. ทดสอบและดู console logs
4. Report ผลกลับมาพร้อม logs

---

**📝 หมายเหตุ:**
- Backend **ไม่ควรแก้** เพราะทำงานถูกต้องแล้ว
- ถ้า Frontend ต้องการข้อมูลเพิ่มเติมใน event payload สามารถแจ้งได้
- แนะนำให้ Frontend Developer ตรวจสอบตาม Checklist ก่อน

---

**เอกสารนี้สร้างขึ้นเมื่อ:** 2025-11-17
**Version:** 1.0
**Status:** ✅ Complete Analysis
