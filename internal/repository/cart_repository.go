package repository

import (
	"errors"
	"ngabaca/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CartRepository interface {
	GetCartByUserID(userID uuid.UUID) (model.Cart, error)
	AddItem(userID, bookID uuid.UUID, quantity int) error
	UpdateCartItem(userID, itemID uuid.UUID, quantity int) error
	RemoveItem(userID, bookID uuid.UUID) error
}

type cartRepository struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) GetCartByUserID(userID uuid.UUID) (model.Cart, error) {
	var cart model.Cart
	err := r.db.Preload("CartItems.Book").Where("user_id = ?", userID).First(&cart).Error
	return cart, err
}

func (r *cartRepository) AddItem(userID, bookID uuid.UUID, quantity int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Dapatkan cart milik user
		var cart model.Cart
		if err := tx.Where("user_id = ?", userID).First(&cart).Error; err != nil {
			return errors.New("cart not found for user")
		}

		// 2. Cek apakah item sudah ada di keranjang
		var item model.CartItem
		err := tx.Where("cart_id = ? AND book_id = ?", cart.ID, bookID).First(&item).Error

		if err == nil {
			// Item sudah ada, update kuantitasnya
			item.Quantity += quantity
			return tx.Save(&item).Error
		}

		if err == gorm.ErrRecordNotFound {
			// Item belum ada, buat baru
			newItem := model.CartItem{
				CartID:   cart.ID,
				BookID:   bookID,
				Quantity: quantity,
			}
			return tx.Create(&newItem).Error
		}

		return err
	})
}

// update such a decrease quantity or increase quantity
func (r *cartRepository) UpdateCartItem(userID, itemID uuid.UUID, quantity int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Dapatkan cart milik user
		var cart model.Cart
		if err := tx.Where("user_id = ?", userID).First(&cart).Error; err != nil {
			return errors.New("cart not found for user")
		}

		// 2. Cari item di keranjang
		var item model.CartItem
		if err := tx.Where("id = ? AND cart_id = ?", itemID, cart.ID).First(&item).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("item not found in cart")
			}
			return err
		}

		// 3. Update kuantitas item
		item.Quantity = quantity
		return tx.Save(&item).Error
	})
}

func (r *cartRepository) RemoveItem(userID, bookID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Dapatkan cart milik user
		var cart model.Cart
		if err := tx.Where("user_id = ?", userID).First(&cart).Error; err != nil {
			return errors.New("cart not found for user")
		}

		// 2. Cari item di keranjang
		var item model.CartItem
		if err := tx.Where("cart_id = ? AND book_id = ?", cart.ID, bookID).First(&item).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("item not found in cart")
			}
			return err
		}

		// 3. Hapus item dari keranjang
		if err := tx.Delete(&item).Error; err != nil {
			return err
		}

		return nil
	})
}
