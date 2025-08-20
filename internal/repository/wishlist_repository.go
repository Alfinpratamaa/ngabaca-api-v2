package repository

import (
	"ngabaca/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WishlistRepository interface {
	Create(wishlist *model.Wishlist) error
	Delete(userID, bookID uuid.UUID) error
	FindByUserID(userID uuid.UUID) ([]model.Wishlist, error)
	Check(userID, bookID uuid.UUID) (bool, error)
}

type wishlistRepository struct {
	db *gorm.DB
}

func NewWishlistRepository(db *gorm.DB) WishlistRepository {
	return &wishlistRepository{db: db}
}

func (r *wishlistRepository) Create(wishlist *model.Wishlist) error {
	return r.db.Create(wishlist).Error
}

func (r *wishlistRepository) Delete(userID, bookID uuid.UUID) error {
	return r.db.Where("user_id = ? AND book_id = ?", userID, bookID).Delete(&model.Wishlist{}).Error
}

func (r *wishlistRepository) FindByUserID(userID uuid.UUID) ([]model.Wishlist, error) {
	var wishlistItems []model.Wishlist
	// Preload data buku agar bisa ditampilkan
	err := r.db.Preload("Book.Category").Where("user_id = ?", userID).Find(&wishlistItems).Error
	return wishlistItems, err
}

func (r *wishlistRepository) Check(userID, bookID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.Wishlist{}).Where("user_id = ? AND book_id = ?", userID, bookID).Count(&count).Error
	return count > 0, err
}
