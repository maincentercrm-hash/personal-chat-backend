// infrastructure/storage/r2/r2_config.go
package r2

// R2Config เก็บการตั้งค่าสำหรับ Cloudflare R2
type R2Config struct {
	AccountID       string // R2 Account ID
	AccessKeyID     string // R2 Access Key ID
	SecretAccessKey string // R2 Secret Access Key
	Bucket          string // R2 Bucket name
	PublicURL       string // Public URL สำหรับเข้าถึงไฟล์ (https://pub-xxx.r2.dev)
	Region          string // Region (default: auto)
	Endpoint        string // Custom endpoint (optional)
}

// GetEndpoint สร้าง endpoint URL สำหรับ R2
func (c *R2Config) GetEndpoint() string {
	if c.Endpoint != "" {
		return c.Endpoint
	}
	// R2 endpoint format: https://<account_id>.r2.cloudflarestorage.com
	return "https://" + c.AccountID + ".r2.cloudflarestorage.com"
}

// GetRegion คืนค่า region (default: auto)
func (c *R2Config) GetRegion() string {
	if c.Region != "" {
		return c.Region
	}
	return "auto"
}
