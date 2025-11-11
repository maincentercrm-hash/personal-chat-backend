# ✅ REFACTOR CHECKLIST - ติดตามความคืบหน้า

**โปรเจ็ค:** ChatBiz Platform → Simple Chat Platform
**เริ่มวันที่:** __________
**คาดว่าเสร็จวันที่:** __________

---

## 📋 Pre-Refactor Checklist

### เตรียมการก่อนเริ่ม:
- [ ] อ่าน `MASTER_REFACTOR_PLAN.md` ทั้งหมดแล้ว
- [ ] อ่าน `result_system.md` (รายงานการวิเคราะห์) แล้ว
- [ ] เข้าใจ dependencies ระหว่างไฟล์ทั้งหมดแล้ว
- [ ] มีเวลาพอที่จะทำให้เสร็จ (ประมาณ 4-6 ชั่วโมง)
- [ ] ไม่มีงานเร่งด่วนอื่นที่ต้องทำระหว่างนี้

---

## 🔧 PHASE 0: Preparation & Backup

### Git Backup:
- [ ] Commit สถานะปัจจุบัน: `git commit -m "Pre-refactor: Save current state"`
- [ ] สร้าง backup branch: `git branch backup-before-refactor`
- [ ] สร้าง working branch: `git checkout -b refactor/remove-business-features`
- [ ] สร้าง tag: `git tag pre-refactor-backup`
- [ ] ตรวจสอบ branch: `git branch` (ต้องอยู่ที่ `refactor/remove-business-features`)

### Database Backup:
- [ ] สำรอง PostgreSQL: `pg_dump -U postgres -d chatbiz_db > backup_before_refactor.sql`
- [ ] ตรวจสอบไฟล์ backup มีขนาดมากกว่า 0 bytes
- [ ] เก็บไฟล์ backup ไว้ในที่ปลอดภัย

### Folder Setup:
- [ ] สร้างโฟลเดอร์: `mkdir -p refector_plan/deleted_files`
- [ ] สร้างโฟลเดอร์: `mkdir -p refector_plan/backup_code`

### Dependencies:
- [ ] บันทึก dependencies: `go list -m all > refector_plan/dependencies_before.txt`

**เวลาที่ใช้:** ________ นาที
**หมายเหตุ:**
```


```

---

## 📦 PHASE 1: Remove Routes (API Layer)

### ลบไฟล์ (12 files):
- [ ] `business_account_routes.go`
- [ ] `business_admin_routes.go`
- [ ] `business_follow_routes.go`
- [ ] `business_conversation_routes.go`
- [ ] `business_message_routes.go`
- [ ] `business_welcome_message_routes.go`
- [ ] `broadcast_routes.go`
- [ ] `broadcast_delivery_routes.go`
- [ ] `analytics_routes.go`
- [ ] `customer_profile_routes.go`
- [ ] `tag_routes.go`
- [ ] `user_tag_routes.go`

### Backup:
- [ ] สำรองไฟล์ทั้งหมดไปยัง `refector_plan/backup_code/`

### Verification:
- [ ] รัน: `ls interfaces/api/routes/ | grep -E "(business|broadcast|analytics|tag|customer)"`
- [ ] ผลลัพธ์ว่างเปล่า (ไม่มีไฟล์ business)

### Git Commit:
- [ ] `git add .`
- [ ] `git commit -m "Phase 1: Remove business API routes (12 files)"`

**สถานะ:** ⬜ ยังไม่เริ่ม | 🟡 กำลังทำ | ✅ เสร็จสมบูรณ์
**เวลาที่ใช้:** ________ นาที
**ปัญหาที่พบ:**
```


```

---

## 📦 PHASE 2: Remove Handlers

