# 📋 รายงานการวิเคราะห์ระบบและแผนการ Refactor

**วันที่วิเคราะห์:** 2025-11-12
**โปรเจ็ค:** ChatBiz Platform Backend v2
**เป้าหมาย:** ตัดส่วน Business Account ออกทั้งหมด เหลือเฉพาะ Regular User Features

---

## 📊 สรุปผลการวิเคราะห์

### 🏗️ ภาพรวมโครงสร้างระบบปัจจุบัน

**ChatBiz Platform** เป็นระบบแชทสำหรับธุรกิจที่พัฒนาด้วย **Go + Fiber Framework** ใช้สถาปัตยกรรมแบบ **Clean Architecture**

**Tech Stack:**
- **Backend:** Go 1.24.3, Fiber v2
- **Database:** PostgreSQL + GORM
- **Cache/Queue:** Redis
- **Storage:** Cloudinary (สำหรับรูปภาพและไฟล์)
- **Real-time:** WebSocket
- **Auth:** JWT (Access Token + Refresh Token)

**สถิติโครงสร้าง:**
- ไฟล์ Go ทั้งหมด: 203 ไฟล์
- Database Models: 29 models
- Services: 26 services
- Handlers: 19 handlers
- Repositories: 29 repositories

---

## 📁 โครงสร้างโฟลเดอร์

```
chat-backend-v2-main/
├── cmd/api/                    # Entry Point (main.go)
├── domain/                     # Business Logic Layer
│   ├── models/                # Database Models (29 models)
│   ├── dto/                   # Data Transfer Objects
│   ├── service/               # Service Interfaces
│   ├── repository/            # Repository Interfaces
│   └── types/                 # Custom Types
├── application/serviceimpl/   # Service Implementations
├── infrastructure/
│   ├── persistence/           # Data Layer
│   │   ├── postgres/         # PostgreSQL Repositories
│   │   └── database/         # Migration & Setup
│   ├── adapter/              # External Adapters
│   └── storage/              # File Storage (Cloudinary)
├── interfaces/
│   ├── api/
│   │   ├── handler/          # HTTP Handlers
│   │   ├── routes/           # Route Definitions
│   │   └── middleware/       # Auth & RBAC Middleware
│   └── websocket/            # WebSocket Hub & Handlers
├── pkg/
│   ├── app/                  # App Setup
│   ├── configs/              # Configurations
│   ├── di/                   # Dependency Injection Container
│   └── utils/                # Utility Functions
└── scheduler/                # Background Jobs
```

---

## 🎯 ฟีเจอร์ที่มีอยู่ในระบบ

### 🟢 **ฟีเจอร์ Regular User** (ต้องเก็บไว้)
1. ✅ Authentication (Register, Login, JWT)
2. ✅ User Profile Management
3. ✅ Friendship System (เพิ่มเพื่อน, ยอมรับ, ปฏิเสธ)
4. ✅ Direct Messaging (แชท 1-to-1)
5. ✅ Group Chat (แชทกลุ่ม)
6. ✅ Message Read Status (สถานะการอ่าน)
7. ✅ Message Edit & Delete (แก้ไข/ลบข้อความ)
8. ✅ File/Image Upload (อัปโหลดไฟล์/รูปภาพ)
9. ✅ Sticker System (สติกเกอร์)
10. ✅ Real-time WebSocket (การสื่อสารแบบ real-time)
11. ✅ Search Users (ค้นหาผู้ใช้)

### 🔴 **ฟีเจอร์ Business Account** (ต้องตัดออก)
1. ❌ Business Account Creation (สร้างบัญชีธุรกิจ)
2. ❌ Business Admin Management (จัดการแอดมิน)
3. ❌ Business Following (ติดตามธุรกิจ)
4. ❌ Broadcast Campaigns (ส่งข้อความแบบกระจาย)
5. ❌ Customer CRM (จัดการลูกค้า)
6. ❌ Customer Tagging (แท็กลูกค้า)
7. ❌ Welcome Messages (ข้อความต้อนรับอัตโนมัติ)
8. ❌ Business Analytics (สถิติและการวิเคราะห์)
9. ❌ Rich Menu System (เมนูโต้ตอบ)
10. ❌ Scheduled Broadcast (ส่งข้อความตามเวลาที่กำหนด)

