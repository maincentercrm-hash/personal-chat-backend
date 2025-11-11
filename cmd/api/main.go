package main

import (
	"context" // เพิ่ม import นี้
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-redis/redis/v8" // เพิ่ม import นี้
	"github.com/joho/godotenv"
	db "github.com/thizplus/gofiber-chat-api/infrastructure/persistence/database"
	"github.com/thizplus/gofiber-chat-api/pkg/app"
	"github.com/thizplus/gofiber-chat-api/pkg/configs"
	"github.com/thizplus/gofiber-chat-api/pkg/di"
)

func main() {
	// โหลดไฟล์ .env
	if err := godotenv.Load(); err != nil {
		log.Println("ไม่พบไฟล์ .env, ใช้ค่า environment ที่มีอยู่")
	}

	// สร้างการเชื่อมต่อฐานข้อมูล
	database, err := configs.NewDatabase()
	if err != nil {
		log.Fatalf("ไม่สามารถเชื่อมต่อกับฐานข้อมูลได้: %v", err)
	}

	// ทำ migration ถ้าจำเป็น
	if err := db.SetupDatabase(database.DB); err != nil {
		log.Fatalf("การ migration ฐานข้อมูลล้มเหลว: %v", err)
	}

	// สร้าง storage service
	storageService, err := configs.SetupStorageService()
	if err != nil {
		log.Fatalf("StorageService error: %v", err)
	}

	// เชื่อมต่อกับ Redis
	redisConfig := configs.LoadRedisConfig()
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Host + ":" + redisConfig.Port,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})

	// ตรวจสอบการเชื่อมต่อกับ Redis
	ctx := context.Background()
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Error connecting to Redis: %v", err)
	}
	log.Println("Connected to Redis successfully")

	// สร้าง container โดยส่ง storageService และ redisClient เข้าไป
	container, err := di.NewContainer(database.DB, storageService, redisClient)
	if err != nil {
		log.Fatalf("ไม่สามารถสร้าง DI container ได้: %v", err)
	}

	// ลบโค้ดเริ่ม WebSocket Hub
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// go container.WebSocketHub.Run(ctx)
	// log.Println("WebSocket Hub started successfully")

	// สร้างและใช้ context สำหรับการจัดการ shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// เริ่ม WebSocket Hub (เพิ่มหลังสร้าง container)
	go container.WebSocketHub.Run(ctx)
	log.Println("WebSocket Hub started successfully")

	// โหลด scheduled broadcasts และเริ่ม scheduler
	err = container.BroadcastScheduler.LoadScheduledBroadcasts()
	if err != nil {
		log.Printf("Warning: Error loading scheduled broadcasts: %v", err)
	}

	err = container.BroadcastScheduler.Start()
	if err != nil {
		log.Printf("Warning: Error starting broadcast scheduler: %v", err)
	}

	// ตั้งค่าและสร้าง Fiber App
	app := app.SetupApp(container)

	// จัดการการปิดเซิร์ฟเวอร์อย่างสง่างาม
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "3000"
		}

		log.Printf("เซิร์ฟเวอร์กำลังทำงานที่พอร์ต %s", port)
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("ไม่สามารถเริ่มเซิร์ฟเวอร์ได้: %v", err)
		}
	}()

	/*
		// พิมพ์รายการเส้นทางทั้งหมด
		for _, route := range app.GetRoutes() {
			fmt.Printf("Method: %s, Path: %s\n", route.Method, route.Path)
		}
	*/

	<-c
	log.Println("กำลังปิดเซิร์ฟเวอร์...")

	// ลบส่วนปิด WebSocket Hub
	// cancel() // This will stop the WebSocket Hub

	// ยังคงให้ cancel เพื่อแจ้ง goroutines อื่นๆ ที่อาจใช้ context นี้
	cancel()

	// หยุด scheduler ก่อนปิดแอป
	if err := container.BroadcastScheduler.Stop(); err != nil {
		log.Printf("Error stopping scheduler: %v", err)
	}

	if err := app.Shutdown(); err != nil {
		log.Fatalf("ผิดพลาดในการปิดเซิร์ฟเวอร์: %v", err)
	}

	if err := database.Close(); err != nil {
		log.Fatalf("ผิดพลาดในการปิดการเชื่อมต่อฐานข้อมูล: %v", err)
	}

	log.Println("เซิร์ฟเวอร์ถูกปิดอย่างสง่างาม")
}
