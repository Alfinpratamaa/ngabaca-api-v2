package model

import "github.com/google/uuid"

// OrderItem mendefinisikan skema untuk setiap item dalam pesanan.
type OrderItem struct {
	Basemodel
	OrderID  uuid.UUID `gorm:"not null" json:"order_id"`
	BookID   uuid.UUID `gorm:"not null" json:"book_id"`
	Quantity int       `gorm:"not null" json:"quantity"`
	Price    float64   `gorm:"not null" json:"price"`

	// Relasi
	Order Order `gorm:"foreignKey:OrderID" json:"-"`
	Book  Book  `gorm:"foreignKey:BookID" json:"book"`
}

// TableName secara eksplisit memberitahu GORM nama tabel yang benar.
// GORM cenderung membuat jamak nama struct (misal: OrderItems), ini untuk memastikan namanya 'order_items'.
func (OrderItem) TableName() string {
	return "order_items"
}