---

## 🗑️ รายละเอียดส่วนที่ต้องตัดออก

### 1. 💾 Database Models (13 Models)

ตำแหน่ง: `domain/models/`

| ไฟล์ | Model | คำอธิบาย |
|------|-------|----------|
| `business_account.go` | BusinessAccount | บัญชีธุรกิจหลัก |
| `business_admin.go` | BusinessAdmin | ผู้ดูแลธุรกิจ |
| `user_business_follow.go` | UserBusinessFollow | การติดตามธุรกิจ |
| `broadcast.go` | Broadcast | แคมเปญส่งข้อความ |
| `broadcast_delivery.go` | BroadcastDelivery | บันทึกการส่ง broadcast |
| `business_welcome_message.go` | BusinessWelcomeMessage | ข้อความต้อนรับ |
| `customer_profile.go` | CustomerProfile | โปรไฟล์ลูกค้าใน CRM |
| `tag.go` | Tag | แท็กสำหรับจัดกลุ่มลูกค้า |
| `user_tag.go` | UserTag | การกำหนดแท็กให้ลูกค้า |
| `analytics_daily.go` | AnalyticsDaily | สถิติรายวัน |
| `rich_menu.go` | RichMenu | เมนูโต้ตอบ |
| `rich_menu_area.go` | RichMenuArea | พื้นที่คลิกบน Rich Menu |
| `user_rich_menu.go` | UserRichMenu | เชื่อมโยง User กับ Rich Menu |

### 2. 🛣️ API Routes (12 Route Files)

ตำแหน่ง: `interfaces/api/routes/`

| ไฟล์ | Endpoints | คำอธิบาย |
|------|-----------|----------|
| `business_account_routes.go` | `/api/v1/businesses/*` | จัดการบัญชีธุรกิจ |
| `business_admin_routes.go` | `/api/v1/businesses/:id/admins/*` | จัดการแอดมิน |
| `business_follow_routes.go` | `/api/v1/businesses/:id/follow` | ติดตาม/เลิกติดตาม |
| `broadcast_routes.go` | `/api/v1/businesses/:id/broadcasts/*` | จัดการ broadcasts |
| `broadcast_delivery_routes.go` | `/api/v1/broadcasts/deliveries/*` | ติดตามการส่ง |
| `customer_profile_routes.go` | `/api/v1/businesses/:id/customers/*` | CRM ลูกค้า |
| `tag_routes.go` | `/api/v1/businesses/:id/tags/*` | จัดการแท็ก |
| `user_tag_routes.go` | `/api/v1/businesses/:id/users/:userId/tags/*` | กำหนดแท็ก |
| `analytics_routes.go` | `/api/v1/businesses/:id/analytics/*` | ดูสถิติ |
| `business_welcome_message_routes.go` | `/api/v1/businesses/:id/welcome-messages/*` | ข้อความต้อนรับ |
| `business_conversation_routes.go` | `/api/v1/businesses/:id/conversations/*` | การสนทนาของธุรกิจ |
| `business_message_routes.go` | `/api/v1/businesses/:id/messages/*` | ข้อความของธุรกิจ |

### 3. 🎮 Handlers (10 Handler Files)

ตำแหน่ง: `interfaces/api/handler/`

| ไฟล์ | คำอธิบาย |
|------|----------|
| `business_account_handler.go` | จัดการ HTTP requests สำหรับบัญชีธุรกิจ |
| `business_admin_handler.go` | จัดการ HTTP requests สำหรับแอดมิน |
| `business_follow_handler.go` | จัดการการติดตาม/เลิกติดตาม |
| `broadcast_handler.go` | จัดการ broadcasts |
| `broadcast_delivery_handler.go` | จัดการสถานะการส่ง |
| `customer_profile_handler.go` | จัดการโปรไฟล์ลูกค้า |
| `tag_handler.go` | จัดการแท็ก |
| `user_tag_handler.go` | จัดการการกำหนดแท็ก |
| `analytics_handler.go` | จัดการสถิติ |
| `business_welcome_message_handler.go` | จัดการข้อความต้อนรับ |

