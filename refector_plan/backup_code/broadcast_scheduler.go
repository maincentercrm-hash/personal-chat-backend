// pkg/scheduler/broadcast_scheduler.go
package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
	"github.com/thizplus/gofiber-chat-api/domain/repository"
	"github.com/thizplus/gofiber-chat-api/domain/service"
)

const (
	scheduledBroadcastsKey = "scheduled_broadcasts"
	processingLockKey      = "processing_broadcasts_lock"
	lockDuration           = 30 * time.Second
)

// BroadcastScheduler จัดการการส่ง broadcasts ตามเวลาที่กำหนด
type BroadcastScheduler struct {
	ctx               context.Context
	rdb               *redis.Client
	broadcastRepo     repository.BroadcastRepository
	broadcastService  service.BroadcastService
	checkInterval     time.Duration
	processingChannel chan string
	workerCount       int
	isRunning         bool
	stopChan          chan struct{}
}

// NewBroadcastScheduler สร้าง instance ใหม่ของ BroadcastScheduler
func NewBroadcastScheduler(
	rdb *redis.Client,
	broadcastRepo repository.BroadcastRepository,
	broadcastService service.BroadcastService,
) *BroadcastScheduler {
	return &BroadcastScheduler{
		ctx:               context.Background(),
		rdb:               rdb,
		broadcastRepo:     broadcastRepo,
		broadcastService:  broadcastService,
		checkInterval:     1 * time.Second,
		processingChannel: make(chan string, 100),
		workerCount:       5, // จำนวน workers ที่จะประมวลผล broadcasts พร้อมกัน
		isRunning:         false,
		stopChan:          make(chan struct{}),
	}
}

// Start เริ่มการทำงานของ scheduler
func (s *BroadcastScheduler) Start() error {
	if s.isRunning {
		return fmt.Errorf("scheduler is already running")
	}

	// เริ่ม workers
	for i := 0; i < s.workerCount; i++ {
		go s.worker()
	}

	// เริ่ม scheduler
	go s.schedulerLoop()

	s.isRunning = true
	log.Println("Broadcast scheduler started")
	return nil
}

// Stop หยุดการทำงานของ scheduler
func (s *BroadcastScheduler) Stop() error {
	if !s.isRunning {
		return fmt.Errorf("scheduler is not running")
	}

	close(s.stopChan)
	close(s.processingChannel)
	s.isRunning = false
	log.Println("Broadcast scheduler stopped")
	return nil
}

// schedulerLoop ทำงานเป็นรอบๆ เพื่อตรวจสอบและประมวลผล broadcasts ที่ถึงเวลากำหนด
func (s *BroadcastScheduler) schedulerLoop() {
	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.processDueBroadcasts()
		case <-s.stopChan:
			return
		}
	}
}

// processDueBroadcasts ตรวจสอบและประมวลผล broadcasts ที่ถึงเวลากำหนด
func (s *BroadcastScheduler) processDueBroadcasts() {
	// พยายามล็อคเพื่อป้องกันการประมวลผลซ้ำ
	lockSuccess, err := s.rdb.SetNX(s.ctx, processingLockKey, "1", lockDuration).Result()
	if err != nil || !lockSuccess {
		return
	}
	defer s.rdb.Del(s.ctx, processingLockKey)

	// ดึง broadcasts ที่ถึงเวลากำหนด
	now := float64(time.Now().Unix())
	broadcasts, err := s.rdb.ZRangeByScore(s.ctx, scheduledBroadcastsKey, &redis.ZRangeBy{
		Min: "0",
		Max: fmt.Sprintf("%f", now),
	}).Result()

	if err != nil {
		log.Printf("Error fetching due broadcasts: %v", err)
		return
	}

	if len(broadcasts) == 0 {
		return
	}

	// ลบ broadcasts ที่ถึงเวลากำหนดออกจาก sorted set
	_, err = s.rdb.ZRemRangeByScore(s.ctx, scheduledBroadcastsKey, "0", fmt.Sprintf("%f", now)).Result()
	if err != nil {
		log.Printf("Error removing due broadcasts: %v", err)
	}

	// ส่ง broadcasts ไปยัง processing channel
	for _, broadcastJSON := range broadcasts {
		select {
		case s.processingChannel <- broadcastJSON:
			// ส่งสำเร็จ
		default:
			// channel เต็ม ให้ลองอีกครั้งในรอบถัดไป
			log.Printf("Processing channel is full, retrying broadcast later")
			// เพิ่มกลับเข้า sorted set ด้วยเวลาในอีก 10 วินาที
			s.rdb.ZAdd(s.ctx, scheduledBroadcastsKey, &redis.Z{
				Score:  now + 10,
				Member: broadcastJSON,
			})
		}
	}
}

