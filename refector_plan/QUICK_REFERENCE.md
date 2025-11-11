# 🚀 QUICK REFERENCE GUIDE

คู่มืออ้างอิงด่วนสำหรับการ Refactor

---

## 📚 ไฟล์เอกสารทั้งหมด

1. **`result_system.md`** - รายงานการวิเคราะห์ระบบทั้งหมด
2. **`MASTER_REFACTOR_PLAN.md`** - แผนหลักพร้อมรายละเอียดทุก Phase
3. **`CHECKLIST.md`** - เช็คลิสต์สำหรับติดตามความคืบหน้า (ไฟล์นี้)
4. **`QUICK_REFERENCE.md`** - คู่มืออ้างอิงด่วน (ไฟล์นี้)

---

## ⚡ คำสั่งที่ใช้บ่อย

### Git Commands

```bash
# สร้าง backup
git commit -m "Pre-refactor: Save current state"
git branch backup-before-refactor
git checkout -b refactor/remove-business-features
git tag pre-refactor-backup

# Commit หลังแต่ละ Phase
git add .
git commit -m "Phase X: Description"

# ดูประวัติ
git log --oneline

# ดูการเปลี่ยนแปลง
git diff
git diff HEAD~1  # เปรียบเทียบกับ commit ก่อนหน้า

# Rollback (ถ้าเกิดปัญหา)
git checkout backup-before-refactor
git checkout pre-refactor-backup

# Rollback specific file
git checkout backup-before-refactor -- path/to/file.go
```

### Database Commands

```bash
# Backup database
pg_dump -U postgres -d chatbiz_db > backup_$(date +%Y%m%d_%H%M%S).sql

# Restore database
psql -U postgres -d chatbiz_db < backup_file.sql

# Connect to database
psql -U postgres -d chatbiz_db

# List tables
\dt

# Describe table
\d table_name

# Drop table (ระวัง!)
DROP TABLE table_name CASCADE;
```

### Go Commands

```bash
# Build
go build -o chat-backend ./cmd/api

# Build specific package
go build ./domain/models/...
go build ./application/serviceimpl/...

# Run
go run cmd/api/main.go

# Test compile without building
go build -o /dev/null ./...

# Clean cache
go clean -cache

# Format code
go fmt ./...
gofmt -w .

# Remove unused imports
goimports -w .

# Tidy dependencies
go mod tidy

# List dependencies
go list -m all

# Check for errors
go vet ./...

# Download goimports
go install golang.org/x/tools/cmd/goimports@latest
```

### Search Commands

```bash
# ค้นหา BusinessAccount references
grep -r "BusinessAccount" --include="*.go" .

# ค้นหาและนับ
grep -r "BusinessAccount" --include="*.go" . | wc -l

# ค้นหาเฉพาะชื่อไฟล์
find . -name "*business*.go"

# ค้นหา imports
grep -r "\".*business.*\"" --include="*.go" .

# นับไฟล์ Go
find . -name "*.go" | wc -l

# นับไฟล์ใน directory
ls -1 interfaces/api/routes/ | wc -l
```

### Docker Commands

```bash
# Start services
docker-compose up -d

# Start specific service
docker-compose up -d postgres
docker-compose up -d redis

# Stop services
docker-compose down

# View logs
docker-compose logs -f postgres

# Remove volumes (ระวัง! ลบข้อมูล)
docker-compose down -v
```

---

## 📁 โครงสร้างไฟล์ที่สำคัญ

```
chat-backend-v2-main/
├── cmd/api/
│   └── main.go                          ⚠️ ต้องแก้ไข (ลบ Scheduler)
│
├── domain/
│   ├── models/
│   │   ├── user.go                      ⚠️ ต้องแก้ไข (ลบ Business relations)
│   │   ├── conversation.go              ⚠️ ต้องแก้ไข (ลบ BusinessID)
│   │   ├── message.go                   ⚠️ ต้องแก้ไข (ลบ BusinessID)
│   │   └── business_*.go                ❌ ต้องลบทั้งหมด
│   │
│   ├── service/
│   │   ├── conversation_service.go      ⚠️ ต้องแก้ไข
│   │   ├── message_service.go           ⚠️ ต้องแก้ไข
│   │   ├── notification_service.go      ⚠️ ต้องแก้ไข
│   │   └── business_*.go                ❌ ต้องลบทั้งหมด
│   │
│   ├── repository/
│   │   └── business_*.go                ❌ ต้องลบทั้งหมด
│   │
│   └── dto/
│       └── business_*.go                ❌ ต้องลบทั้งหมด
│
├── application/serviceimpl/
│   ├── conversations_service.go         ⚠️ ต้องแก้ไข
│   ├── message_service.go               ⚠️ ต้องแก้ไข
│   ├── notification_service.go          ⚠️ ต้องแก้ไข
│   └── business_*.go                    ❌ ต้องลบทั้งหมด
│
├── infrastructure/
│   ├── persistence/
│   │   ├── database/
│   │   │   └── migration.go             ⚠️ ต้องแก้ไข
│   │   └── postgres/
│   │       └── business_*.go            ❌ ต้องลบทั้งหมด
│   │
│   └── adapter/
│       └── websocket_adapter.go         ⚠️ อาจต้องแก้ไข
│
├── interfaces/
│   ├── api/
│   │   ├── handler/
│   │   │   └── business_*.go            ❌ ต้องลบทั้งหมด
│   │   │
│   │   ├── routes/
│   │   │   ├── routes.go                ⚠️ ต้องแก้ไข
│   │   │   └── business_*.go            ❌ ต้องลบทั้งหมด
│   │   │
│   │   └── middleware/
│   │       └── business_admin.go        ❌ ต้องลบ
│   │
│   └── websocket/
│       ├── hub.go                       ⚠️ ต้องแก้ไข
│       ├── handlers.go                  ⚠️ ต้องแก้ไข
│       └── broadcast.go                 ⚠️ ต้องแก้ไข
│
├── pkg/
│   ├── di/
│   │   └── container.go                 ⚠️ ต้องแก้ไข (ลบ Business DI)
│   │
│   └── app/
│       └── app.go                       ⚠️ ต้องแก้ไข
│
└── scheduler/
    └── broadcast_scheduler.go           ❌ ต้องลบ
```

