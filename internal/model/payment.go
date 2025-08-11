package model

import (
	"time"

	"github.com/google/uuid"
)

// Payment mendefinisikan skema untuk tabel pembayaran.
type Payment struct {
	Basemodel
	OrderID                uuid.UUID `gorm:"type:uuid;not null" json:"order_id"`
	TransactionID          string    `json:"transaction_id"`
	TotalPrice             float64   `gorm:"not null" json:"total_price"`
	Currency               string    `gorm:"default:'IDR'" json:"currency"`
	PaymentMethod          string    `json:"payment_method"`
	ProofURL               string    `json:"proof_url"`
	Status                 string    `gorm:"default:'pending'" json:"status"`
	PaymentStatusGateway   string    `json:"payment_status_gateway"`
	PaymentGatewayResponse JSONB     `gorm:"type:jsonb" json:"payment_gateway_response"`
	VerifiedAt             time.Time `gorm:"null" json:"verified_at"`
	VerifiedBy             *uint     `json:"verified_by"`
	ExpiresAt              time.Time `json:"expires_at"`

	// Relasi
	Order    *Order `gorm:"foreignKey:OrderID" json:"order"`
	Verifier User   `gorm:"foreignKey:VerifiedBy" json:"verifier,omitempty"`
}