### ลบไฟล์ (10 files):
- [ ] `business_account_handler.go`
- [ ] `business_admin_handler.go`
- [ ] `business_follow_handler.go`
- [ ] `business_welcome_message_handler.go`
- [ ] `broadcast_handler.go`
- [ ] `broadcast_delivery_handler.go`
- [ ] `customer_profile_handler.go`
- [ ] `tag_handler.go`
- [ ] `user_tag_handler.go`
- [ ] `analytics_handler.go`

### Backup:
- [ ] สำรองไฟล์ทั้งหมดไปยัง `refector_plan/backup_code/`

### Verification:
- [ ] รัน: `ls interfaces/api/handler/ | grep -E "(business|broadcast|analytics|tag|customer)"`
- [ ] ผลลัพธ์ว่างเปล่า

### Git Commit:
- [ ] `git add .`
- [ ] `git commit -m "Phase 2: Remove business handlers (10 files)"`

**สถานะ:** ⬜ ยังไม่เริ่ม | 🟡 กำลังทำ | ✅ เสร็จสมบูรณ์
**เวลาที่ใช้:** ________ นาที

---

## 📦 PHASE 3: Remove Middleware

### ลบไฟล์ (1 file):
- [ ] `interfaces/api/middleware/business_admin.go`

### Backup:
- [ ] สำรองไฟล์

### Git Commit:
- [ ] `git add .`
- [ ] `git commit -m "Phase 3: Remove business admin middleware (1 file)"`

**สถานะ:** ⬜ ยังไม่เริ่ม | 🟡 กำลังทำ | ✅ เสร็จสมบูรณ์
**เวลาที่ใช้:** ________ นาที

---

## 📦 PHASE 4: Remove Scheduler

### ลบไฟล์ (1 file):
- [ ] `scheduler/broadcast_scheduler.go`

### Backup:
- [ ] สำรองไฟล์

### Additional:
- [ ] ตรวจสอบว่ายังมีไฟล์อื่นใน `scheduler/` หรือไม่
- [ ] ถ้าไม่มี สามารถลบโฟลเดอร์ได้

### Git Commit:
- [ ] `git add .`
- [ ] `git commit -m "Phase 4: Remove broadcast scheduler (1 file)"`

**สถานะ:** ⬜ ยังไม่เริ่ม | 🟡 กำลังทำ | ✅ เสร็จสมบูรณ์
**เวลาที่ใช้:** ________ นาที

---

## 📦 PHASE 5: Remove DTOs

### ลบไฟล์ (8 files):
- [ ] `business_account_dto.go`
- [ ] `business_admin_dto.go`
- [ ] `business_follow_dto.go`
- [ ] `business_welcome_message_dto.go`
- [ ] `boardcast_dto.go`
- [ ] `broadcast_delivery_dto.go`
- [ ] `customer_profile_dto.go`
- [ ] `analytics_dto.go`

### Backup:
- [ ] สำรองไฟล์ทั้งหมด

### Verification:
- [ ] รัน: `ls domain/dto/ | grep -E "(business|broadcast|analytics|customer)"`
- [ ] ผลลัพธ์ว่างเปล่า

### Git Commit:
- [ ] `git add .`
- [ ] `git commit -m "Phase 5: Remove business DTOs (8 files)"`

**สถานะ:** ⬜ ยังไม่เริ่ม | 🟡 กำลังทำ | ✅ เสร็จสมบูรณ์
**เวลาที่ใช้:** ________ นาที

---

## 📦 PHASE 6: Edit Core Models ⚠️ HIGH RISK

### 6.1 แก้ไข User Model:
- [ ] เปิดไฟล์: `domain/models/user.go`
- [ ] ลบ field: `OwnedBusinesses`
- [ ] ลบ field: `BusinessAdmins`
- [ ] ลบ field: `BusinessFollows`
- [ ] ลบ field: `CustomerProfiles`
- [ ] ลบ import ของ Business models (ถ้ามี)
- [ ] ตรวจสอบ: `grep -n "BusinessAccount" domain/models/user.go` (ไม่มีผลลัพธ์)

