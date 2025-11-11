# 🎯 MASTER REFACTOR PLAN - ตัดฟีเจอร์ Business Account

**โปรเจ็ค:** ChatBiz Platform Backend v2
**วัตถุประสงค์:** ลบฟีเจอร์ Business Account ออกทั้งหมด เหลือเฉพาะ Simple Chat Platform
**วันที่สร้างแผน:** 2025-11-12
**ระดับความเสี่ยง:** 🔴 HIGH (ต้องระมัดระวังสูงสุด)

---

## 📋 สถิติโครงการ

### ก่อน Refactor:
- ไฟล์ Go ทั้งหมด: **203 files**
- Database Models: **29 models**
- Services: **26 services**
- API Endpoints: **19 route groups**

### หลัง Refactor (คาดการณ์):
- ไฟล์ Go ทั้งหมด: **~140 files** (-30%)
- Database Models: **16 models** (-13 models)
- Services: **12 services** (-14 services)
- API Endpoints: **7 route groups** (-12 groups)

### สรุปงานที่ต้องทำ:
- 🗑️ **ลบไฟล์ทั้งหมด:** 61 ไฟล์
- ✏️ **แก้ไขไฟล์:** 15 ไฟล์
- ✅ **เก็บไว้ไม่แก้:** 127 ไฟล์

---

## ⚠️ หลักการสำคัญในการ Refactor

### 1. Safety First
- ✅ Backup ก่อนทำทุกครั้ง
- ✅ Commit บ่อยๆ (ทุก phase)
- ✅ ทดสอบ compile หลังทุกการเปลี่ยนแปลง
- ✅ ไม่ลบหลายไฟล์พร้อมกัน แต่ทำทีละกลุ่ม

### 2. Outside-In Approach
ลบจาก layer นอกสุด (API/Routes) เข้าไปใน layer ในสุด (Models/Database)
```
Routes → Handlers → Services → Repositories → Models
```

### 3. Edit Before Delete
แก้ไขไฟล์ที่มี dependencies กับ Business ก่อน แล้วค่อยลบไฟล์ Business

### 4. Compile After Each Phase
ต้อง compile ผ่านหลังจากแต่ละ Phase เสร็จ

---

## 📊 Dependency Map

```
┌─────────────────────────────────────────────────────┐
│                    API Layer                         │
│  Routes → Handlers → Middleware                      │
│  [ลบได้ทันที - ไม่มีใครขึ้นต่อ]                     │
└──────────────────┬──────────────────────────────────┘
                   │
                   ↓
┌─────────────────────────────────────────────────────┐
│                Service Layer                         │
│  Business Services → Regular User Services           │
│  [ต้องแก้ Regular User Services ก่อนลบ Business]    │
└──────────────────┬──────────────────────────────────┘
                   │
                   ↓
┌─────────────────────────────────────────────────────┐
│              Repository Layer                        │
│  Business Repos → Regular User Repos                 │
│  [ลบได้หลังจาก Services หมดแล้ว]                   │
└──────────────────┬──────────────────────────────────┘
                   │
                   ↓
┌─────────────────────────────────────────────────────┐
│                Domain Layer                          │
│  Business Models ← Regular User Models               │
│  [ต้องแก้ User/Conversation/Message ก่อนลบ Business]│
└─────────────────────────────────────────────────────┘
```

---

## 🔄 ลำดับการ Refactor (12 Phases)

### 📦 PHASE 0: Preparation & Backup
**ระยะเวลาประมาณ:** 15 นาที
**ความเสี่ยง:** 🟢 LOW

#### ขั้นตอน:
1. **สำรอง Git Repository**
   ```bash
   # 1. Commit สถานะปัจจุบัน
   git add .
   git commit -m "Pre-refactor: Save current state before removing business features"

   # 2. สร้าง backup branch
   git branch backup-before-refactor

   # 3. สร้าง working branch
   git checkout -b refactor/remove-business-features

   # 4. สร้าง tag สำหรับ rollback
   git tag pre-refactor-backup
   ```

2. **สำรองฐานข้อมูล**
   ```bash
   # PostgreSQL backup
   pg_dump -U postgres -d chatbiz_db > backup_before_refactor_$(date +%Y%m%d_%H%M%S).sql
   ```

3. **สร้างโฟลเดอร์เก็บไฟล์ชั่วคราว**
   ```bash
   mkdir -p refector_plan/deleted_files
   mkdir -p refector_plan/backup_code
   ```

4. **บันทึกรายการ dependencies ปัจจุบัน**
   ```bash
   go list -m all > refector_plan/dependencies_before.txt
   ```

#### Verification:
- [ ] Git commit สำเร็จ
- [ ] Backup branch สร้างแล้ว
- [ ] Working branch สร้างแล้ว
- [ ] Tag สร้างแล้ว
- [ ] Database backup เสร็จ
- [ ] โฟลเดอร์ backup สร้างแล้ว

---

### 📦 PHASE 1: Remove Routes (API Layer)
**ระยะเวลาประมาณ:** 20 นาที
**ความเสี่ยง:** 🟢 LOW
**ไฟล์ที่ต้องจัดการ:** 11 ไฟล์

#### เป้าหมาย:
ลบ Business API Routes ออกทั้งหมด เพื่อตัดการเข้าถึง Business endpoints จาก Frontend

#### ไฟล์ที่ต้องลบ:
```
interfaces/api/routes/
  ✗ business_account_routes.go
  ✗ business_admin_routes.go
  ✗ business_follow_routes.go
  ✗ business_conversation_routes.go
  ✗ business_message_routes.go
  ✗ business_welcome_message_routes.go
  ✗ broadcast_routes.go
  ✗ broadcast_delivery_routes.go
  ✗ analytics_routes.go
  ✗ customer_profile_routes.go
  ✗ tag_routes.go
  ✗ user_tag_routes.go
```

#### ขั้นตอน:
1. **สำรองไฟล์ก่อนลบ:**
   ```bash
   cp interfaces/api/routes/business_*.go refector_plan/backup_code/
   cp interfaces/api/routes/broadcast_*.go refector_plan/backup_code/
   cp interfaces/api/routes/*profile*.go refector_plan/backup_code/
   cp interfaces/api/routes/*tag*.go refector_plan/backup_code/
   cp interfaces/api/routes/analytics_*.go refector_plan/backup_code/
   ```