---

## 🔍 ตรวจสอบว่าลบครบหรือยัง

### Business Models (13 files):
```bash
ls domain/models/ | grep -E "(business|broadcast|tag|customer|analytics|rich_menu)"
```
**ต้องไม่มีผลลัพธ์**

### Business Routes (12 files):
```bash
ls interfaces/api/routes/ | grep -E "(business|broadcast|tag|customer|analytics)"
```
**ต้องไม่มีผลลัพธ์**

### Business Handlers (10 files):
```bash
ls interfaces/api/handler/ | grep -E "(business|broadcast|tag|customer|analytics)"
```
**ต้องไม่มีผลลัพธ์**

### Business Services (13 files):
```bash
ls application/serviceimpl/ | grep -E "(business|broadcast|tag|customer|analytics)"
```
**ต้องไม่มีผลลัพธ์**

### Business Repositories (20 files):
```bash
ls domain/repository/ | grep -E "(business|broadcast|tag|customer|analytics)"
ls infrastructure/persistence/postgres/ | grep -E "(business|broadcast|tag|customer|analytics)"
```
**ต้องไม่มีผลลัพธ์**

### ไม่มี BusinessAccount references:
```bash
grep -r "BusinessAccount" --include="*.go" --exclude-dir=refector_plan .
```
**ควรไม่มีผลลัพธ์ (นอกเหนือจากใน backup)**

---

## 🧪 ทดสอบ API ด้วย curl

### Authentication:
```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "test123",
    "display_name": "Test User"
  }'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "test123"
  }'

# บันทึก token
export TOKEN="<YOUR_ACCESS_TOKEN>"

# Logout
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer $TOKEN"
```

### User Profile:
```bash
# Get own profile
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $TOKEN"

# Update profile
curl -X PATCH http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "Updated Name",
    "bio": "My new bio"
  }'
```

### Friendship:
```bash
# Send friend request
curl -X POST http://localhost:8080/api/v1/friendships/request \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "friend_id": "<FRIEND_USER_ID>"
  }'

# Accept friend request
curl -X POST http://localhost:8080/api/v1/friendships/<FRIENDSHIP_ID>/accept \
  -H "Authorization: Bearer $TOKEN"

# Get friends
curl -X GET http://localhost:8080/api/v1/friendships \
  -H "Authorization: Bearer $TOKEN"
```

### Conversations:
```bash
# Create conversation
curl -X POST http://localhost:8080/api/v1/conversations \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "private",
    "member_ids": ["<FRIEND_USER_ID>"]
  }'

# Get conversations
curl -X GET http://localhost:8080/api/v1/conversations \
  -H "Authorization: Bearer $TOKEN"

# Get specific conversation
curl -X GET http://localhost:8080/api/v1/conversations/<CONVERSATION_ID> \
  -H "Authorization: Bearer $TOKEN"
```

### Messages:
```bash
# Send text message
curl -X POST http://localhost:8080/api/v1/conversations/<CONVERSATION_ID>/messages \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "text",
    "content": "Hello, this is a test message!"
  }'

# Get messages
curl -X GET http://localhost:8080/api/v1/conversations/<CONVERSATION_ID>/messages \
  -H "Authorization: Bearer $TOKEN"

# Edit message
curl -X PATCH http://localhost:8080/api/v1/messages/<MESSAGE_ID> \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Updated message content"
  }'

# Delete message
curl -X DELETE http://localhost:8080/api/v1/messages/<MESSAGE_ID> \
  -H "Authorization: Bearer $TOKEN"
```

### File Upload:
```bash
# Upload image
curl -X POST http://localhost:8080/api/v1/files/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@/path/to/image.jpg" \
  -F "type=image"
```

---

## ❌ Endpoints ที่ต้องไม่มีหลัง Refactor

ทดสอบว่า Business endpoints ไม่ทำงานแล้ว (ควรได้ 404):

