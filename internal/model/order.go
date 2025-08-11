package model

import "github.com/google/uuid"

// Order mendefinisikan skema untuk tabel pesanan.
type Order struct {
	Basemodel
	UserID          uuid.UUID `gorm:"not null" json:"user_id"`
	TotalPrice      float64   `gorm:"not null" json:"total_price"`
	Taxes           float64   `gorm:"default:0" json:"taxes"`
	ShippingCost    float64   `gorm:"default:0" json:"shipping_cost"`
	Status          string    `gorm:"default:'pending';not null" json:"status"`
	Notes           string    `json:"notes"`
	ShippingAddress string    `json:"shipping_address"`

	// Relasi
	User       User        `gorm:"foreignKey:UserID" json:"user"`
	OrderItems []OrderItem `gorm:"foreignKey:OrderID" json:"order_items"`
	Payment    Payment     `gorm:"foreignKey:OrderID" json:"payment,omitempty"`
}