### 6.2 แก้ไข Conversation Model:
- [ ] เปิดไฟล์: `domain/models/conversation.go`
- [ ] ลบ field: `BusinessID *uuid.UUID`
- [ ] ลบ field: `Business *BusinessAccount`
- [ ] แก้ Type constraint: ลบ `'business'` ออก
- [ ] ตรวจสอบ: `grep -n "BusinessAccount" domain/models/conversation.go` (ไม่มีผลลัพธ์)

### 6.3 แก้ไข Message Model:
- [ ] เปิดไฟล์: `domain/models/message.go`
- [ ] ลบ field: `BusinessID *uuid.UUID`
- [ ] ลบ field: `Business *BusinessAccount`
- [ ] แก้ SenderType constraint: ลบ `'business'` ออก
- [ ] ตรวจสอบ: `grep -n "BusinessAccount" domain/models/message.go` (ไม่มีผลลัพธ์)

### Compile Test:
- [ ] รัน: `go build ./domain/models/...`
- [ ] Compile ผ่าน (อาจมี warnings)

### Git Commit:
- [ ] `git add domain/models/user.go domain/models/conversation.go domain/models/message.go`
- [ ] `git commit -m "Phase 6: Remove business references from core models"`

**สถานะ:** ⬜ ยังไม่เริ่ม | 🟡 กำลังทำ | ✅ เสร็จสมบูรณ์
**เวลาที่ใช้:** ________ นาที
**ปัญหาที่พบ:**
```


```

---

## 📦 PHASE 7: Edit Services ⚠️ HIGH RISK

### 7.1 ConversationService Interface:
- [ ] เปิดไฟล์: `domain/service/conversation_service.go`
- [ ] ลบเมธอด: `CreateBusinessConversation()`
- [ ] ลบเมธอด: `GetBusinessConversations()`
- [ ] ลบเมธอด: `GetBusinessConversationsBeforeTime()`

### 7.2 ConversationService Implementation:
- [ ] เปิดไฟล์: `application/serviceimpl/conversations_service.go`
- [ ] ลบ constructor parameters: `businessRepo`, `businessAdminRepo`, `customerProfileRepo`
- [ ] ลบ struct fields: `businessRepo`, `businessAdminRepo`, `customerProfileRepo`
- [ ] ลบฟังก์ชัน `CreateBusinessConversation()`
- [ ] ลบฟังก์ชัน `GetBusinessConversations()`
- [ ] ลบฟังก์ชัน `GetBusinessConversationsBeforeTime()`
- [ ] ลบ business logic ใน `GetConversations()` (บรรทัด ~176-183)
- [ ] ลบ business logic ใน `mapConversationToDTO()` (บรรทัด ~238-248)

### 7.3 MessageService Interface:
- [ ] เปิดไฟล์: `domain/service/message_service.go`
- [ ] ลบเมธอด: `CheckBusinessAdmin()`
- [ ] ลบเมธอด: `CheckBusinessFollower()`
- [ ] ลบเมธอด: `SendBusinessTextMessage()`
- [ ] ลบเมธอด: `SendBusinessImageMessage()`

### 7.4 MessageService Implementation:
- [ ] เปิดไฟล์: `application/serviceimpl/message_service.go`
- [ ] ลบ constructor parameters: `businessAccountRepo`, `businessAdminRepo`
- [ ] ลบ struct fields: `businessAccountRepo`, `businessAdminRepo`
- [ ] ลบฟังก์ชัน `CheckBusinessAdmin()`
- [ ] ลบฟังก์ชัน `CheckBusinessFollower()`
- [ ] ลบฟังก์ชัน `SendBusinessTextMessage()`
- [ ] ลบฟังก์ชัน `SendBusinessImageMessage()`

### 7.5 NotificationService Interface:
- [ ] เปิดไฟล์: `domain/service/notification_service.go`
- [ ] ลบเมธอดทั้งหมดที่เกี่ยวกับ Business

