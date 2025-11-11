// domain/dto/sticker_dto.go
package dto

import (
	"mime/multipart"
	"time"
)

// =====================================
// Request DTOs
// =====================================

// CreateStickerSetRequest สำหรับการสร้างชุดสติกเกอร์ใหม่
type CreateStickerSetRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Author      string `json:"author"`
	IsOfficial  bool   `json:"is_official"`
	IsDefault   bool   `json:"is_default"`
}

// UpdateStickerSetRequest สำหรับการอัปเดตข้อมูลชุดสติกเกอร์
type UpdateStickerSetRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Author      string `json:"author"`
	IsOfficial  bool   `json:"is_official"`
	IsDefault   bool   `json:"is_default"`
}

// UploadStickerSetCoverRequest สำหรับการอัปโหลดรูปปกชุดสติกเกอร์
type UploadStickerSetCoverRequest struct {
	File *multipart.FileHeader `json:"-"` // ไม่ถูกใช้ใน JSON แต่จะใช้กับ multipart form
}

// AddStickerToSetRequest สำหรับการเพิ่มสติกเกอร์ใหม่ลงในชุด
type AddStickerToSetRequest struct {
	Name       string                `json:"name" form:"name"`
	SortOrder  int                   `json:"sort_order" form:"sort_order"`
	IsAnimated bool                  `json:"is_animated" form:"is_animated"`
	File       *multipart.FileHeader `json:"-"` // ไม่ถูกใช้ใน JSON แต่จะใช้กับ multipart form
}

// UpdateStickerRequest สำหรับการอัปเดตข้อมูลสติกเกอร์
type UpdateStickerRequest struct {
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

// SetStickerSetAsFavoriteRequest สำหรับการตั้งค่าชุดสติกเกอร์เป็นรายการโปรด
type SetStickerSetAsFavoriteRequest struct {
	IsFavorite bool `json:"is_favorite"`
}

// =====================================
// Response DTOs
// =====================================

// StickerSetInfo ข้อมูลของชุดสติกเกอร์
type StickerSetInfo struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Author        string    `json:"author"`
	CoverImageURL string    `json:"cover_image_url"`
	CreatedAt     time.Time `json:"created_at"`
	IsOfficial    bool      `json:"is_official"`
	IsDefault     bool      `json:"is_default"`
	SortOrder     int       `json:"sort_order"`
	StickerCount  int       `json:"sticker_count,omitempty"` // จำนวนสติกเกอร์ในชุด (ถ้ามี)
	IsFavorite    bool      `json:"is_favorite,omitempty"`   // สำหรับผู้ใช้ปัจจุบันเท่านั้น
}

// StickerInfo ข้อมูลของสติกเกอร์
type StickerInfo struct {
	ID             string    `json:"id"`
	StickerSetID   string    `json:"sticker_set_id"`
	Name           string    `json:"name"`
	StickerURL     string    `json:"sticker_url"`
	ThumbnailURL   string    `json:"thumbnail_url"`
	CreatedAt      time.Time `json:"created_at"`
	IsAnimated     bool      `json:"is_animated"`
	SortOrder      int       `json:"sort_order"`
	StickerSetName string    `json:"sticker_set_name,omitempty"` // ชื่อของชุดสติกเกอร์ (ถ้ามี)
}

// UserStickerSetInfo ข้อมูลชุดสติกเกอร์ของผู้ใช้
type UserStickerSetInfo struct {
	ID           string          `json:"id"`
	UserID       string          `json:"user_id"`
	StickerSetID string          `json:"sticker_set_id"`
	PurchasedAt  time.Time       `json:"purchased_at"`
	IsFavorite   bool            `json:"is_favorite"`
	StickerSet   *StickerSetInfo `json:"sticker_set,omitempty"` // ข้อมูลของชุดสติกเกอร์ (ถ้ามี)
}

// UserStickerInfo ข้อมูลสติกเกอร์ที่เกี่ยวข้องกับผู้ใช้
type UserStickerInfo struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	StickerID string      `json:"sticker_id"`
	CreatedAt time.Time   `json:"created_at,omitempty"` // สำหรับสติกเกอร์โปรด
	UsedAt    time.Time   `json:"used_at,omitempty"`    // สำหรับสติกเกอร์ที่ใช้ล่าสุด
	Sticker   StickerInfo `json:"sticker"`
}