### 4. ⚙️ Services (14 Service Files)

ตำแหน่ง: `application/serviceimpl/` (Implementation) และ `domain/service/` (Interface)

| ไฟล์ Implementation | ไฟล์ Interface | คำอธิบาย |
|---------------------|----------------|----------|
| `business_account_service.go` | `business_account_service.go` | สร้าง/อัปเดต/ลบบัญชีธุรกิจ |
| `business_admin_service.go` | `business_admin_service.go` | จัดการแอดมิน, ตรวจสอบสิทธิ์ |
| `business_follow_service.go` | `business_follow_service.go` | ติดตาม/เลิกติดตามธุรกิจ |
| `business_welcome_message_service.go` | `business_welcome_message_service.go` | จัดการข้อความต้อนรับ |
| `broadcast_service.go` | `broadcast_service.go` | สร้าง/ส่ง/จัดการ broadcasts |
| `broadcast_delivery_service.go` | `broadcast_delivery_service.go` | ติดตามสถานะการส่ง |
| `customer_profile_service.go` | `customer_profile_service.go` | จัดการโปรไฟล์ลูกค้า CRM |
| `tag_service.go` | `tag_service.go` | จัดการแท็กของธุรกิจ |
| `user_tag_service.go` | `user_tag_service.go` | กำหนดแท็กให้ลูกค้า |
| `analytics_service.go` | `analytics_service.go` | รวบรวมและวิเคราะห์ข้อมูล |
| `message_send_business_service.go` | - | ส่งข้อความในฐานะธุรกิจ |
| `message_send_broadcast_service.go` | - | ส่ง broadcast messages |
| `message_send_welcome_service.go` | - | ส่งข้อความต้อนรับ |

### 5. 🗄️ Repositories (10 Repository Files)

ตำแหน่ง: `infrastructure/persistence/postgres/` (Implementation) และ `domain/repository/` (Interface)

| ไฟล์ Implementation | ไฟล์ Interface | คำอธิบาย |
|---------------------|----------------|----------|
| `business_account_repository.go` | `business_account_repository.go` | Data access สำหรับ business_accounts |
| `business_admin_repository.go` | `business_admin_repository.go` | Data access สำหรับ business_admins |
| `business_follow_repository.go` | `business_follow_repository.go` | Data access สำหรับ user_business_follows |
| `business_welcome_message_repository.go` | `business_welcome_message_repository.go` | Data access สำหรับ business_welcome_messages |
| `broadcast_repository.go` | `broadcast_repository.go` | Data access สำหรับ broadcasts |
| `broadcast_delivery_repository.go` | `broadcast_delivery_repository.go` | Data access สำหรับ broadcast_deliveries |
| `customer_profile_repository.go` | `customer_profile_repository.go` | Data access สำหรับ customer_profiles |
| `tag_repository.go` | `tag_repository.go` | Data access สำหรับ tags |
| `user_tag_repository.go` | `user_tag_repository.go` | Data access สำหรับ user_tags |
| `analytics_daily_repository.go` | `analytics_daily_repository.go` | Data access สำหรับ analytics_daily |

### 6. 🛡️ Middleware (1 File)

ตำแหน่ง: `interfaces/api/middleware/`

| ไฟล์ | Functions | คำอธิบาย |
|------|-----------|----------|
| `business_admin.go` | `CheckBusinessAdmin()`, `CheckBusinessAdminWithRoles()` | ตรวจสอบสิทธิ์แอดมินธุรกิจ |

### 7. ⏰ Scheduler (1 File)

ตำแหน่ง: `scheduler/`