### 7.6 NotificationService Implementation:
- [ ] เปิดไฟล์: `application/serviceimpl/notification_service.go`
- [ ] ลบ constructor parameter: `businessAccountRepo`
- [ ] ลบ struct field: `businessAccountRepo`
- [ ] ลบฟังก์ชันทั้งหมดที่เกี่ยวกับ Business
- [ ] ลบ business logic ใน `NotifyNewMessage()` (บรรทัด ~96-106)
- [ ] ลบ business logic ใน `buildMessageDTO()` (บรรทัด ~120-130)

### Git Commit:
- [ ] `git add application/serviceimpl/ domain/service/`
- [ ] `git commit -m "Phase 7: Remove business logic from regular user services"`

**สถานะ:** ⬜ ยังไม่เริ่ม | 🟡 กำลังทำ | ✅ เสร็จสมบูรณ์
**เวลาที่ใช้:** ________ นาที

---

## 📦 PHASE 8: Edit WebSocket Hub ⚠️ HIGH RISK

### 8.1 hub.go:
- [ ] ลบ struct fields: `businessConnections`, `businessConnectionsMux`, `businessAdminService`
- [ ] แก้ Constructor: ลบ parameter `businessAdminService`
- [ ] ลบ Message Types: `TypeBusinessBroadcast`, `TypeBusinessStatus`, `TypeBusinessNewFollower`
- [ ] ลบฟังก์ชัน: `loadBusinessConversations()`
- [ ] ลบฟังก์ชัน: `sendToBusiness()`
- [ ] ลบฟังก์ชัน: `BroadcastToBusiness()`

### 8.2 handlers.go:
- [ ] ลบการเรียก `CreateBusinessConversation()`
- [ ] ลบ business case handlers

### 8.3 broadcast.go:
- [ ] ลบฟังก์ชันทั้งหมดที่เกี่ยวกับ business

### Git Commit:
- [ ] `git add interfaces/websocket/`
- [ ] `git commit -m "Phase 8: Remove business logic from WebSocket hub"`

**สถานะ:** ⬜ ยังไม่เริ่ม | 🟡 กำลังทำ | ✅ เสร็จสมบูรณ์
**เวลาที่ใช้:** ________ นาที

---

## 📦 PHASE 9: Remove Service Implementations

### ลบไฟล์ (13 files):
- [ ] `business_account_service.go`
- [ ] `business_admin_service.go`
- [ ] `business_follow_service.go`
- [ ] `business_welcome_message_service.go`
- [ ] `broadcast_service.go`
- [ ] `broadcast_delivery_service.go`
- [ ] `customer_profile_service.go`
- [ ] `tag_service.go`
- [ ] `user_tag_service.go`
- [ ] `analytics_service.go`
- [ ] `message_send_business_service.go`
- [ ] `message_send_welcome_service.go`
- [ ] `message_send_broadcast_service.go`

### Backup:
- [ ] สำรองไฟล์ทั้งหมด

### Git Commit:
- [ ] `git add .`
- [ ] `git commit -m "Phase 9: Remove business service implementations (13 files)"`

**สถานะ:** ⬜ ยังไม่เริ่ม | 🟡 กำลังทำ | ✅ เสร็จสมบูรณ์
**เวลาที่ใช้:** ________ นาที

---

## 📦 PHASE 10: Remove Service Interfaces & Repositories

### ลบ Service Interfaces (10 files):
- [ ] `domain/service/business_account_service.go`
- [ ] `domain/service/business_admin_service.go`
- [ ] `domain/service/business_follow_service.go`
- [ ] `domain/service/business_welcome_message_service.go`
- [ ] `domain/service/broadcast_service.go`
- [ ] `domain/service/broadcast_delivery_service.go`
- [ ] `domain/service/customer_profile_service.go`
- [ ] `domain/service/tag_service.go`
- [ ] `domain/service/user_tag_service.go`
- [ ] `domain/service/analytics_service.go`