2. **ลบไฟล์ทีละไฟล์:**
   ```bash
   rm interfaces/api/routes/business_account_routes.go
   rm interfaces/api/routes/business_admin_routes.go
   rm interfaces/api/routes/business_follow_routes.go
   rm interfaces/api/routes/business_conversation_routes.go
   rm interfaces/api/routes/business_message_routes.go
   rm interfaces/api/routes/business_welcome_message_routes.go
   rm interfaces/api/routes/broadcast_routes.go
   rm interfaces/api/routes/broadcast_delivery_routes.go
   rm interfaces/api/routes/analytics_routes.go
   rm interfaces/api/routes/customer_profile_routes.go
   rm interfaces/api/routes/tag_routes.go
   rm interfaces/api/routes/user_tag_routes.go
   ```

3. **ลบ imports ใน routes.go:**
   - เปิดไฟล์: `interfaces/api/routes/routes.go`
   - ลบบรรทัด import ที่ไม่ใช้แล้ว (ถ้ามี)

#### Verification:
```bash
# ตรวจสอบว่าลบครบแล้ว
ls interfaces/api/routes/ | grep -E "(business|broadcast|analytics|tag|customer)"

# ควรไม่มีผลลัพธ์ออกมา (empty output)
```

#### Expected Result:
- ไม่มี compilation error (อาจมี unused variable warnings ใน routes.go)
- ไฟล์ทั้งหมดถูกลบแล้ว

#### Git Checkpoint:
```bash
git add .
git commit -m "Phase 1: Remove business API routes (11 files)"
```

---

### 📦 PHASE 2: Remove Handlers
**ระยะเวลาประมาณ:** 20 นาที
**ความเสี่ยง:** 🟢 LOW
**ไฟล์ที่ต้องจัดการ:** 10 ไฟล์

#### เป้าหมาย:
ลบ Business Handlers ที่รับ HTTP requests

#### ไฟล์ที่ต้องลบ:
```
interfaces/api/handler/
  ✗ business_account_handler.go
  ✗ business_admin_handler.go
  ✗ business_follow_handler.go
  ✗ business_welcome_message_handler.go
  ✗ broadcast_handler.go
  ✗ broadcast_delivery_handler.go
  ✗ customer_profile_handler.go
  ✗ tag_handler.go
  ✗ user_tag_handler.go
  ✗ analytics_handler.go
```

#### ขั้นตอน:
1. **สำรองไฟล์:**
   ```bash
   cp interfaces/api/handler/business_*.go refector_plan/backup_code/
   cp interfaces/api/handler/broadcast_*.go refector_plan/backup_code/
   cp interfaces/api/handler/*tag*.go refector_plan/backup_code/
   cp interfaces/api/handler/*profile*.go refector_plan/backup_code/
   cp interfaces/api/handler/analytics_*.go refector_plan/backup_code/
   ```

2. **ลบไฟล์:**
   ```bash
   rm interfaces/api/handler/business_account_handler.go
   rm interfaces/api/handler/business_admin_handler.go
   rm interfaces/api/handler/business_follow_handler.go
   rm interfaces/api/handler/business_welcome_message_handler.go
   rm interfaces/api/handler/broadcast_handler.go
   rm interfaces/api/handler/broadcast_delivery_handler.go
   rm interfaces/api/handler/customer_profile_handler.go
   rm interfaces/api/handler/tag_handler.go
   rm interfaces/api/handler/user_tag_handler.go
   rm interfaces/api/handler/analytics_handler.go
   ```

#### Verification:
```bash
ls interfaces/api/handler/ | grep -E "(business|broadcast|analytics|tag|customer)"
```

#### Git Checkpoint:
```bash
git add .
git commit -m "Phase 2: Remove business handlers (10 files)"
```

---

### 📦 PHASE 3: Remove Middleware
**ระยะเวลาประมาณ:** 5 นาที
**ความเสี่ยง:** 🟢 LOW
**ไฟล์ที่ต้องจัดการ:** 1 ไฟล์

#### เป้าหมาย:
ลบ Business Admin Middleware

#### ไฟล์ที่ต้องลบ:
```
interfaces/api/middleware/
  ✗ business_admin.go
```

#### ขั้นตอน:
```bash
cp interfaces/api/middleware/business_admin.go refector_plan/backup_code/
rm interfaces/api/middleware/business_admin.go
```

#### Git Checkpoint:
```bash
git add .
git commit -m "Phase 3: Remove business admin middleware (1 file)"
```

---

### 📦 PHASE 4: Remove Scheduler
**ระยะเวลาประมาณ:** 10 นาที
**ความเสี่ยง:** 🟢 LOW
**ไฟล์ที่ต้องจัดการ:** 1 ไฟล์

#### เป้าหมาย:
ลบ Broadcast Scheduler

#### ไฟล์ที่ต้องลบ:
```
scheduler/
  ✗ broadcast_scheduler.go
```

#### ขั้นตอน:
```bash
cp scheduler/broadcast_scheduler.go refector_plan/backup_code/
rm scheduler/broadcast_scheduler.go
```

#### Verification:
```bash
# ตรวจสอบว่ายังมีไฟล์อื่นใน scheduler/ หรือไม่
ls scheduler/

# ถ้าไม่มีไฟล์อื่น สามารถลบโฟลเดอร์ได้
# rmdir scheduler/
```

#### Git Checkpoint:
```bash
git add .
git commit -m "Phase 4: Remove broadcast scheduler (1 file)"
```

---

### 📦 PHASE 5: Remove DTOs
**ระยะเวลาประมาณ:** 15 นาที
**ความเสี่ยง:** 🟢 LOW
**ไฟล์ที่ต้องจัดการ:** 8 ไฟล์

#### เป้าหมาย:
ลบ Business DTOs

#### ไฟล์ที่ต้องลบ:
```
domain/dto/
  ✗ business_account_dto.go
  ✗ business_admin_dto.go
  ✗ business_follow_dto.go
  ✗ business_welcome_message_dto.go
  ✗ boardcast_dto.go
  ✗ broadcast_delivery_dto.go
  ✗ customer_profile_dto.go
  ✗ analytics_dto.go
```

#### ขั้นตอน:
```bash
cp domain/dto/business_*.go refector_plan/backup_code/
cp domain/dto/*broadcast*.go refector_plan/backup_code/
cp domain/dto/*customer*.go refector_plan/backup_code/
cp domain/dto/analytics_*.go refector_plan/backup_code/

rm domain/dto/business_account_dto.go
rm domain/dto/business_admin_dto.go
rm domain/dto/business_follow_dto.go
rm domain/dto/business_welcome_message_dto.go
rm domain/dto/boardcast_dto.go
rm domain/dto/broadcast_delivery_dto.go
rm domain/dto/customer_profile_dto.go
rm domain/dto/analytics_dto.go
```

