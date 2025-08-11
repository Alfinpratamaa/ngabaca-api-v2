package model

import "github.com/google/uuid"

// Review mendefinisikan skema untuk tabel ulasan buku.
type Review struct {
	Basemodel
	Rating  int       `gorm:"not null" json:"rating"` // Rating bintang 1-5
	Comment string    `json:"comment"`
	BookID  uuid.UUID `gorm:"not null" json:"book_id"`
	UserID  uuid.UUID `gorm:"not null" json:"user_id"`

	// Relasi
	User User `gorm:"foreignKey:UserID" json:"user"`
	Book Book `gorm:"foreignKey:BookID" json:"-"` // Hindari data buku berulang di JSON
}