### ลบ Repository Interfaces (10 files):
- [ ] `domain/repository/business_account_repository.go`
- [ ] `domain/repository/business_admin_repository.go`
- [ ] `domain/repository/business_follow_repository.go`
- [ ] `domain/repository/business_welcome_message_repository.go`
- [ ] `domain/repository/broadcast_repository.go`
- [ ] `domain/repository/broadcast_delivery_repository.go`
- [ ] `domain/repository/customer_profile_repository.go`
- [ ] `domain/repository/tag_repository.go`
- [ ] `domain/repository/user_tag_repository.go`
- [ ] `domain/repository/analytics_daily_repository.go`

### ลบ Repository Implementations (10 files):
- [ ] `infrastructure/persistence/postgres/business_account_repository.go`
- [ ] `infrastructure/persistence/postgres/business_admin_repository.go`
- [ ] `infrastructure/persistence/postgres/business_follow_repository.go`
- [ ] `infrastructure/persistence/postgres/business_welcome_message_repository.go`
- [ ] `infrastructure/persistence/postgres/broadcast_repository.go`
- [ ] `infrastructure/persistence/postgres/broadcast_delivery_repository.go`
- [ ] `infrastructure/persistence/postgres/customer_profile_repository.go`
- [ ] `infrastructure/persistence/postgres/tag_repository.go`
- [ ] `infrastructure/persistence/postgres/user_tag_repository.go`
- [ ] `infrastructure/persistence/postgres/analytics_daily_repository.go`

### Backup:
- [ ] สำรองไฟล์ทั้งหมด

### Git Commit:
- [ ] `git add .`
- [ ] `git commit -m "Phase 10: Remove business service interfaces and repositories (30 files)"`

**สถานะ:** ⬜ ยังไม่เริ่ม | 🟡 กำลังทำ | ✅ เสร็จสมบูรณ์
**เวลาที่ใช้:** ________ นาที

---

## 📦 PHASE 11: Remove Business Models

### ลบไฟล์ (13 files):
- [ ] `domain/models/business_account.go`
- [ ] `domain/models/business_admin.go`
- [ ] `domain/models/business_welcome_message.go`
- [ ] `domain/models/broadcast.go`
- [ ] `domain/models/broadcast_delivery.go`
- [ ] `domain/models/customer_profile.go`
- [ ] `domain/models/tag.go`
- [ ] `domain/models/user_tag.go`
- [ ] `domain/models/user_business_follow.go`
- [ ] `domain/models/analytics_daily.go`
- [ ] `domain/models/rich_menu.go`
- [ ] `domain/models/rich_menu_area.go`
- [ ] `domain/models/user_rich_menu.go`

### Backup:
- [ ] สำรองไฟล์ทั้งหมด

### Git Commit:
- [ ] `git add .`
- [ ] `git commit -m "Phase 11: Remove business models (13 files)"`

**สถานะ:** ⬜ ยังไม่เริ่ม | 🟡 กำลังทำ | ✅ เสร็จสมบูรณ์
**เวลาที่ใช้:** ________ นาที

---

## 📦 PHASE 12: Update Infrastructure ⚠️ HIGH RISK

### 12.1 DI Container (`pkg/di/container.go`):
- [ ] ลบ Business Repository fields ทั้งหมด
- [ ] ลบ Business Service fields ทั้งหมด
- [ ] ลบ BroadcastScheduler field
- [ ] ลบ Business Handler fields ทั้งหมด
- [ ] ลบการสร้าง Business instances ใน `NewContainer()`
- [ ] แก้ไข ConversationService constructor (ลบ business repos)
- [ ] แก้ไข MessageService constructor (ลบ business repos)
- [ ] แก้ไข NotificationService constructor (ลบ business repo)
- [ ] แก้ไข WebSocketHub constructor (ลบ business admin service)

### 12.2 Main (`cmd/api/main.go`):
- [ ] ลบ BroadcastScheduler.LoadScheduledBroadcasts()
- [ ] ลบ BroadcastScheduler.Start()
- [ ] ลบ BroadcastScheduler.Stop()