#### Verification:
```bash
ls domain/dto/ | grep -E "(business|broadcast|analytics|customer)"
```

#### Git Checkpoint:
```bash
git add .
git commit -m "Phase 5: Remove business DTOs (8 files)"
```

---

### 📦 PHASE 6: Edit Core Models (Remove Business References)
**ระยะเวลาประมาณ:** 30 นาที
**ความเสี่ยง:** 🔴 HIGH
**ไฟล์ที่ต้องจัดการ:** 3 ไฟล์

#### เป้าหมาย:
แก้ไข User, Conversation, Message models เพื่อลบ references ไปยัง Business

#### 6.1 แก้ไข User Model

**ไฟล์:** `domain/models/user.go`

**ต้องลบ:**
```go
// Business associations
OwnedBusinesses []*BusinessAccount     `json:"owned_businesses,omitempty" gorm:"foreignkey:OwnerID"`
BusinessAdmins  []*BusinessAdmin       `json:"business_admins,omitempty" gorm:"foreignkey:UserID"`
BusinessFollows []*UserBusinessFollow  `json:"business_follows,omitempty" gorm:"foreignkey:UserID"`
CustomerProfiles []*CustomerProfile    `json:"customer_profiles,omitempty" gorm:"foreignkey:UserID"`
```

**ขั้นตอน:**
1. เปิดไฟล์ `domain/models/user.go`
2. ค้นหา struct field ที่เกี่ยวกับ Business (มักอยู่ใน section "Associations")
3. ลบ 4 บรรทัดด้านบนออก
4. ลบ import ของ Business models ถ้ามี

#### 6.2 แก้ไข Conversation Model

**ไฟล์:** `domain/models/conversation.go`

**ต้องลบ:**
```go
// Business fields
BusinessID *uuid.UUID       `json:"business_id,omitempty" gorm:"type:uuid"`
Business   *BusinessAccount `json:"business,omitempty" gorm:"foreignkey:BusinessID"`
```

**ต้องแก้ไข Type constraint:**
```go
// เดิม
Type string `json:"type" gorm:"type:varchar(20);not null;check:type IN ('private','group','business')"`

// แก้เป็น
Type string `json:"type" gorm:"type:varchar(20);not null;check:type IN ('private','group')"`
```

#### 6.3 แก้ไข Message Model

**ไฟล์:** `domain/models/message.go`

**ต้องลบ:**
```go
// Business fields
BusinessID *uuid.UUID       `json:"business_id,omitempty" gorm:"type:uuid"`
Business   *BusinessAccount `json:"business,omitempty" gorm:"foreignkey:BusinessID"`
```

**ต้องแก้ไข SenderType:**
```go
// เดิม
SenderType string `json:"sender_type" gorm:"type:varchar(20);not null;check:sender_type IN ('user','business')"`

// แก้เป็น
SenderType string `json:"sender_type" gorm:"type:varchar(20);not null;default:'user';check:sender_type IN ('user')"`
```

#### Verification:
```bash
# ตรวจสอบว่าไม่มี import ของ BusinessAccount ใน 3 ไฟล์นี้
grep -n "BusinessAccount" domain/models/user.go
grep -n "BusinessAccount" domain/models/conversation.go
grep -n "BusinessAccount" domain/models/message.go

# ควรไม่มีผลลัพธ์
```

```bash
# พยายาม compile
go build ./domain/models/...

# ควร compile ผ่าน (อาจมี warnings)
```

#### Git Checkpoint:
```bash
git add domain/models/user.go domain/models/conversation.go domain/models/message.go
git commit -m "Phase 6: Remove business references from core models (User, Conversation, Message)"
```

---

### 📦 PHASE 7: Edit Services (Remove Business Logic)
**ระยะเวลาประมาณ:** 45 นาที
**ความเสี่ยง:** 🔴 HIGH
**ไฟล์ที่ต้องจัดการ:** 4 ไฟล์

#### เป้าหมาย:
แก้ไข Regular User Services เพื่อลบ Business logic และ dependencies

---

#### 7.1 แก้ไข ConversationService Interface

**ไฟล์:** `domain/service/conversation_service.go`

**ต้องลบเมธอด:**
```go
// ลบเมธอดเหล่านี้ทั้งหมด
CreateBusinessConversation(userID, businessID uuid.UUID) (*dto.ConversationDTO, error)
GetBusinessConversations(businessID, userID uuid.UUID, limit, offset int) ([]*dto.ConversationDTO, int64, error)
GetBusinessConversationsBeforeTime(businessID, userID uuid.UUID, beforeTime time.Time, limit int) ([]*dto.ConversationDTO, error)
```

#### 7.2 แก้ไข ConversationService Implementation

**ไฟล์:** `application/serviceimpl/conversations_service.go`

**ต้องลบใน Constructor:**
```go
// ลบ parameters เหล่านี้
businessRepo repository.BusinessAccountRepository
businessAdminRepo repository.BusinessAdminRepository
customerProfileRepo repository.CustomerProfileRepository
```

**ต้องลบ struct fields:**
```go
// ลบ fields เหล่านี้
businessRepo        repository.BusinessAccountRepository
businessAdminRepo   repository.BusinessAdminRepository
customerProfileRepo repository.CustomerProfileRepository
```

**ต้องลบฟังก์ชันทั้งหมด:**
```go
// ลบฟังก์ชันเหล่านี้ทั้งหมด
func (s *conversationService) CreateBusinessConversation(...)
func (s *conversationService) GetBusinessConversations(...)
func (s *conversationService) GetBusinessConversationsBeforeTime(...)
```

**ต้องลบ Logic ใน Existing Methods:**

ค้นหาและลบ logic เหล่านี้:
- ใน `GetConversations()`: ลบส่วนที่กรอง business conversations (ประมาณบรรทัด 176-183)
- ใน `mapConversationToDTO()`: ลบส่วนที่โหลดข้อมูล business (ประมาณบรรทัด 238-248)

---

#### 7.3 แก้ไข MessageService Interface

**ไฟล์:** `domain/service/message_service.go`

**ต้องลบเมธอด:**
```go
// ลบเมธอดเหล่านี้
CheckBusinessAdmin(userID, businessID uuid.UUID) (bool, bool, error)
CheckBusinessFollower(userID, businessID uuid.UUID) (bool, error)
SendBusinessTextMessage(...)
SendBusinessImageMessage(...)
// ... และเมธอด business อื่นๆ ทั้งหมด
```

#### 7.4 แก้ไข MessageService Implementation

**ไฟล์:** `application/serviceimpl/message_service.go`

