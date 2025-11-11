// infrastructure/storage/cloudinary/cloudinary_config.go
package cloudinary

// CloudinaryConfig เก็บข้อมูลการตั้งค่า Cloudinary
type CloudinaryConfig struct {
	CloudName    string
	APIKey       string
	APISecret    string
	UploadFolder string
}