### 12.3 Routes Setup (`interfaces/api/routes/routes.go`):
- [ ] ลบ Business Handler parameters
- [ ] ลบ BusinessAdminService parameter
- [ ] ลบ Business Route Setup calls ทั้งหมด

### 12.4 App Setup (`pkg/app/app.go`):
- [ ] ลบ Business Handlers จาก routes.SetupRoutes()
- [ ] ลบ BusinessAdminService จาก routes.SetupRoutes()

### 12.5 Migration (`infrastructure/persistence/database/migration.go`):
- [ ] ลบ Business Models จาก AutoMigrate
- [ ] ลบ Custom Indices สำหรับ Business

### Compile Test:
- [ ] รัน: `go build ./...`
- [ ] ตรวจสอบ errors และแก้ไข

### Git Commit:
- [ ] `git add pkg/di/container.go cmd/api/main.go interfaces/api/routes/routes.go pkg/app/app.go infrastructure/persistence/database/migration.go`
- [ ] `git commit -m "Phase 12: Update infrastructure (DI, Main, Routes, Migration)"`

**สถานะ:** ⬜ ยังไม่เริ่ม | 🟡 กำลังทำ | ✅ เสร็จสมบูรณ์
**เวลาที่ใช้:** ________ นาที
**ปัญหาที่พบ:**
```


```

---

## 📦 PHASE 13: Final Cleanup & Testing

### Cleanup:
- [ ] รัน: `goimports -w .` (ลบ unused imports)
- [ ] รัน: `go mod tidy` (ลบ unused dependencies)
- [ ] รัน: `go list -m all > refector_plan/dependencies_after.txt`
- [ ] เปรียบเทียบ dependencies: `diff refector_plan/dependencies_before.txt refector_plan/dependencies_after.txt`

### Build:
- [ ] รัน: `go build -o chat-backend ./cmd/api`
- [ ] Build สำเร็จไม่มี errors

### Database:
- [ ] Start containers: `docker-compose up -d postgres redis`
- [ ] Run migration: `go run cmd/api/main.go migrate`
- [ ] Migration สำเร็จ

### Run Application:
- [ ] Start app: `go run cmd/api/main.go`
- [ ] Application รันได้ไม่มี errors

### API Testing:

#### Authentication:
- [ ] Register: `POST /api/v1/auth/register`
- [ ] Login: `POST /api/v1/auth/login`
- [ ] Refresh Token: `POST /api/v1/auth/refresh`
- [ ] Logout: `POST /api/v1/auth/logout`

#### User:
- [ ] Get Profile: `GET /api/v1/users/me`
- [ ] Update Profile: `PATCH /api/v1/users/me`
- [ ] Upload Profile Image: `PUT /api/v1/users/me/profile-image`

#### Friendship:
- [ ] Send Friend Request: `POST /api/v1/friendships/request`
- [ ] Accept Friend: `POST /api/v1/friendships/:id/accept`
- [ ] Get Friends: `GET /api/v1/friendships`

#### Conversations:
- [ ] Create Conversation: `POST /api/v1/conversations`
- [ ] Get Conversations: `GET /api/v1/conversations`
- [ ] Get Conversation Details: `GET /api/v1/conversations/:id`

#### Messages:
- [ ] Send Message: `POST /api/v1/conversations/:id/messages`
- [ ] Get Messages: `GET /api/v1/conversations/:id/messages`
- [ ] Edit Message: `PATCH /api/v1/messages/:id`
- [ ] Delete Message: `DELETE /api/v1/messages/:id`

#### WebSocket:
- [ ] Connect: `ws://localhost:8080/ws?token=<TOKEN>`
- [ ] Send message via WebSocket
- [ ] Receive message via WebSocket

