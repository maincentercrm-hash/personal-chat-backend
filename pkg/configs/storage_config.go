// pkg/configs/storage_config.go
package configs

import (
	"fmt"
	"os"

	"github.com/thizplus/gofiber-chat-api/domain/service"
	"github.com/thizplus/gofiber-chat-api/infrastructure/storage/cloudinary"
)

// SetupStorageService สร้าง FileStorageService ตาม environment
func SetupStorageService() (service.FileStorageService, error) {
	// ถ้าในอนาคตต้องการเปลี่ยนมาใช้ S3 หรือ storage อื่น สามารถตรวจสอบจาก env ได้
	storageType := os.Getenv("STORAGE_TYPE")

	// ถ้าไม่ได้กำหนด ใช้ Cloudinary เป็นค่าเริ่มต้น
	if storageType == "" || storageType == "cloudinary" {
		return cloudinary.NewCloudinaryStorage(&cloudinary.CloudinaryConfig{
			CloudName:    os.Getenv("CLOUDINARY_CLOUD_NAME"),
			APIKey:       os.Getenv("CLOUDINARY_API_KEY"),
			APISecret:    os.Getenv("CLOUDINARY_API_SECRET"),
			UploadFolder: os.Getenv("CLOUDINARY_UPLOAD_FOLDER"),
		})
	}

	// ในอนาคตอาจเพิ่ม case อื่นๆ เช่น "s3" หรือ "local"
	// else if storageType == "s3" {
	//     return s3.NewS3Storage(&s3.S3Config{
	//         ...
	//     })
	// }

	// กรณีที่ระบุประเภทบริการจัดเก็บไฟล์ที่ไม่รองรับ
	return nil, fmt.Errorf("ไม่รองรับบริการจัดเก็บไฟล์ประเภท: %s", storageType)
}