// =====================================
// Response Wrapper DTOs
// =====================================

// CreateStickerSetResponse การตอบกลับสำหรับการสร้างชุดสติกเกอร์
type CreateStickerSetResponse struct {
	GenericResponse
	Data StickerSetInfo `json:"data"`
}

// GetStickerSetResponse การตอบกลับสำหรับการดึงข้อมูลชุดสติกเกอร์
type GetStickerSetResponse struct {
	GenericResponse
	Data struct {
		StickerSet StickerSetInfo `json:"sticker_set"`
		Stickers   []StickerInfo  `json:"stickers"`
	} `json:"data"`
}

// GetAllStickerSetsResponse การตอบกลับสำหรับการดึงข้อมูลชุดสติกเกอร์ทั้งหมด
type GetAllStickerSetsResponse struct {
	GenericResponse
	Data struct {
		StickerSets []StickerSetInfo `json:"sticker_sets"`
		Count       int              `json:"count"`
		Limit       int              `json:"limit"`
		Offset      int              `json:"offset"`
	} `json:"data"`
}

// GetDefaultStickerSetsResponse การตอบกลับสำหรับการดึงข้อมูลชุดสติกเกอร์เริ่มต้น
type GetDefaultStickerSetsResponse struct {
	GenericResponse
	Data []StickerSetInfo `json:"data"`
}

// UpdateStickerSetResponse การตอบกลับสำหรับการอัปเดตข้อมูลชุดสติกเกอร์
type UpdateStickerSetResponse struct {
	GenericResponse
	Data StickerSetInfo `json:"data"`
}

// UploadStickerSetCoverResponse การตอบกลับสำหรับการอัปโหลดรูปปกชุดสติกเกอร์
type UploadStickerSetCoverResponse struct {
	GenericResponse
	Data struct {
		StickerSet    StickerSetInfo `json:"sticker_set"`
		CoverImageURL string         `json:"cover_image_url"`
	} `json:"data"`
}

// DeleteStickerSetResponse การตอบกลับสำหรับการลบชุดสติกเกอร์
type DeleteStickerSetResponse struct {
	GenericResponse
}

// AddStickerToSetResponse การตอบกลับสำหรับการเพิ่มสติกเกอร์ใหม่ลงในชุด
type AddStickerToSetResponse struct {
	GenericResponse
	Data StickerInfo `json:"data"`
}

// UpdateStickerResponse การตอบกลับสำหรับการอัปเดตข้อมูลสติกเกอร์
type UpdateStickerResponse struct {
	GenericResponse
	Data StickerInfo `json:"data"`
}

// DeleteStickerResponse การตอบกลับสำหรับการลบสติกเกอร์
type DeleteStickerResponse struct {
	GenericResponse
}

// AddStickerSetToUserResponse การตอบกลับสำหรับการเพิ่มชุดสติกเกอร์ให้ผู้ใช้
type AddStickerSetToUserResponse struct {
	GenericResponse
}

// GetUserStickerSetsResponse การตอบกลับสำหรับการดึงชุดสติกเกอร์ของผู้ใช้
type GetUserStickerSetsResponse struct {
	GenericResponse
	Data []UserStickerSetInfo `json:"data"`
}

// SetStickerSetAsFavoriteResponse การตอบกลับสำหรับการตั้งค่าชุดสติกเกอร์เป็นรายการโปรด
type SetStickerSetAsFavoriteResponse struct {
	GenericResponse
}

// RemoveStickerSetFromUserResponse การตอบกลับสำหรับการลบชุดสติกเกอร์ออกจากผู้ใช้
type RemoveStickerSetFromUserResponse struct {
	GenericResponse
}

// RecordStickerUsageResponse การตอบกลับสำหรับการบันทึกการใช้งานสติกเกอร์
type RecordStickerUsageResponse struct {
	GenericResponse
}

// GetUserRecentStickersResponse การตอบกลับสำหรับการดึงสติกเกอร์ที่ใช้ล่าสุดของผู้ใช้
type GetUserRecentStickersResponse struct {
	GenericResponse
	Data []UserStickerInfo `json:"data"`
}

// GetUserFavoriteStickersResponse การตอบกลับสำหรับการดึงสติกเกอร์โปรดของผู้ใช้
type GetUserFavoriteStickersResponse struct {
	GenericResponse
	Data []UserStickerInfo `json:"data"`
}
