package model

// Category mendefinisikan skema untuk tabel kategori buku.
type Category struct {
	Basemodel
	Name string `gorm:"unique;not null" json:"name"`
	Slug string `gorm:"unique;not null" json:"slug"`

	// Relasi
	Books []Book `gorm:"foreignKey:CategoryID" json:"books,omitempty"`
}
