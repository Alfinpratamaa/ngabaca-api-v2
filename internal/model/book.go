package model

import "github.com/google/uuid"

// Book mendefinisikan skema untuk tabel buku.
type Book struct {
	Basemodel
	Title           string    `gorm:"not null" json:"title"`
	Slug            string    `gorm:"unique;not null" json:"slug"`
	Author          string    `gorm:"not null" json:"author"`
	Description     string    `json:"description"`
	Price           float64   `gorm:"not null" json:"price"`
	Stock           int       `gorm:"not null" json:"stock"`
	PublishedYear   int       `gorm:"not null" json:"published_year"`
	CoverImageURL   string    `json:"cover_image_url"`
	PrivateFilePath string    `json:"private_file_path"`
	CategoryID      uuid.UUID `json:"category_id"`

	// Relasi
	Reviews     []Review    `gorm:"foreignKey:BookID" json:"reviews,omitempty"`
	AvgRating   float64     `gorm:"-" json:"avg_rating"`
	ReviewCount int         `gorm:"-" json:"review_count"`
	Category    Category    `gorm:"foreignKey:CategoryID" json:"category"`
	OrderItems  []OrderItem `gorm:"foreignKey:BookID" json:"-"`
}
