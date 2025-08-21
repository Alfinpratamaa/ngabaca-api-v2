package model

import "github.com/google/uuid"

// CartItem merepresentasikan satu item buku di dalam keranjang.
type CartItem struct {
	Basemodel
	CartID   uuid.UUID `gorm:"not null"`
	BookID   uuid.UUID `gorm:"not null"`
	Quantity int       `gorm:"not null"`
	Cart     Cart      `gorm:"foreignKey:CartID"`
	Book     Book      `gorm:"foreignKey:BookID"`
}