// worker ทำงานเพื่อประมวลผล broadcasts ที่ถึงเวลากำหนด
func (s *BroadcastScheduler) worker() {
	for broadcastJSON := range s.processingChannel {
		var broadcastData map[string]string
		if err := json.Unmarshal([]byte(broadcastJSON), &broadcastData); err != nil {
			log.Printf("Error unmarshaling broadcast data: %v", err)
			continue
		}

		broadcastID, ok := broadcastData["id"]
		if !ok {
			log.Printf("Invalid broadcast data: missing id")
			continue
		}

		// แปลง string เป็น UUID
		id, err := uuid.Parse(broadcastID)
		if err != nil {
			log.Printf("Invalid broadcast ID: %v", err)
			continue
		}

		// ดึงข้อมูล broadcast จากฐานข้อมูล
		broadcast, err := s.broadcastRepo.GetByID(id)
		if err != nil {
			log.Printf("Error fetching broadcast: %v", err)
			continue
		}

		// ตรวจสอบสถานะ
		if broadcast.Status != "scheduled" {
			log.Printf("Broadcast %s is not in scheduled status", id)
			continue
		}

		// ดึงข้อมูล business และผู้สร้าง broadcast เพื่อใช้ในการส่ง
		businessID := broadcast.BusinessID
		createdByID := uuid.Nil
		if broadcast.CreatedBy != nil {
			createdByID = *broadcast.CreatedBy
		}

		// ส่ง broadcast โดยใช้ BroadcastService
		err = s.broadcastService.SendBroadcast(id, businessID, createdByID)
		if err != nil {
			log.Printf("Error sending broadcast %s: %v", id, err)
		} else {
			log.Printf("Successfully sent broadcast %s", id)
		}
	}
}

// ScheduleBroadcast เพิ่ม broadcast ลงใน scheduler
func (s *BroadcastScheduler) ScheduleBroadcast(broadcast *models.Broadcast) error {
	if broadcast.ScheduledAt == nil {
		return fmt.Errorf("broadcast has no scheduled time")
	}

	// แปลงข้อมูล broadcast เป็น JSON
	broadcastData := map[string]string{
		"id": broadcast.ID.String(),
	}
	broadcastJSON, err := json.Marshal(broadcastData)
	if err != nil {
		return err
	}

	// เพิ่ม broadcast ลงใน sorted set
	score := float64(broadcast.ScheduledAt.Unix())
	_, err = s.rdb.ZAdd(s.ctx, scheduledBroadcastsKey, &redis.Z{
		Score:  score,
		Member: string(broadcastJSON),
	}).Result()

	return err
}

// CancelScheduledBroadcast ยกเลิก broadcast ที่กำหนดเวลาไว้
func (s *BroadcastScheduler) CancelScheduledBroadcast(broadcastID uuid.UUID) error {
	// ค้นหาและลบ broadcast ออกจาก sorted set
	broadcastData := map[string]string{
		"id": broadcastID.String(),
	}
	broadcastJSON, err := json.Marshal(broadcastData)
	if err != nil {
		return err
	}

	removed, err := s.rdb.ZRem(s.ctx, scheduledBroadcastsKey, string(broadcastJSON)).Result()
	if err != nil {
		return err
	}

	if removed == 0 {
		return fmt.Errorf("broadcast not found in scheduler")
	}

	return nil
}

// LoadScheduledBroadcasts โหลด broadcasts ที่มีการกำหนดเวลาไว้จากฐานข้อมูลเข้าสู่ Redis
func (s *BroadcastScheduler) LoadScheduledBroadcasts() error {
	// ดึง broadcasts ที่มีสถานะเป็น scheduled จากฐานข้อมูล
	currentTime := time.Now()
	broadcasts, err := s.broadcastRepo.GetScheduledBroadcasts(currentTime)
	if err != nil {
		return err
	}

	// เพิ่ม broadcasts ลงใน Redis
	for _, broadcast := range broadcasts {
		if broadcast.ScheduledAt != nil {
			err := s.ScheduleBroadcast(broadcast)
			if err != nil {
				log.Printf("Error scheduling broadcast %s: %v", broadcast.ID, err)
			}
		}
	}

	log.Printf("Loaded %d scheduled broadcasts", len(broadcasts))
	return nil
}
