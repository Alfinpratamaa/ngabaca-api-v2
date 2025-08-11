package repository

import (
	"ngabaca/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PaymentRepository interface {
	Create(payment *model.Payment) (*model.Payment, error)
	FindByOrderID(orderID uuid.UUID) (model.Payment, error)
	Update(payment *model.Payment) (*model.Payment, error)
}

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(payment *model.Payment) (*model.Payment, error) {
	err := r.db.Clauses(clause.Returning{}).Create(payment).Error
	return payment, err
}

func (r *paymentRepository) FindByOrderID(orderID uuid.UUID) (model.Payment, error) {
	var payment model.Payment
	err := r.db.Where("order_id = ?", orderID).First(&payment).Error
	return payment, err
}

func (r *paymentRepository) Update(payment *model.Payment) (*model.Payment, error) {
	err := r.db.Save(payment).Error
	return payment, err
}