| ไฟล์ | คำอธิบาย |
|------|----------|
| `broadcast_scheduler.go` | ตรวจสอบและส่ง broadcasts ที่ตั้งเวลาไว้ (ใช้ Redis + 5 workers) |

### 8. 📝 DTOs (Business-related)

ตำแหน่ง: `domain/dto/`

ต้องตรวจสอบและลบ DTO files ที่เกี่ยวข้องกับ Business, เช่น:
- `business_*.go`
- `broadcast_*.go`
- `customer_*.go`
- `tag_*.go`
- `analytics_*.go`

---

## ⚠️ ส่วนที่ต้องตรวจสอบและแก้ไข

### 1. 👤 User Model

**ไฟล์:** `domain/models/user.go`

**ต้องลบ Relations:**
```go
// ลบทั้ง 4 relations นี้
OwnedBusinesses   []BusinessAccount      `gorm:"foreignKey:OwnerID"`
BusinessAdmins    []BusinessAdmin        `gorm:"foreignKey:UserID"`
CustomerProfiles  []CustomerProfile      `gorm:"foreignKey:UserID"`
BusinessFollows   []UserBusinessFollow   `gorm:"foreignKey:UserID"`
```

### 2. 💬 Conversation Model

**ไฟล์:** `domain/models/conversation.go`

**ต้องตรวจสอบ:**
- Field `Type` มี value `"business"` → ต้องลบ type นี้ (เหลือ "private", "group")
- Field `BusinessID *uint` → ต้องลบ field นี้ออก
- Relation `Business *BusinessAccount` → ต้องลบ

**ตัวอย่างที่ต้องแก้:**
```go
// เดิม
Type string `gorm:"type:varchar(20);not null"` // private, group, business

// แก้เป็น
Type string `gorm:"type:varchar(20);not null"` // private, group
```

```go
// ลบ field นี้
BusinessID *uint `gorm:"index"`
Business   *BusinessAccount `gorm:"foreignKey:BusinessID"`
```

### 3. 📨 Message Model

**ไฟล์:** `domain/models/message.go`

**ต้องตรวจสอบ:**
- Field `SenderType` มี value `"business"` → ต้องลบ (เหลือ "user")
- Field `BusinessID *uint` → ต้องลบ field นี้ออก
- Relation `Business *BusinessAccount` → ต้องลบ

**ตัวอย่างที่ต้องแก้:**
```go
// เดิม
SenderType string `gorm:"type:varchar(20);not null"` // user, business

// แก้เป็น
SenderType string `gorm:"type:varchar(20);not null;default:'user'"` // user
```

```go
// ลบ field นี้
BusinessID *uint `gorm:"index"`
Business   *BusinessAccount `gorm:"foreignKey:BusinessID"`
```

### 4. 💼 Conversation Service

**ไฟล์:** `application/serviceimpl/conversations_service.go`

**ต้องตรวจสอบและลบ Logic:**
- การสร้าง business conversation
- การตรวจสอบสิทธิ์ business admin
- การจัดการ business-related conversations

**ตัวอย่าง:**
```go
// ต้องหาและลบ logic ประมาณนี้
if conversation.Type == "business" {
    // business logic...
}

if businessID != nil {
    // business logic...
}
```

### 5. 📤 Message Service

**ไฟล์:** `application/serviceimpl/message_service.go`

**ต้องตรวจสอบและลบ Logic:**
- การส่งข้อความในฐานะธุรกิจ
- การส่ง broadcast messages
- การตรวจสอบ BusinessID

**ตัวอย่าง:**
```go
// ต้องหาและลบ logic ประมาณนี้
if message.SenderType == "business" {
    // business logic...
}

if message.BusinessID != nil {
    // business logic...
}
```

### 6. 🔌 WebSocket Hub

**ไฟล์:** `interfaces/websocket/hub.go`, `broadcast.go`

**ต้องตรวจสอบและลบ Logic:**
- การ broadcast ข้อความธุรกิจ
- Notification เกี่ยวกับ business events

### 7. 📦 Dependency Injection Container

**ไฟล์:** `pkg/di/container.go`

