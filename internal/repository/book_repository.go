package repository

import (
	"ngabaca/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BookRepository mendefinisikan "kontrak" atau fungsi apa saja yang harus dimiliki oleh repository buku.
type BookRepository interface {
	FindAll() ([]model.Book, error)
	FindByID(id uuid.UUID) (model.Book, error)
	FindBySlug(slug string) (model.Book, error)
	Create(book *model.Book) (*model.Book, error)
	Update(book *model.Book) (*model.Book, error)
	Delete(book *model.Book) error
	IsSlugExist(slug string, id uuid.UUID) (bool, error)
	Search(keyword string) ([]model.Book, error)
}

// bookRepository adalah implementasi nyata dari BookRepository.
type bookRepository struct {
	db *gorm.DB
}

// book requset
type BookRequest struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Slug          string  `json:"slug"`
	Description   string  `json:"description"`
	Author        string  `json:"author"`
	Price         float64 `json:"price"`
	Stock         int     `json:"stock"`
	PublishedYear int     `json:"published_year"`
	CoverImageURL string  `json:"cover_image_url"`
	CategoryID    string  `json:"category_id"`
}

// NewBookRepository adalah "pabrik" untuk membuat instance bookRepository baru.
func NewBookRepository(db *gorm.DB) BookRepository {
	return &bookRepository{db: db}
}

// Implementasi setiap fungsi dari interface
func (r *bookRepository) FindAll() ([]model.Book, error) {
	var books []model.Book
	err := r.db.Preload("Category").Find(&books).Error
	return books, err
}

func (r *bookRepository) FindByID(id uuid.UUID) (model.Book, error) {
	var book model.Book
	err := r.db.Preload("Category").First(&book, id).Error
	return book, err
}

func (r *bookRepository) FindBySlug(slug string) (model.Book, error) {
	var book model.Book
	err := r.db.Preload("Category").Preload("Reviews.User").Where("slug = ?", slug).First(&book).Error
	if err != nil {
		return book, err
	}

	// Hitung rata-rata rating dan jumlah ulasan
	type RatingResult struct {
		AvgRating   float64
		ReviewCount int
	}
	var result RatingResult

	r.db.Model(&model.Review{}).
		Where("book_id = ?", book.ID).
		Select("COALESCE(AVG(rating), 0) as avg_rating, COUNT(*) as review_count").
		Scan(&result)

	book.AvgRating = result.AvgRating
	book.ReviewCount = result.ReviewCount

	return book, nil
}
func (r *bookRepository) Create(book *model.Book) (*model.Book, error) {
	err := r.db.Clauses(clause.Returning{}).Create(book).Error
	return book, err
}

func (r *bookRepository) Update(book *model.Book) (*model.Book, error) {
	err := r.db.Save(book).Error
	return book, err
}

func (r *bookRepository) Delete(book *model.Book) error {
	return r.db.Delete(book).Error
}

func (r *bookRepository) IsSlugExist(slug string, id uuid.UUID) (bool, error) {
	var count int64
	query := r.db.Model(&model.Book{}).Where("slug = ?", slug)
	if id != uuid.Nil {
		query = query.Where("id <> ?", id)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

func (r *bookRepository) Search(keyword string) ([]model.Book, error) {
	var books []model.Book
	searchPattern := "%" + keyword + "%"

	err := r.db.
		Preload("Category").
		Where("title ILIKE ? OR author ILIKE ?", searchPattern, searchPattern).
		Find(&books).Error

	return books, err
}
