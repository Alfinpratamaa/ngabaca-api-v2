package repository

import (
	"ngabaca/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OrderRepository mendefinisikan kontrak untuk data pesanan.
type OrderRepository interface {
	FindAll(status string) ([]model.Order, error)
	FindByID(id uuid.UUID) (model.Order, error)
	Update(order *model.Order) (*model.Order, error)
	FindByUserID(userID uuid.UUID) ([]model.Order, error) // <-- TAMBAHKAN INI
	FindByIDAndUserID(id, userID uuid.UUID) (model.Order, error)
	Create(order *model.Order) (*model.Order, error)
}

type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository adalah constructor untuk orderRepository.
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

// Create menyimpan order baru ke database.
func (r *orderRepository) Create(order *model.Order) (*model.Order, error) {
	err := r.db.Create(order).Error
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (r *orderRepository) FindAll(status string) ([]model.Order, error) {
	var orders []model.Order
	query := r.db.Preload("User").Order("created_at desc")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Find(&orders).Error
	return orders, err
}

func (r *orderRepository) FindByID(id uuid.UUID) (model.Order, error) {
	var order model.Order
	err := r.db.Preload("User").
		Preload("OrderItems.Book").
		Preload("Payment").
		First(&order, id).Error
	return order, err
}

func (r *orderRepository) Update(order *model.Order) (*model.Order, error) {
	err := r.db.Save(order).Error
	return order, err
}

func (r *orderRepository) FindByUserID(userID uuid.UUID) ([]model.Order, error) {
	var orders []model.Order
	err := r.db.Where("user_id = ?", userID).Preload("OrderItems.Book").Find(&orders).Error
	return orders, err
}

func (r *orderRepository) FindByIDAndUserID(id, userID uuid.UUID) (model.Order, error) {
	var order model.Order
	err := r.db.Where("id = ? AND user_id = ?", id, userID).
		Preload("OrderItems.Book").
		Preload("Payment").
		First(&order).Error
	return order, err
}
