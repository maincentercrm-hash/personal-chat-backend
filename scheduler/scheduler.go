// pkg/scheduler/scheduler.go
package scheduler

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

// Scheduler interface สำหรับทุกประเภทของ scheduler
type Scheduler interface {
	// Start เริ่มการทำงานของ scheduler
	Start() error

	// Stop หยุดการทำงานของ scheduler
	Stop() error
}

// BroadcastSchedulerInterface interface สำหรับ BroadcastScheduler
type BroadcastSchedulerInterface interface {
	Scheduler

	// ScheduleBroadcast เพิ่ม broadcast ลงใน scheduler
	ScheduleBroadcast(broadcast *models.Broadcast) error

	// CancelScheduledBroadcast ยกเลิก broadcast ที่กำหนดเวลาไว้
	CancelScheduledBroadcast(broadcastID uuid.UUID) error

	// LoadScheduledBroadcasts โหลด broadcasts ที่มีการกำหนดเวลาไว้จากฐานข้อมูลเข้าสู่ Redis
	LoadScheduledBroadcasts() error
}
