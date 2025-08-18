package model

// User mendefinisikan skema untuk tabel pengguna.
type User struct {
	Basemodel
	Name        string  `gorm:"not null" json:"name"`
	Email       string  `gorm:"unique;not null" json:"email"`
	GoogleID    *string `gorm:"unique" json:"google_id,omitempty"`
	Password    string  `gorm:"not null" json:"-"`
	Avatar      string  `json:"avatar"`
	PhoneNumber string  `json:"phone_number"`
	Role        string  `gorm:"default:'pelanggan';not null" json:"role"`

	// Relasi
	Orders           []Order   `gorm:"foreignKey:UserID" json:"orders,omitempty"`
	VerifiedPayments []Payment `gorm:"foreignKey:VerifiedBy" json:"verified_payments,omitempty"`
	Reviews          []Review  `gorm:"foreignKey:UserID" json:"reviews,omitempty"`
}