**ต้องลบ Dependencies ทั้งหมดที่เกี่ยวกับ Business:**
- BusinessAccountRepo/Service/Handler
- BusinessAdminRepo/Service/Handler
- BusinessFollowRepo/Service/Handler
- BroadcastRepo/Service/Handler
- BroadcastDeliveryRepo/Service/Handler
- CustomerProfileRepo/Service/Handler
- TagRepo/Service/Handler
- UserTagRepo/Service/Handler
- AnalyticsRepo/Service/Handler
- BusinessWelcomeMessageRepo/Service/Handler

### 8. 🚀 Main Entry Point

**ไฟล์:** `cmd/api/main.go`

**ต้องลบ:**
```go
// ลบการ initialize BroadcastScheduler
broadcastScheduler := scheduler.NewBroadcastScheduler(...)
broadcastScheduler.Start()
defer broadcastScheduler.Stop()
```

**ต้องลบการ Setup Routes:**
```go
// ลบทุกบรรทัดที่เรียก Business routes
routes.SetupBusinessAccountRoutes(...)
routes.SetupBusinessAdminRoutes(...)
routes.SetupBusinessFollowRoutes(...)
routes.SetupBroadcastRoutes(...)
routes.SetupBusinessWelcomeMessageRoutes(...)
routes.SetupCustomerProfileRoutes(...)
routes.SetupTagRoutes(...)
routes.SetupUserTagRoutes(...)
routes.SetupAnalyticsRoutes(...)
routes.SetupBusinessConversationRoutes(...)
routes.SetupBusinessMessageRoutes(...)
```

### 9. 🗃️ Database Migration

**ไฟล์:** `infrastructure/persistence/database/migration.go`

**ต้องลบ Models จาก AutoMigrate:**
```go
// ลบทั้งหมดนี้
&models.BusinessAccount{},
&models.BusinessAdmin{},
&models.BusinessWelcomeMessage{},
&models.Broadcast{},
&models.BroadcastDelivery{},
&models.Tag{},
&models.UserTag{},
&models.UserBusinessFollow{},
&models.AnalyticsDaily{},
&models.CustomerProfile{},
&models.RichMenu{},
&models.RichMenuArea{},
&models.UserRichMenu{},
```

**ต้องลบ Custom Indices:**
```go
// ลบ indices ที่เกี่ยวข้อง
db.Exec("CREATE INDEX IF NOT EXISTS idx_user_business_follows_business_id ON user_business_follows(business_id)")
db.Exec("CREATE INDEX IF NOT EXISTS idx_broadcasts_business_id ON broadcasts(business_id)")
// ... etc
```

---

## 📋 แผนการดำเนินการ Refactor

### 🎯 ลำดับขั้นตอนที่แนะนำ:

#### Phase 1: เตรียมการและ Backup
1. ✅ **Backup โปรเจ็คทั้งหมด** (สำคัญมาก!)
2. ✅ สร้าง branch ใหม่สำหรับ refactor: `git checkout -b refactor/remove-business-features`
3. ✅ Commit สถานะปัจจุบัน

#### Phase 2: ลบ API Layer (Frontend-facing)
4. ❌ ลบ Route files (12 files) จาก `interfaces/api/routes/`
5. ❌ ลบ Handler files (10 files) จาก `interfaces/api/handler/`
6. ❌ ลบ Business Middleware (`business_admin.go`)

#### Phase 3: ลบ Business Logic Layer
7. ❌ ลบ Service Implementation files (14 files) จาก `application/serviceimpl/`
8. ❌ ลบ Service Interface files จาก `domain/service/`

#### Phase 4: ลบ Data Access Layer
9. ❌ ลบ Repository Implementation files (10 files) จาก `infrastructure/persistence/postgres/`
10. ❌ ลบ Repository Interface files จาก `domain/repository/`

#### Phase 5: ลบ Domain Layer
11. ❌ ลบ Model files (13 files) จาก `domain/models/`
12. ❌ ลบ DTO files ที่เกี่ยวข้องจาก `domain/dto/`