### Database Verification:
- [ ] เชื่อมต่อ PostgreSQL: `psql -U postgres -d chatbiz_db`
- [ ] ตรวจสอบ tables: `\dt`
- [ ] ไม่มี business_* tables
- [ ] Tables ที่เหลือ: users, conversations, messages, user_friendships, stickers, etc.

### Git Commit:
- [ ] `git add .`
- [ ] `git commit -m "Phase 13: Final cleanup and testing completed"`

**สถานะ:** ⬜ ยังไม่เริ่ม | 🟡 กำลังทำ | ✅ เสร็จสมบูรณ์
**เวลาที่ใช้:** ________ นาที

---

## ✅ Final Verification Checklist

### Files Count:
- [ ] นับไฟล์ Go ก่อน refactor: `find . -name "*.go" | wc -l` = ~203
- [ ] นับไฟล์ Go หลัง refactor: `find . -name "*.go" | wc -l` = ~140
- [ ] ลดลงประมาณ 30%

### No Business References:
- [ ] ค้นหา BusinessAccount: `grep -r "BusinessAccount" --include="*.go" .` (ไม่มีผลลัพธ์นอก backup)
- [ ] ค้นหา Broadcast: `grep -r "Broadcast" --include="*.go" .` (ไม่มีผลลัพธ์นอก backup)
- [ ] ค้นหา CustomerProfile: `grep -r "CustomerProfile" --include="*.go" .` (ไม่มีผลลัพธ์นอก backup)

### Functionality:
- [ ] ✅ Authentication ทำงานได้
- [ ] ✅ User Profile ทำงานได้
- [ ] ✅ Friendship ทำงานได้
- [ ] ✅ Direct Messaging ทำงานได้
- [ ] ✅ Group Chat ทำงานได้
- [ ] ✅ Message Edit/Delete ทำงานได้
- [ ] ✅ File Upload ทำงานได้
- [ ] ✅ Stickers ทำงานได้
- [ ] ✅ WebSocket ทำงานได้
- [ ] ✅ Search ทำงานได้

### No Business Features:
- [ ] ❌ ไม่มี /businesses/* endpoints
- [ ] ❌ ไม่มี /broadcasts/* endpoints
- [ ] ❌ ไม่มี business admin roles
- [ ] ❌ ไม่มี CRM features
- [ ] ❌ ไม่มี analytics

---

## 📊 สถิติสุดท้าย

### เวลาที่ใช้ทั้งหมด:
- Phase 0: ________ นาที
- Phase 1: ________ นาที
- Phase 2: ________ นาที
- Phase 3: ________ นาที
- Phase 4: ________ นาที
- Phase 5: ________ นาที
- Phase 6: ________ นาที
- Phase 7: ________ นาที
- Phase 8: ________ นาที
- Phase 9: ________ นาที
- Phase 10: ________ นาที
- Phase 11: ________ นาที
- Phase 12: ________ นาที
- Phase 13: ________ นาที
- **รวม:** ________ นาที (________ ชั่วโมง)

### ปัญหาที่พบและวิธีแก้:
```
1.

2.

3.

```

### บทเรียนที่ได้:
```
1.

2.

3.

```

---

## 🎉 ประกาศความสำเร็จ

- [ ] ✅ Refactor เสร็จสมบูรณ์ทั้งหมด
- [ ] ✅ ทดสอบทุกฟีเจอร์แล้ว
- [ ] ✅ ไม่มี business features เหลืออยู่
- [ ] ✅ Application รันได้ปกติ
- [ ] ✅ Merge เข้า main branch: `git checkout main && git merge refactor/remove-business-features`
- [ ] ✅ Push to remote: `git push origin main`
- [ ] ✅ ลบ working branch: `git branch -d refactor/remove-business-features`

**ลงชื่อ:** __________________
**วันที่เสร็จ:** __________________

---

**หมายเหตุ:** ถ้าพบปัญหาระหว่างทาง ให้กลับไปดูที่ `MASTER_REFACTOR_PLAN.md` สำหรับรายละเอียดเพิ่มเติม
