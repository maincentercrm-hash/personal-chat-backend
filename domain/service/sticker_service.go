// domain/service/sticker_service.go
package service

import (
	"mime/multipart"

	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

// StickerService เป็น interface สำหรับตรรกะทางธุรกิจเกี่ยวกับสติกเกอร์
type StickerService interface {
	// สำหรับชุดสติกเกอร์
	CreateStickerSet(name, description, author string, isOfficial, isDefault bool) (*models.StickerSet, error)
	GetStickerSetByID(id uuid.UUID) (*models.StickerSet, error)
	GetAllStickerSets(limit, offset int) ([]*models.StickerSet, int64, error)
	GetDefaultStickerSets() ([]*models.StickerSet, error)
	UpdateStickerSet(id uuid.UUID, name, description, author string, isOfficial, isDefault bool) (*models.StickerSet, error)
	DeleteStickerSet(id uuid.UUID) error
	UploadStickerSetCover(id uuid.UUID, file *multipart.FileHeader) (*models.StickerSet, error)

	// สำหรับสติกเกอร์
	AddStickerToSet(setID uuid.UUID, name string, file *multipart.FileHeader, isAnimated bool, sortOrder int) (*models.Sticker, error)
	GetStickerByID(id uuid.UUID) (*models.Sticker, error)
	GetStickersBySetID(setID uuid.UUID) ([]*models.Sticker, error)
	UpdateSticker(id uuid.UUID, name string, sortOrder int) (*models.Sticker, error)
	DeleteSticker(id uuid.UUID) error

	// สำหรับ User Sticker Set
	AddStickerSetToUser(userID, stickerSetID uuid.UUID) error
	GetUserStickerSets(userID uuid.UUID) ([]*models.StickerSet, error)
	SetStickerSetAsFavorite(userID, stickerSetID uuid.UUID, isFavorite bool) error
	RemoveStickerSetFromUser(userID, stickerSetID uuid.UUID) error

	// สำหรับประวัติการใช้สติกเกอร์
	RecordStickerUsage(userID, stickerID uuid.UUID) error
	GetUserRecentStickers(userID uuid.UUID, limit int) ([]*models.Sticker, error)
	GetUserFavoriteStickers(userID uuid.UUID) ([]*models.Sticker, error)
}