#### Phase 6: ลบ Scheduler & Background Jobs
13. ❌ ลบ `scheduler/broadcast_scheduler.go`

#### Phase 7: แก้ไข Core Models
14. ⚠️ แก้ไข `domain/models/user.go` - ลบ Business relations
15. ⚠️ แก้ไข `domain/models/conversation.go` - ลบ BusinessID และ type "business"
16. ⚠️ แก้ไข `domain/models/message.go` - ลบ BusinessID และ SenderType "business"

#### Phase 8: แก้ไข Services ที่เหลือ
17. ⚠️ แก้ไข `application/serviceimpl/conversations_service.go` - ลบ business logic
18. ⚠️ แก้ไข `application/serviceimpl/message_service.go` - ลบ business logic
19. ⚠️ แก้ไข `interfaces/websocket/` - ลบ business broadcasting logic

#### Phase 9: อัปเดต Infrastructure
20. ⚠️ แก้ไข `pkg/di/container.go` - ลบ Business dependencies
21. ⚠️ แก้ไข `cmd/api/main.go` - ลบ Scheduler และ Business routes
22. ⚠️ แก้ไข `infrastructure/persistence/database/migration.go` - ลบ Business models

#### Phase 10: Testing & Cleanup
23. 🧪 รัน `go mod tidy` เพื่อลบ unused dependencies
24. 🧪 รัน `go build` เพื่อตรวจสอบ compilation errors
25. 🧪 แก้ไข import errors ทั้งหมด
26. 🧪 ทดสอบ API endpoints ที่เหลือ
27. 🧪 ทดสอบ WebSocket connections
28. 🧪 ทดสอบ Database migrations

#### Phase 11: Database Cleanup (ข้อมูลจริง)
29. 🗑️ สำรองฐานข้อมูลปัจจุบัน
30. 🗑️ Drop tables ที่เกี่ยวกับ Business (ถ้าต้องการ)
31. 🗑️ หรือ Migrate เฉพาะ tables ที่เหลือไปยัง database ใหม่

---

## ⚠️ ข้อควรระวัง

### 1. Database Relations
- User model มี foreign key relations กับ Business models หลายตัว
- ต้องระวังเรื่อง CASCADE DELETE
- ควร backup database ก่อนทำการลบ

### 2. WebSocket Events
- อาจมี event types ที่เกี่ยวกับ Business ที่ต้องลบออก
- ตรวจสอบ `interfaces/websocket/handlers.go`

### 3. DTOs & Validation
- DTO structs หลายตัวอาจมี business-related fields
- ต้องตรวจสอบและลบให้ครบ

### 4. Configuration Files
- ตรวจสอบ `.env` ว่ามี business-related configs หรือไม่
- ตรวจสอบ `pkg/configs/` ว่ามี business configs หรือไม่

### 5. Tests
- ถ้ามี test files ที่เกี่ยวกับ Business ก็ต้องลบด้วย
- อัปเดต test suites สำหรับฟีเจอร์ที่เหลือ

### 6. Dependencies ที่อาจไม่ใช้แล้ว
- หลังจาก refactor เสร็จ ควรรัน `go mod tidy`
- ตรวจสอบว่า dependencies ใดที่ไม่ใช้แล้วและอาจถอนได้

---

## 📊 สรุปผลกระทบ

### จำนวนไฟล์ที่ต้องจัดการ:

| ประเภท | ลบทั้งไฟล์ | แก้ไขบางส่วน | เก็บไว้ |
|--------|------------|--------------|---------|
| Models | 13 | 3 (User, Conversation, Message) | 13 |
| Routes | 12 | 0 | 7 |
| Handlers | 10 | 0 | 9 |
| Services | 14 | 2 (Conversation, Message) | 10 |
| Repositories | 10 | 0 | 19 |
| Middleware | 1 | 0 | 2 |
| Scheduler | 1 | 0 | 0 |
| Infrastructure | 0 | 3 (DI, Main, Migration) | - |
| **รวม** | **61 files** | **8 files** | **60 files** |

