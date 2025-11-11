# ChatBiz Platform - System Workflow

**ชื่อระบบ:** ChatBiz Platform (Enterprise Chat & Business Communication)
**ประเภท:** Multi-tenant Chat Backend with CRM & Broadcasting

## 🏗️ System Architecture Overview

```
Frontend Apps
     ↕️
API Gateway (Fiber v2)
     ↕️
Service Layer (Clean Architecture)
     ↕️
PostgreSQL + Redis + WebSocket Hub
     ↕️
External: Cloudinary (Media Storage)
```

**Tech Stack:** Go + Fiber + PostgreSQL + Redis + WebSocket + JWT

---

## 👥 User Roles & Permissions

### 1. **Regular User** 👤
- สร้างบัญชีส่วนตัว
- แชทกับเพื่อน (1-to-1)
- สร้าง/เข้าร่วมกลุ่มแชท
- ติดตาม Business Account
- อัพโหลดไฟล์/รูปภาพ

### 2. **Business Owner** 🏢
- สร้าง Business Account
- จัดการโปรไฟล์ธุรกิจ
- เพิ่ม Business Admin
- ส่ง Broadcast ข้อความ
- ดู Analytics
- จัดการข้อมูลลูกค้า (CRM)

### 3. **Business Admin** 👨‍💼
- จัดการ Business Account ที่ได้รับมอบหมาย
- ส่ง Broadcast ข้อความ
- แท็กและจัดกลุ่มลูกค้า
- ดู Analytics
- ตอบกลับข้อความลูกค้า

### 4. **Customer** 👥
- ติดตาม Business
- รับ Broadcast ข้อความ
- แชทกับ Business
- มีโปรไฟล์ในระบบ CRM

---

## 📋 Main Workflows

### 🔐 1. Authentication Flow
```
Register → Validate → Hash Password → Create User → Generate JWT
Login → Verify Password → Generate Tokens → Store Refresh Token
```

### 💬 2. Direct Messaging Flow
```
User A → Send Message → Store in DB → WebSocket Notify → User B
User B → Read Message → Update Read Status → Notify User A
```

### 🏢 3. Business Account Flow
```
Create Business → Set Profile → Add Admins → Followers Join
Welcome Message → Customer Profile Creation → CRM Management
```

### 📢 4. Broadcast Campaign Flow
```
Create Broadcast → Select Target Audience → Set Schedule
↓
Execute Send → Create Delivery Records → Track Opens/Clicks
↓
Analytics Update → Performance Metrics
```

### 🔄 5. Real-time Communication Flow
```
WebSocket Connect → Subscribe to Events → Real-time Updates
Message Sent → Hub Routing → Connected Clients Notified
```

---

## 🚀 Key Features by Role

### **Regular User Features**
- ✅ Registration & Login
- ✅ Friend Management
- ✅ Direct Messaging
- ✅ Group Chat
- ✅ Media Upload
- ✅ Business Following

### **Business Owner Features**
- ✅ Business Account Creation
- ✅ Admin Management
- ✅ Customer CRM
- ✅ Broadcast Campaigns
- ✅ Analytics Dashboard
- ✅ Welcome Messages
- ✅ Customer Segmentation

### **Business Admin Features**
- ✅ Assigned Business Management
- ✅ Customer Communication
- ✅ Broadcast Creation
- ✅ Customer Tagging
- ✅ Analytics Viewing

---

## 🔄 Core Business Processes

### **Customer Journey**
```
Discovery → Follow Business → Receive Welcome → Engage → Get Tagged → Receive Targeted Broadcasts
```

### **Business Communication Strategy**
```
Setup Business → Import/Create Customers → Segment Audience → Create Campaigns → Send Broadcasts → Analyze Results
```

### **Message Lifecycle**
```
Compose → Send → Deliver → Read → (Optional: Edit/Delete) → Archive
```

---

## 📊 Data Flow Examples

### **Send Message Process**
```
1. Client sends POST /api/v1/conversations/{id}/messages
2. MessageHandler validates & processes
3. MessageService stores to database
4. WebSocket Hub notifies recipients
5. Real-time delivery to connected clients
```

### **Broadcast Process**
```
1. Create broadcast via API
2. Store in database with target criteria
3. Redis scheduler queues for processing
4. Worker pool executes delivery
5. Track delivery status & analytics
```

### **Business Analytics Flow**
```
1. User interactions generate events
2. Analytics service aggregates data
3. Daily analytics calculated
4. Dashboard displays metrics
5. Export reports available
```

---

## 🌐 API Structure

### **Core Endpoints**
- `/auth/*` - Authentication
- `/users/*` - User management
- `/conversations/*` - Messaging
- `/businesses/*` - Business accounts
- `/broadcasts/*` - Campaign management
- `/ws/*` - WebSocket connections

### **Business-specific Endpoints**
- `/businesses/{id}/customers` - CRM
- `/businesses/{id}/broadcasts` - Campaigns
- `/businesses/{id}/analytics` - Metrics
- `/businesses/{id}/admins` - Admin management

---

## 🔧 Technical Components

### **Core Services**
- **AuthService** - JWT & user authentication
- **MessageService** - Chat functionality
- **BroadcastService** - Campaign management
- **NotificationService** - Real-time updates
- **BusinessService** - Business account management

### **Infrastructure**
- **PostgreSQL** - Primary database (29 models)
- **Redis** - Caching & job scheduling
- **WebSocket Hub** - Real-time communication
- **Cloudinary** - Media storage
- **Docker** - Containerization

---

## 📈 Scalability Features

- ✅ **Stateless API** design
- ✅ **Connection pooling** for database
- ✅ **Worker pools** for broadcast processing
- ✅ **Redis clustering** support
- ✅ **WebSocket horizontal scaling**
- ✅ **Rate limiting** per client

---

## 🔒 Security Measures

- ✅ **JWT Authentication** with refresh tokens
- ✅ **bcrypt Password** hashing
- ✅ **Role-based Access Control**
- ✅ **Token blacklisting** on logout
- ✅ **Rate limiting** for WebSocket
- ✅ **CORS protection**

---

## 📱 Use Cases Summary

**ChatBiz Platform** เป็นระบบแชทองค์กรที่รองรับ:
- การสื่อสารส่วนตัวและกลุ่ม
- การจัดการลูกค้าด้วยระบบ CRM
- การส่งข้อความแบบ Broadcasting
- การวิเคราะห์ผลลัพธ์และ Analytics
- การสื่อสารแบบ Real-time ผ่าน WebSocket