**ต้องลบใน Constructor:**
```go
// ลบ parameters
businessAccountRepo repository.BusinessAccountRepository
businessAdminRepo repository.BusinessAdminRepository
```

**ต้องลบ struct fields:**
```go
// ลบ fields
businessAccountRepo repository.BusinessAccountRepository
businessAdminRepo   repository.BusinessAdminRepository
```

**ต้องลบฟังก์ชันทั้งหมด:**
```go
// ลบฟังก์ชัน
func (s *messageService) CheckBusinessAdmin(...)
func (s *messageService) CheckBusinessFollower(...)
func (s *messageService) SendBusinessTextMessage(...)
func (s *messageService) SendBusinessImageMessage(...)
// ... และฟังก์ชัน business อื่นๆ
```

---

#### 7.5 แก้ไข NotificationService Interface

**ไฟล์:** `domain/service/notification_service.go`

**ต้องลบเมธอด:**
```go
NotifyBusinessBroadcast(...)
NotifyBusinessNewFollower(...)
NotifyBusinessWelcomeMessage(...)
NotifyBusinessFollowStatusChanged(...)
NotifyBusinessStatusChanged(...)
```

#### 7.6 แก้ไข NotificationService Implementation

**ไฟล์:** `application/serviceimpl/notification_service.go`

**ต้องลบใน Constructor:**
```go
// ลบ parameter
businessAccountRepo repository.BusinessAccountRepository
```

**ต้องลบ struct field:**
```go
businessAccountRepo repository.BusinessAccountRepository
```

**ต้องลบฟังก์ชัน:**
```go
func (s *notificationService) NotifyBusinessBroadcast(...)
func (s *notificationService) NotifyBusinessNewFollower(...)
func (s *notificationService) NotifyBusinessWelcomeMessage(...)
func (s *notificationService) NotifyBusinessFollowStatusChanged(...)
func (s *notificationService) NotifyBusinessStatusChanged(...)
```

**ต้องลบ Logic ใน Existing Methods:**
- ใน `NotifyNewMessage()`: ลบส่วนที่ดึงข้อมูล business (บรรทัด 96-106)
- ใน `buildMessageDTO()`: ลบส่วนที่จัดการ business reply (บรรทัด 120-130)

---

#### Verification:
```bash
# พยายาม compile services
go build ./application/serviceimpl/...

# ควร compile ไม่ผ่าน เพราะยังมี imports และ dependencies ที่เหลือ
# แต่ error ควรเป็นเรื่อง "undefined: BusinessAccountRepository" เท่านั้น
```

#### Git Checkpoint:
```bash
git add application/serviceimpl/conversations_service.go
git add application/serviceimpl/message_service.go
git add application/serviceimpl/notification_service.go
git add domain/service/conversation_service.go
git add domain/service/message_service.go
git add domain/service/notification_service.go
git commit -m "Phase 7: Remove business logic from regular user services"
```

---

### 📦 PHASE 8: Edit WebSocket Hub
**ระยะเวลาประมาณ:** 30 นาที
**ความเสี่ยง:** 🔴 HIGH
**ไฟล์ที่ต้องจัดการ:** 3 ไฟล์

#### เป้าหมาย:
แก้ไข WebSocket Hub เพื่อลบ Business broadcasting logic

---

#### 8.1 แก้ไข hub.go

**ไฟล์:** `interfaces/websocket/hub.go`

**ต้องลบ struct fields:**
```go
// ลบ
businessConnections    map[uuid.UUID][]uuid.UUID
businessConnectionsMux sync.RWMutex
businessAdminService   service.BusinessAdminService
```

**ต้องแก้ไข Constructor:**
```go
// เดิม
func NewHub(
    conversationService service.ConversationService,
    businessAdminService service.BusinessAdminService,
    notificationService service.NotificationService,
) *Hub

// แก้เป็น
func NewHub(
    conversationService service.ConversationService,
    notificationService service.NotificationService,
) *Hub
```

**ต้องลบ Message Types:**
```go
// ลบ constants เหล่านี้
TypeBusinessBroadcast   MessageType = "business.broadcast"
TypeBusinessStatus      MessageType = "business.status"
TypeBusinessNewFollower MessageType = "business.new_follower"
```

**ต้องลบฟังก์ชัน:**
```go
func (h *Hub) loadBusinessConversations(client *Client) { ... }
func (h *Hub) sendToBusiness(...) { ... }
func (h *Hub) BroadcastToBusiness(...) { ... }
```

**ต้องลบ Logic ใน Run():**
- ค้นหาและลบส่วนที่จัดการ businessConnections

---

#### 8.2 แก้ไข handlers.go

**ไฟล์:** `interfaces/websocket/handlers.go`

**ต้องลบ:**
- การเรียก `CreateBusinessConversation()` (ประมาณบรรทัด 708-711)
- Case handlers ที่เกี่ยวกับ business

---

#### 8.3 แก้ไข broadcast.go

**ไฟล์:** `interfaces/websocket/broadcast.go`

**ต้องลบฟังก์ชันทั้งหมดที่เกี่ยวกับ business:**
```go
// ลบฟังก์ชัน
func (h *Hub) BroadcastBusinessMessage(...)
func (h *Hub) BroadcastBusinessStatus(...)
func (h *Hub) BroadcastToBusinessAdmins(...)
// ... และฟังก์ชัน business อื่นๆ
```

---

#### Verification:
```bash
go build ./interfaces/websocket/...

# อาจมี errors เกี่ยวกับ undefined types
```

#### Git Checkpoint:
```bash
git add interfaces/websocket/
git commit -m "Phase 8: Remove business logic from WebSocket hub"
```

---

### 📦 PHASE 9: Remove Service Implementations
**ระยะเวลาประมาณ:** 20 นาที
**ความเสี่ยง:** 🟡 MEDIUM
**ไฟล์ที่ต้องจัดการ:** 13 ไฟล์

#### เป้าหมาย:
ลบ Business Service Implementations

#### ไฟล์ที่ต้องลบ:
```
application/serviceimpl/
  ✗ business_account_service.go
  ✗ business_admin_service.go
  ✗ business_follow_service.go
  ✗ business_welcome_message_service.go
  ✗ broadcast_service.go
  ✗ broadcast_delivery_service.go
  ✗ customer_profile_service.go
  ✗ tag_service.go
  ✗ user_tag_service.go
  ✗ analytics_service.go
  ✗ message_send_business_service.go
  ✗ message_send_welcome_service.go
  ✗ message_send_broadcast_service.go
```