### ฟีเจอร์ที่เหลือหลัง Refactor:

**Simple Chat Platform** พร้อมฟีเจอร์:
- 👤 User Authentication & Profile
- 👥 Friendship System
- 💬 Direct Messaging (1-to-1)
- 👨‍👨‍👧‍👦 Group Chat
- 📎 File/Image Upload
- 😀 Sticker System
- 🔔 Real-time Notifications (WebSocket)
- 🔍 User Search
- ✏️ Message Edit/Delete
- 👀 Read Status

---

## 🔍 Checklist สำหรับการ Refactor

### Pre-Refactor Checklist
- [ ] Backup โปรเจ็คทั้งหมด
- [ ] Backup database
- [ ] สร้าง Git branch ใหม่
- [ ] Commit สถานะปัจจุบัน
- [ ] อ่านเอกสารนี้ให้เข้าใจทั้งหมด

### Refactor Checklist (ตาม Phase)
- [ ] Phase 1: เตรียมการและ Backup
- [ ] Phase 2: ลบ API Layer
- [ ] Phase 3: ลบ Business Logic Layer
- [ ] Phase 4: ลบ Data Access Layer
- [ ] Phase 5: ลบ Domain Layer
- [ ] Phase 6: ลบ Scheduler
- [ ] Phase 7: แก้ไข Core Models
- [ ] Phase 8: แก้ไข Services ที่เหลือ
- [ ] Phase 9: อัปเดต Infrastructure
- [ ] Phase 10: Testing & Cleanup

### Post-Refactor Checklist
- [ ] โปรเจ็ค compile ผ่าน (`go build`)
- [ ] ไม่มี import errors
- [ ] Test API endpoints ทั้งหมด
- [ ] Test WebSocket connections
- [ ] Test Database migrations
- [ ] รัน `go mod tidy`
- [ ] อัปเดต README.md (ถ้ามี)
- [ ] อัปเดต API documentation (ถ้ามี)
- [ ] Commit การเปลี่ยนแปลง
- [ ] สร้าง Pull Request

---

## 📞 สิ่งที่ต้องตัดสินใจเพิ่มเติม

1. **Database Cleanup:**
   - จะ drop tables ที่เกี่ยวกับ Business ทันทีหรือเก็บไว้?
   - จะ migrate ข้อมูล User ที่เหลือไปยัง database ใหม่หรือไม่?

2. **API Versioning:**
   - จะเปลี่ยน API version จาก v1 เป็น v2 หรือไม่?
   - จะรักษา backward compatibility หรือไม่?

3. **Documentation:**
   - จะอัปเดต API docs พร้อมกับ refactor หรือทีหลัง?
   - จะสร้าง migration guide สำหรับผู้ใช้เดิมหรือไม่?

4. **Testing:**
   - จะเขียน unit tests ใหม่หรือไม่?
   - จะทำ integration testing หรือไม่?

---

## 🎉 ผลลัพธ์ที่คาดหวัง

หลังจาก Refactor เสร็จสิ้น จะได้:

1. ✅ **Simple Chat Platform** ที่เรียบง่ายและมีเฉพาะฟีเจอร์ User-to-User
2. ✅ **Codebase ที่สะอาดขึ้น** ลดจำนวนไฟล์ลง ~30%
3. ✅ **Maintenance ง่ายขึ้น** เพราะมี complexity น้อยลง
4. ✅ **Performance ดีขึ้น** เพราะมี database queries น้อยลง
5. ✅ **Deploy เร็วขึ้น** เพราะ binary size เล็กลง

---

**หมายเหตุ:** เอกสารนี้เป็นเพียง roadmap สำหรับการ refactor ควรทบทวนแต่ละขั้นตอนอีกครั้งก่อนดำเนินการจริง และควรทำทีละ phase พร้อม commit เป็นระยะเพื่อความปลอดภัย

**วันที่สร้าง:** 2025-11-12
**Version:** 1.0
**Status:** Ready for Review
