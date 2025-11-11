// pkg/configs/database.go

package configs

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database เป็นตัวแทนของการเชื่อมต่อฐานข้อมูล PostgreSQL ด้วย GORM
type Database struct {
	DB *gorm.DB
}

// NewDatabase สร้างและเปิดการเชื่อมต่อกับฐานข้อมูล PostgreSQL ด้วย GORM
func NewDatabase() (*Database, error) {
	// สร้าง DSN (Data Source Name) สำหรับ PostgreSQL
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSL_MODE"),
		os.Getenv("DB_TIMEZONE"),
	)

	// ตั้งค่า GORM logger
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level (ปรับเป็น Silent, Error, Warn, Info ตามต้องการ)
			IgnoreRecordNotFoundError: true,        // ไม่บันทึก error เมื่อไม่พบข้อมูล
			Colorful:                  true,        // ใช้สีในการแสดงผล
		},
	)

	// เชื่อมต่อกับฐานข้อมูล
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อกับฐานข้อมูลได้: %w", err)
	}

	// ตั้งค่า connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("ไม่สามารถรับ sql.DB จาก GORM ได้: %w", err)
	}

	// ตั้งค่าจำนวนการเชื่อมต่อสูงสุดและต่ำสุด
	sqlDB.SetMaxOpenConns(25)                 // จำนวนการเชื่อมต่อสูงสุดที่ active
	sqlDB.SetMaxIdleConns(10)                 // จำนวนการเชื่อมต่อที่ว่างสูงสุด
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // เวลาสูงสุดที่ใช้การเชื่อมต่อ

	return &Database{DB: db}, nil
}

// Close ปิดการเชื่อมต่อกับฐานข้อมูล
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("ไม่สามารถรับ sql.DB จาก GORM ได้: %w", err)
	}
	return sqlDB.Close()
}

// Ping ตรวจสอบการเชื่อมต่อกับฐานข้อมูล
func (d *Database) Ping() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("ไม่สามารถรับ sql.DB จาก GORM ได้: %w", err)
	}
	return sqlDB.Ping()
}

// AutoMigrate สร้างหรือปรับปรุงตารางตาม Model ที่ระบุ
func (d *Database) AutoMigrate(models ...interface{}) error {
	return d.DB.AutoMigrate(models...)
}