#### ขั้นตอน:
```bash
cp application/serviceimpl/business_*.go refector_plan/backup_code/
cp application/serviceimpl/broadcast_*.go refector_plan/backup_code/
cp application/serviceimpl/*tag*.go refector_plan/backup_code/
cp application/serviceimpl/*customer*.go refector_plan/backup_code/
cp application/serviceimpl/analytics_*.go refector_plan/backup_code/
cp application/serviceimpl/message_send_business*.go refector_plan/backup_code/
cp application/serviceimpl/message_send_welcome*.go refector_plan/backup_code/
cp application/serviceimpl/message_send_broadcast*.go refector_plan/backup_code/

rm application/serviceimpl/business_account_service.go
rm application/serviceimpl/business_admin_service.go
rm application/serviceimpl/business_follow_service.go
rm application/serviceimpl/business_welcome_message_service.go
rm application/serviceimpl/broadcast_service.go
rm application/serviceimpl/broadcast_delivery_service.go
rm application/serviceimpl/customer_profile_service.go
rm application/serviceimpl/tag_service.go
rm application/serviceimpl/user_tag_service.go
rm application/serviceimpl/analytics_service.go
rm application/serviceimpl/message_send_business_service.go
rm application/serviceimpl/message_send_welcome_service.go
rm application/serviceimpl/message_send_broadcast_service.go
```

#### Git Checkpoint:
```bash
git add .
git commit -m "Phase 9: Remove business service implementations (13 files)"
```

---

### 📦 PHASE 10: Remove Service Interfaces & Repositories
**ระยะเวลาประมาณ:** 25 นาที
**ความเสี่ยง:** 🟡 MEDIUM
**ไฟล์ที่ต้องจัดการ:** 20 ไฟล์

#### เป้าหมาย:
ลบ Business Service Interfaces และ Repositories

#### ไฟล์ที่ต้องลบ:
```
domain/service/
  ✗ business_account_service.go
  ✗ business_admin_service.go
  ✗ business_follow_service.go
  ✗ business_welcome_message_service.go
  ✗ broadcast_service.go
  ✗ broadcast_delivery_service.go
  ✗ customer_profile_service.go
  ✗ tag_service.go
  ✗ user_tag_service.go
  ✗ analytics_service.go

domain/repository/
  ✗ business_account_repository.go
  ✗ business_admin_repository.go
  ✗ business_follow_repository.go
  ✗ business_welcome_message_repository.go
  ✗ broadcast_repository.go
  ✗ broadcast_delivery_repository.go
  ✗ customer_profile_repository.go
  ✗ tag_repository.go
  ✗ user_tag_repository.go
  ✗ analytics_daily_repository.go

infrastructure/persistence/postgres/
  ✗ business_account_repository.go
  ✗ business_admin_repository.go
  ✗ business_follow_repository.go
  ✗ business_welcome_message_repository.go
  ✗ broadcast_repository.go
  ✗ broadcast_delivery_repository.go
  ✗ customer_profile_repository.go
  ✗ tag_repository.go
  ✗ user_tag_repository.go
  ✗ analytics_daily_repository.go
```

#### ขั้นตอน:
```bash
# Backup
cp domain/service/business_*.go refector_plan/backup_code/
cp domain/service/broadcast_*.go refector_plan/backup_code/
cp domain/service/*tag*.go refector_plan/backup_code/
cp domain/service/*customer*.go refector_plan/backup_code/
cp domain/service/analytics_*.go refector_plan/backup_code/

cp domain/repository/business_*.go refector_plan/backup_code/
cp domain/repository/broadcast_*.go refector_plan/backup_code/
cp domain/repository/*tag*.go refector_plan/backup_code/
cp domain/repository/*customer*.go refector_plan/backup_code/
cp domain/repository/analytics_*.go refector_plan/backup_code/

cp infrastructure/persistence/postgres/business_*.go refector_plan/backup_code/
cp infrastructure/persistence/postgres/broadcast_*.go refector_plan/backup_code/
cp infrastructure/persistence/postgres/*tag*.go refector_plan/backup_code/
cp infrastructure/persistence/postgres/*customer*.go refector_plan/backup_code/
cp infrastructure/persistence/postgres/analytics_*.go refector_plan/backup_code/

# Delete service interfaces
rm domain/service/business_account_service.go
rm domain/service/business_admin_service.go
rm domain/service/business_follow_service.go
rm domain/service/business_welcome_message_service.go
rm domain/service/broadcast_service.go
rm domain/service/broadcast_delivery_service.go
rm domain/service/customer_profile_service.go
rm domain/service/tag_service.go
rm domain/service/user_tag_service.go
rm domain/service/analytics_service.go

# Delete repository interfaces
rm domain/repository/business_account_repository.go
rm domain/repository/business_admin_repository.go
rm domain/repository/business_follow_repository.go
rm domain/repository/business_welcome_message_repository.go
rm domain/repository/broadcast_repository.go
rm domain/repository/broadcast_delivery_repository.go
rm domain/repository/customer_profile_repository.go
rm domain/repository/tag_repository.go
rm domain/repository/user_tag_repository.go
rm domain/repository/analytics_daily_repository.go

# Delete repository implementations
rm infrastructure/persistence/postgres/business_account_repository.go
rm infrastructure/persistence/postgres/business_admin_repository.go
rm infrastructure/persistence/postgres/business_follow_repository.go
rm infrastructure/persistence/postgres/business_welcome_message_repository.go
rm infrastructure/persistence/postgres/broadcast_repository.go
rm infrastructure/persistence/postgres/broadcast_delivery_repository.go
rm infrastructure/persistence/postgres/customer_profile_repository.go
rm infrastructure/persistence/postgres/tag_repository.go
rm infrastructure/persistence/postgres/user_tag_repository.go
rm infrastructure/persistence/postgres/analytics_daily_repository.go
```

#### Git Checkpoint:
```bash
git add .
git commit -m "Phase 10: Remove business service interfaces and repositories (30 files)"
```

---

### 📦 PHASE 11: Remove Business Models
**ระยะเวลาประมาณ:** 20 นาที
**ความเสี่ยง:** 🟡 MEDIUM
**ไฟล์ที่ต้องจัดการ:** 13 ไฟล์

#### เป้าหมาย:
ลบ Business Models

#### ไฟล์ที่ต้องลบ:
```
domain/models/
  ✗ business_account.go
  ✗ business_admin.go
  ✗ business_welcome_message.go
  ✗ broadcast.go
  ✗ broadcast_delivery.go
  ✗ customer_profile.go
  ✗ tag.go
  ✗ user_tag.go
  ✗ user_business_follow.go
  ✗ analytics_daily.go
  ✗ rich_menu.go
  ✗ rich_menu_area.go
  ✗ user_rich_menu.go
```

