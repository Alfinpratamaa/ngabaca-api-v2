package repository

import (
	"ngabaca/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReviewRepository interface {
	Create(review *model.Review) error
	FindByBookID(bookID uuid.UUID) ([]model.Review, error)
	CheckExisting(userID, bookID uuid.UUID) (bool, error)
}

type reviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Create(review *model.Review) error {
	return r.db.Create(review).Error
}

func (r *reviewRepository) FindByBookID(bookID uuid.UUID) ([]model.Review, error) {
	var reviews []model.Review
	err := r.db.Preload("User").Where("book_id = ?", bookID).Find(&reviews).Error
	return reviews, err
}

func (r *reviewRepository) CheckExisting(userID, bookID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.Review{}).Where("user_id = ? AND book_id = ?", userID, bookID).Count(&count).Error
	return count > 0, err
}
