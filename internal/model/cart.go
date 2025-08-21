package model

import "github.com/google/uuid"

// Cart merepresentasikan keranjang belanja milik satu pengguna.
type Cart struct {
	Basemodel
	UserID    uuid.UUID  `gorm:"unique;not null"`
	CartItems []CartItem `gorm:"foreignKey:CartID"`
}