#### ขั้นตอน:
```bash
# Backup
cp domain/models/business_*.go refector_plan/backup_code/
cp domain/models/broadcast_*.go refector_plan/backup_code/
cp domain/models/*tag*.go refector_plan/backup_code/
cp domain/models/*customer*.go refector_plan/backup_code/
cp domain/models/analytics_*.go refector_plan/backup_code/
cp domain/models/*rich_menu*.go refector_plan/backup_code/

# Delete
rm domain/models/business_account.go
rm domain/models/business_admin.go
rm domain/models/business_welcome_message.go
rm domain/models/broadcast.go
rm domain/models/broadcast_delivery.go
rm domain/models/customer_profile.go
rm domain/models/tag.go
rm domain/models/user_tag.go
rm domain/models/user_business_follow.go
rm domain/models/analytics_daily.go
rm domain/models/rich_menu.go
rm domain/models/rich_menu_area.go
rm domain/models/user_rich_menu.go
```

#### Git Checkpoint:
```bash
git add .
git commit -m "Phase 11: Remove business models (13 files)"
```

---

### 📦 PHASE 12: Update Infrastructure (DI, Main, Migration)
**ระยะเวลาประมาณ:** 45 นาที
**ความเสี่ยง:** 🔴 HIGH
**ไฟล์ที่ต้องจัดการ:** 4 ไฟล์

#### เป้าหมาย:
อัปเดต DI Container, Main.go, Routes Setup, และ Migration

---

#### 12.1 แก้ไข DI Container

**ไฟล์:** `pkg/di/container.go`

**ต้องลบ struct fields ทั้งหมดที่เกี่ยวกับ Business:**

ใน `Container` struct ลบ:
```go
// Repositories
BusinessAccountRepo            repository.BusinessAccountRepository
BusinessAdminRepo              repository.BusinessAdminRepository
BusinessFollowRepo             repository.BusinessFollowRepository
CustomerProfileRepo            repository.CustomerProfileRepository
TagRepo                        repository.TagRepository
UserTagRepo                    repository.UserTagRepository
BusinessWelcomeMessageRepo     repository.BusinessWelcomeMessageRepository
BroadcastRepo                  repository.BroadcastRepository
BroadcastDeliveryRepo          repository.BroadcastDeliveryRepository
AnalyticsDailyRepo             repository.AnalyticsDailyRepository

// Services
BusinessAccountService         service.BusinessAccountService
BusinessAdminService           service.BusinessAdminService
BusinessFollowService          service.BusinessFollowService
CustomerProfileService         service.CustomerProfileService
TagService                     service.TagService
UserTagService                 service.UserTagService
BusinessWelcomeMessageService  service.BusinessWelcomeMessageService
BroadcastService               service.BroadcastService
BroadcastDeliveryService       service.BroadcastDeliveryService
AnalyticsService               service.AnalyticsService

// Scheduler
BroadcastScheduler             *scheduler.BroadcastScheduler

// Handlers
BusinessAccountHandler         *handler.BusinessAccountHandler
BusinessAdminHandler           *handler.BusinessAdminHandler
BusinessFollowHandler          *handler.BusinessFollowHandler
CustomerProfileHandler         *handler.CustomerProfileHandler
TagHandler                     *handler.TagHandler
UserTagHandler                 *handler.UserTagHandler
BusinessWelcomeMessageHandler  *handler.BusinessWelcomeMessageHandler
BroadcastHandler               *handler.BroadcastHandler
BroadcastDeliveryHandler       *handler.BroadcastDeliveryHandler
AnalyticsHandler               *handler.AnalyticsHandler
```

**ต้องลบการสร้าง instances:**

ใน `NewContainer()` function ลบทุกบรรทัดที่สร้าง Business components

**ต้องแก้ไข Constructor Calls:**

```go
// ConversationService - เดิม
ConversationService: serviceimpl.NewConversationService(
    container.ConversationRepo,
    container.UserRepo,
    container.BusinessAccountRepo,     // ลบ
    container.MessageRepo,
    container.BusinessAdminRepo,       // ลบ
    container.CustomerProfileRepo,     // ลบ
),

// แก้เป็น
ConversationService: serviceimpl.NewConversationService(
    container.ConversationRepo,
    container.UserRepo,
    container.MessageRepo,
),
```

```go
// MessageService - เดิม
MessageService: serviceimpl.NewMessageService(
    container.MessageRepo,
    container.MessageReadRepo,
    container.ConversationRepo,
    container.UserRepo,
    container.BusinessAccountRepo,     // ลบ
    container.BusinessAdminRepo,       // ลบ
),

// แก้เป็น
MessageService: serviceimpl.NewMessageService(
    container.MessageRepo,
    container.MessageReadRepo,
    container.ConversationRepo,
    container.UserRepo,
),
```

```go
// NotificationService - เดิม
NotificationService: serviceimpl.NewNotificationService(
    container.WebSocketPort,
    container.UserRepo,
    container.MessageRepo,
    container.ConversationRepo,
    container.BusinessAccountRepo,     // ลบ
),

// แก้เป็น
NotificationService: serviceimpl.NewNotificationService(
    container.WebSocketPort,
    container.UserRepo,
    container.MessageRepo,
    container.ConversationRepo,
),
```

```go
// WebSocket Hub - เดิม
WebSocketHub: websocket.NewHub(
    container.ConversationService,
    container.BusinessAdminService,    // ลบ
    nil,
),

// แก้เป็น
WebSocketHub: websocket.NewHub(
    container.ConversationService,
    nil,
),
```

---

#### 12.2 แก้ไข Main.go

**ไฟล์:** `cmd/api/main.go`

**ต้องลบ:**
```go
// Lines 77-86: BroadcastScheduler initialization
err = container.BroadcastScheduler.LoadScheduledBroadcasts()
if err != nil {
    log.Printf("Warning: Error loading scheduled broadcasts: %v", err)
}

err = container.BroadcastScheduler.Start()
if err != nil {
    log.Printf("Warning: Error starting broadcast scheduler: %v", err)
}

// Lines 123-126: Stop scheduler
if err := container.BroadcastScheduler.Stop(); err != nil {
    log.Printf("Error stopping scheduler: %v", err)
}
```

---

#### 12.3 แก้ไข Routes Setup

**ไฟล์:** `interfaces/api/routes/routes.go`

**ต้องลบ parameters ที่เกี่ยวกับ Business:**

