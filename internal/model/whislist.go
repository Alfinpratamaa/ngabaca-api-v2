package model

import (
	"time"

	"github.com/google/uuid"
)

type Wishlist struct {
	UserID    uuid.UUID `gorm:"primaryKey"`
	BookID    uuid.UUID `gorm:"primaryKey"`
	CreatedAt time.Time

	// Relasi
	User User `gorm:"foreignKey:UserID"`
	Book Book `gorm:"foreignKey:BookID"`
}
