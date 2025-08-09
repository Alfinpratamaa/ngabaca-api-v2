package model

import "gorm.io/gorm"

// Book mendefinisikan skema untuk tabel buku.
type Book struct {
	gorm.Model
	Title           string  `gorm:"not null" json:"title"`
	Slug            string  `gorm:"unique;not null" json:"slug"`
	Author          string  `gorm:"not null" json:"author"`
	Description     string  `json:"description"`
	Price           float64 `gorm:"not null" json:"price"`
	Stock           int     `gorm:"not null" json:"stock"`
	PublishedYear   int     `json:"published_year"`
	CoverImageURL   string  `json:"cover_image_url"`
	PrivateFilePath string  `json:"-"`
	CategoryID      uint    `json:"category_id"`

	// Relasi
	Category   Category    `gorm:"foreignKey:CategoryID" json:"category"`
	OrderItems []OrderItem `gorm:"foreignKey:BookID" json:"-"`
}