ใน function signature:
```go
// เดิม
func SetupRoutes(
    app *fiber.App,
    authHandler *handler.AuthHandler,
    userHandler *handler.UserHandler,
    businessAccountHandler *handler.BusinessAccountHandler,  // ลบ
    businessAdminHandler *handler.BusinessAdminHandler,      // ลบ
    businessFollowHandler *handler.BusinessFollowHandler,    // ลบ
    // ... ลบ business handlers ทั้งหมด
    businessAdminService service.BusinessAdminService,       // ลบ
)

// แก้เป็น
func SetupRoutes(
    app *fiber.App,
    authHandler *handler.AuthHandler,
    userHandler *handler.UserHandler,
    userFriendshipHandler *handler.UserFriendshipHandler,
    conversationHandler *handler.ConversationHandler,
    messageHandler *handler.MessageHandler,
    // ... เฉพาะ regular user handlers
)
```

**ต้องลบ Route Setup Calls:**
```go
// ลบทั้งหมดนี้
SetupBusinessAccountRoutes(api, businessAccountHandler)
SetupBusinessAdminRoutes(api, businessAdminHandler, businessAdminService)
SetupBusinessFollowRoutes(api, businessFollowHandler)
SetupCustomerProfileRoutes(api, customerProfileHandler, businessAdminService)
SetupTagRoutes(api, tagHandler, businessAdminService)
SetupUserTagRoutes(api, userTagHandler, businessAdminService)
SetupAnalyticsRoutes(api, analyticsHandler, businessAdminService)
SetupBusinessConversationRoutes(api, conversationHandler, businessAdminService)
SetupBusinessMessageRoutes(api, messageHandler, businessAdminService)
SetupBusinessWelcomeMessageRoutes(api, businessWelcomeMessageHandler, businessAdminService)
SetupBroadcastRoutes(api, broadcastHandler, businessAdminService)
SetupBroadcastDeliveryRoutes(api, broadcastDeliveryHandler)
```

---

#### 12.4 แก้ไข App Setup

**ไฟล์:** `pkg/app/app.go`

**ต้องลบ Business Handlers จาก routes.SetupRoutes():**

```go
// เดิม
routes.SetupRoutes(
    app,
    container.AuthHandler,
    container.UserHandler,
    container.BusinessAccountHandler,      // ลบ
    container.BusinessAdminHandler,        // ลบ
    // ... ลบ business handlers ทั้งหมด
    container.BusinessAdminService,        // ลบ
)

// แก้เป็น
routes.SetupRoutes(
    app,
    container.AuthHandler,
    container.UserHandler,
    container.UserFriendshipHandler,
    container.ConversationHandler,
    container.MessageHandler,
    // ... เฉพาะ regular user handlers
)
```

---

#### 12.5 แก้ไข Migration

**ไฟล์:** `infrastructure/persistence/database/migration.go`

**ต้องลบ Business Models จาก AutoMigrate:**

```go
// เดิม
err := db.AutoMigrate(
    &models.User{},
    &models.BusinessAccount{},             // ลบ
    &models.BusinessAdmin{},               // ลบ
    &models.BusinessWelcomeMessage{},      // ลบ
    &models.Broadcast{},                   // ลบ
    &models.BroadcastDelivery{},           // ลบ
    &models.Tag{},                         // ลบ
    &models.UserBusinessFollow{},          // ลบ
    &models.UserTag{},                     // ลบ
    &models.CustomerProfile{},             // ลบ
    &models.AnalyticsDaily{},              // ลบ
    &models.RichMenu{},                    // ลบ
    &models.RichMenuArea{},                // ลบ
    &models.UserRichMenu{},                // ลบ
    &models.Conversation{},
    &models.ConversationMember{},
    &models.Message{},
    // ...
)

// แก้เป็น
err := db.AutoMigrate(
    &models.User{},
    &models.Conversation{},
    &models.ConversationMember{},
    &models.Message{},
    &models.MessageRead{},
    &models.MessageEditHistory{},
    &models.MessageDeleteHistory{},
    &models.UserFriendship{},
    &models.RefreshToken{},
    &models.TokenBlacklist{},
    &models.StickerSet{},
    &models.Sticker{},
    &models.UserStickerSet{},
    &models.UserFavoriteSticker{},
    &models.UserRecentSticker{},
)
```

**ต้องลบ Custom Indices สำหรับ Business:**
```go
// ลบบรรทัดเหล่านี้
db.Exec("CREATE INDEX IF NOT EXISTS idx_user_business_follows_business_id ...")
db.Exec("CREATE INDEX IF NOT EXISTS idx_broadcasts_business_id ...")
// ... indices อื่นๆ ที่เกี่ยวกับ business
```

---

#### Verification:
```bash
# พยายาม compile
go build ./...

# ควรมี errors น้อยลง
```

#### Git Checkpoint:
```bash
git add pkg/di/container.go
git add cmd/api/main.go
git add interfaces/api/routes/routes.go
git add pkg/app/app.go
git add infrastructure/persistence/database/migration.go
git commit -m "Phase 12: Update infrastructure (DI, Main, Routes, Migration)"
```

---

### 📦 PHASE 13: Final Cleanup & Testing
**ระยะเวลาประมาณ:** 60 นาที
**ความเสี่ยง:** 🟡 MEDIUM

#### เป้าหมาย:
ทำความสะอาด imports, dependencies และทดสอบทั้งระบบ

#### ขั้นตอน:

1. **ลบ Unused Imports:**
   ```bash
   # ใช้ goimports เพื่อลบ imports ที่ไม่ใช้
   go install golang.org/x/tools/cmd/goimports@latest

   # Format และลบ unused imports ทั้งโปรเจ็ค
   find . -name "*.go" -type f -exec goimports -w {} \;
   ```

2. **Clean Up Dependencies:**
   ```bash
   # ลบ dependencies ที่ไม่ใช้แล้ว
   go mod tidy

   # ตรวจสอบ dependencies ใหม่
   go list -m all > refector_plan/dependencies_after.txt

   # เปรียบเทียบกับก่อน refactor
   diff refector_plan/dependencies_before.txt refector_plan/dependencies_after.txt
   ```

3. **Compile ทั้งโปรเจ็ค:**
   ```bash
   # Compile
   go build -o chat-backend ./cmd/api

   # ต้อง compile ผ่านไม่มี errors
   ```

4. **รันโปรแกรม:**
   ```bash
   # Start database (docker)
   docker-compose up -d postgres redis

   # Run migration
   go run cmd/api/main.go migrate

   # Start application
   go run cmd/api/main.go
   ```

