// domain/repository/sticker_repository.go
package repository

import (
	"github.com/google/uuid"
	"github.com/thizplus/gofiber-chat-api/domain/models"
)

// StickerRepository เป็น interface สำหรับจัดการข้อมูลสติกเกอร์
type StickerRepository interface {
	// สำหรับชุดสติกเกอร์
	CreateStickerSet(stickerSet *models.StickerSet) error
	GetStickerSetByID(id uuid.UUID) (*models.StickerSet, error)
	GetAllStickerSets(limit, offset int) ([]*models.StickerSet, int64, error)
	GetDefaultStickerSets() ([]*models.StickerSet, error)
	UpdateStickerSet(stickerSet *models.StickerSet) error
	DeleteStickerSet(id uuid.UUID) error

	// สำหรับสติกเกอร์
	CreateSticker(sticker *models.Sticker) error
	GetStickerByID(id uuid.UUID) (*models.Sticker, error)
	GetStickersBySetID(setID uuid.UUID) ([]*models.Sticker, error)
	UpdateSticker(sticker *models.Sticker) error
	DeleteSticker(id uuid.UUID) error

	// สำหรับ User Sticker Set
	AddStickerSetToUser(userStickerSet *models.UserStickerSet) error
	GetUserStickerSets(userID uuid.UUID) ([]*models.StickerSet, error)
	SetStickerSetAsFavorite(userID, stickerSetID uuid.UUID, isFavorite bool) error
	RemoveStickerSetFromUser(userID, stickerSetID uuid.UUID) error

	// สำหรับประวัติการใช้สติกเกอร์
	AddRecentSticker(userRecentSticker *models.UserRecentSticker) error
	GetUserRecentStickers(userID uuid.UUID, limit int) ([]*models.Sticker, error)
	GetUserFavoriteStickers(userID uuid.UUID) ([]*models.Sticker, error)
}