```bash
# ทั้งหมดนี้ควรได้ 404 Not Found
curl -X GET http://localhost:8080/api/v1/businesses
curl -X GET http://localhost:8080/api/v1/businesses/<ID>/broadcasts
curl -X GET http://localhost:8080/api/v1/businesses/<ID>/customers
curl -X GET http://localhost:8080/api/v1/businesses/<ID>/analytics
```

---

## 🐛 แก้ไขปัญหาที่พบบ่อย

### Compile Error: undefined reference

**ปัญหา:**
```
undefined: repository.BusinessAccountRepository
```

**วิธีแก้:**
1. ค้นหาไฟล์ที่ยังใช้ reference นี้อยู่
   ```bash
   grep -r "BusinessAccountRepository" --include="*.go" .
   ```
2. แก้ไขหรือลบ import นั้นออก

---

### Compile Error: missing argument

**ปัญหา:**
```
not enough arguments in call to serviceimpl.NewConversationService
```

**วิธีแก้:**
1. เช็ค constructor signature ใน implementation
2. เช็ค DI container ว่าส่ง parameters ครบหรือไม่
3. แก้ไข DI container ให้ตรงกับ constructor ใหม่

---

### Migration Error: table does not exist

**ปัญหา:**
```
ERROR: table "business_accounts" does not exist
```

**วิธีแก้:**
1. ตรวจสอบว่าลบ Business models ออกจาก migration แล้วหรือยัง
2. ลบ migration schema_migrations ใน database
3. รัน migration ใหม่

---

### Runtime Error: nil pointer dereference

**ปัญหา:**
```
panic: runtime error: invalid memory address or nil pointer dereference
```

**วิธีแก้:**
1. เช็คว่า DI container inject dependencies ครบหรือไม่
2. เช็คว่า service constructors รับ parameters ครบหรือไม่
3. เช็ค logs เพื่อหา stack trace

---

## 📊 ตัวเลขที่ควรได้หลัง Refactor

### จำนวนไฟล์:
```bash
# ก่อน refactor
find . -name "*.go" | wc -l
# ควรได้ ~203

# หลัง refactor
find . -name "*.go" | wc -l
# ควรได้ ~140 (-30%)
```

### จำนวน Models:
```bash
# ก่อน refactor
ls domain/models/*.go | wc -l
# ควรได้ ~29

# หลัง refactor
ls domain/models/*.go | wc -l
# ควรได้ ~16 (-13 models)
```

### Database Tables:
```sql
-- ก่อน refactor
SELECT COUNT(*) FROM information_schema.tables
WHERE table_schema = 'public';
-- ควรได้ ~29

-- หลัง refactor
SELECT COUNT(*) FROM information_schema.tables
WHERE table_schema = 'public';
-- ควรได้ ~16 (-13 tables)
```

---

## 🎯 Regular User Features ที่ต้องทำงานได้

- ✅ Register & Login
- ✅ User Profile (View/Edit)
- ✅ Upload Profile Image
- ✅ Add/Remove Friends
- ✅ Create Private Chat
- ✅ Create Group Chat
- ✅ Send Text Messages
- ✅ Send Image Messages
- ✅ Edit Messages
- ✅ Delete Messages
- ✅ View Message History
- ✅ Read Status
- ✅ Send Stickers
- ✅ Real-time via WebSocket
- ✅ Search Users

---

## 🚫 Business Features ที่ต้องไม่มี

- ❌ Create Business Account
- ❌ Business Admin Management
- ❌ Follow/Unfollow Business
- ❌ Broadcast Messages
- ❌ Customer CRM
- ❌ Customer Tagging
- ❌ Welcome Messages
- ❌ Business Analytics
- ❌ Rich Menu
- ❌ Scheduled Broadcasts

---

## 📞 Emergency Rollback

ถ้าเกิดปัญหาร้ายแรงและต้อง rollback ทันที:

```bash
# 1. Stop application
# Ctrl+C

# 2. Rollback code
git checkout backup-before-refactor

# 3. Restore database (ถ้าจำเป็น)
psql -U postgres -d chatbiz_db < backup_before_refactor.sql

# 4. Restart application
go run cmd/api/main.go
```

---

## 📝 Checklist ก่อนเริ่มแต่ละ Phase

- [ ] อ่านรายละเอียด Phase ใน MASTER_REFACTOR_PLAN.md แล้ว
- [ ] Commit code ปัจจุบัน
- [ ] ไม่มี uncommitted changes
- [ ] รู้ว่าจะลบ/แก้ไขไฟล์อะไรบ้าง
- [ ] เตรียม rollback plan ถ้าเกิดปัญหา

## 📝 Checklist หลังเสร็จแต่ละ Phase

- [ ] ลบ/แก้ไขไฟล์ครบตามแผนแล้ว
- [ ] ทดสอบ compile (go build)
- [ ] ตรวจสอบผลลัพธ์ (verification steps)
- [ ] Commit พร้อม message ที่ชัดเจน
- [ ] อัปเดต CHECKLIST.md

---

**สร้างโดย:** Claude Code Assistant
**วันที่:** 2025-11-12
**Version:** 1.0.0