5. **ทดสอบ API Endpoints ที่เหลือ:**

   **Test Authentication:**
   ```bash
   # Register
   curl -X POST http://localhost:8080/api/v1/auth/register \
     -H "Content-Type: application/json" \
     -d '{"username":"testuser","email":"test@example.com","password":"test123"}'

   # Login
   curl -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"test@example.com","password":"test123"}'
   ```

   **Test User Profile:**
   ```bash
   # Get profile
   curl -X GET http://localhost:8080/api/v1/users/me \
     -H "Authorization: Bearer <YOUR_TOKEN>"
   ```

   **Test Conversations:**
   ```bash
   # Get conversations
   curl -X GET http://localhost:8080/api/v1/conversations \
     -H "Authorization: Bearer <YOUR_TOKEN>"
   ```

   **Test Messages:**
   ```bash
   # Send message
   curl -X POST http://localhost:8080/api/v1/conversations/{id}/messages \
     -H "Authorization: Bearer <YOUR_TOKEN>" \
     -H "Content-Type: application/json" \
     -d '{"content":"Hello","type":"text"}'
   ```

6. **ทดสอบ WebSocket:**
   ```bash
   # ใช้ WebSocket client tool (เช่น wscat)
   wscat -c "ws://localhost:8080/ws?token=<YOUR_TOKEN>"
   ```

7. **ตรวจสอบ Database:**
   ```bash
   # เชื่อมต่อ PostgreSQL
   psql -U postgres -d chatbiz_db

   # ตรวจสอบ tables ที่เหลือ
   \dt

   # ควรไม่มี business_* tables
   ```

#### Expected Results:
- ✅ Compile ผ่านไม่มี errors
- ✅ Application รันได้ปกติ
- ✅ API endpoints ที่เหลือทำงานได้
- ✅ WebSocket ทำงานได้
- ✅ Database migration สำเร็จ

#### Git Checkpoint:
```bash
git add .
git commit -m "Phase 13: Final cleanup and testing completed"
```

---

## 📊 ตรวจสอบผลลัพธ์สุดท้าย

### Checklist การตรวจสอบ:

#### ✅ Files Removed (61 files):
- [ ] Routes: 12 files
- [ ] Handlers: 10 files
- [ ] Middleware: 1 file
- [ ] Scheduler: 1 file
- [ ] DTOs: 8 files
- [ ] Service Implementations: 13 files
- [ ] Service Interfaces: 10 files
- [ ] Repositories: 20 files (10 + 10)
- [ ] Models: 13 files

#### ✅ Files Modified (15 files):
- [ ] domain/models/user.go
- [ ] domain/models/conversation.go
- [ ] domain/models/message.go
- [ ] domain/service/conversation_service.go
- [ ] domain/service/message_service.go
- [ ] domain/service/notification_service.go
- [ ] application/serviceimpl/conversations_service.go
- [ ] application/serviceimpl/message_service.go
- [ ] application/serviceimpl/notification_service.go
- [ ] interfaces/websocket/hub.go
- [ ] interfaces/websocket/handlers.go
- [ ] interfaces/websocket/broadcast.go
- [ ] pkg/di/container.go
- [ ] cmd/api/main.go
- [ ] interfaces/api/routes/routes.go
- [ ] pkg/app/app.go
- [ ] infrastructure/persistence/database/migration.go

#### ✅ Functionality Working:
- [ ] Authentication (Register, Login, Logout)
- [ ] User Profile Management
- [ ] Friendship System
- [ ] Direct Messaging (1-to-1)
- [ ] Group Chat
- [ ] Message Edit/Delete
- [ ] File Upload
- [ ] Stickers
- [ ] Real-time WebSocket
- [ ] Search Users

#### ✅ No Business Features:
- [ ] ไม่มี business endpoints
- [ ] ไม่มี broadcast functionality
- [ ] ไม่มี CRM features
- [ ] ไม่มี analytics
- [ ] ไม่มี business admin roles

---

## 🔄 Rollback Plan (ถ้าเกิดปัญหา)

### ถ้าต้องการ Rollback ทั้งหมด:

```bash
# 1. กลับไปยัง backup branch
git checkout backup-before-refactor

# 2. หรือใช้ tag
git checkout pre-refactor-backup

# 3. สร้าง branch ใหม่จาก backup
git checkout -b restore-from-backup backup-before-refactor
```

### ถ้าต้องการ Rollback บางส่วน:

```bash
# Rollback specific files
git checkout backup-before-refactor -- path/to/file.go

# Rollback specific phase
git revert <commit-hash-of-phase>
```

### ถ้าต้องการกู้คืน Database:

```bash
# Restore from backup
psql -U postgres -d chatbiz_db < backup_before_refactor_*.sql
```

---

## 📝 หมายเหตุสำคัญ

### ⚠️ ข้อควรระวัง:

1. **ไม่ควร skip phase ใดๆ** - ต้องทำตามลำดับ
2. **Commit บ่อยๆ** - หลังแต่ละ phase
3. **ทดสอบทุกครั้ง** - หลังแต่ละการแก้ไข
4. **อ่านคำเตือนจาก compiler** - อย่าละเว้น warnings
5. **สำรอง database** - ก่อนรัน migration ใหม่

### 🎯 เป้าหมายสุดท้าย:

หลังจาก refactor เสร็จ จะได้:

✅ **Simple Chat Platform** พร้อมฟีเจอร์:
- User Authentication
- User-to-User Messaging
- Group Chat
- File Sharing
- Stickers
- Real-time Communication

✅ **Codebase ที่สะอาดขึ้น:**
- ลดจำนวนไฟล์ 30%
- ไม่มี business complexity
- ง่ายต่อการ maintain

✅ **Performance ดีขึ้น:**
- Database queries น้อยลง
- API endpoints น้อยลง
- Binary size เล็กลง

---

## 📞 ติดต่อและสนับสนุน

ถ้ามีปัญหาระหว่าง refactor:
1. ตรวจสอบ error messages จาก compiler
2. ดู Git history เพื่อหา commit ที่ทำให้เกิดปัญหา
3. ใช้ `git diff` เพื่อดูการเปลี่ยนแปลง
4. Rollback ไปยัง phase ก่อนหน้าถ้าจำเป็น

---

**เอกสารนี้ถูกสร้างโดย:** Claude Code Assistant
**วันที่:** 2025-11-12
**Version:** 1.0.0
**Status:** ✅ Ready for Execution

**คำเตือน:** อ่านและทำความเข้าใจทุก phase ก่อนเริ่มทำ Refactor!
