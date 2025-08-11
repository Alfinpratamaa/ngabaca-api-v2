package service

import (
	"fmt"
	"ngabaca/internal/model"
	"ngabaca/internal/repository"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreateOrderItemRequest struct {
	BookID   uuid.UUID `json:"book_id" validate:"required"`
	Quantity int       `json:"quantity" validate:"required,gt=0"`
}

type CreateOrderRequest struct {
	Items           []CreateOrderItemRequest `json:"items" validate:"required,min=1,dive"`
	ShippingAddress string                   `json:"shipping_address" validate:"required"`
	Notes           string                   `json:"notes"`
}
type OrderService interface {
	CreateOrder(userID uuid.UUID, req *CreateOrderRequest) (*model.Order, float64, error)
}

// CheckExisting implements repository.ReviewRepository.
func (s *orderService) CheckExisting(userID uuid.UUID, bookID uuid.UUID) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

// Create implements repository.ReviewRepository.
func (s *orderService) Create(review *model.Review) error {
	return fmt.Errorf("not implemented")
}

// FindByBookID implements repository.ReviewRepository.
func (s *orderService) FindByBookID(bookID uuid.UUID) ([]model.Review, error) {
	return nil, fmt.Errorf("not implemented")
}

type orderService struct {
	db          *gorm.DB
	bookRepo    repository.BookRepository
	orderRepo   repository.OrderRepository
	paymentRepo repository.PaymentRepository
}

func NewOrderService(db *gorm.DB, bookRepo repository.BookRepository, orderRepo repository.OrderRepository, paymentRepo repository.PaymentRepository) OrderService {
	return &orderService{db, bookRepo, orderRepo, paymentRepo}
}

// CreateOrder berisi semua logika transaksi checkout
func (s *orderService) CreateOrder(userID uuid.UUID, req *CreateOrderRequest) (*model.Order, float64, error) {
	var totalPrice float64
	var orderItems []model.OrderItem
	var order model.Order

	// Gunakan Transaksi Database
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Buat instance repository baru yang menggunakan 'tx' (transaksi)
		txBookRepo := repository.NewBookRepository(tx)
		txOrderRepo := repository.NewOrderRepository(tx)
		txPaymentRepo := repository.NewPaymentRepository(tx)

		for _, item := range req.Items {
			book, err := txBookRepo.FindByID(item.BookID)
			if err != nil {
				return fmt.Errorf("Book with ID %s not found", item.BookID)
			}
			if book.Stock < item.Quantity {
				return fmt.Errorf("Insufficient stock for book %s", book.Title)
			}

			// Kurangi stok buku
			book.Stock -= item.Quantity
			if _, err := txBookRepo.Update(&book); err != nil {
				return err
			}

			totalPrice += book.Price * float64(item.Quantity)
			orderItems = append(orderItems, model.OrderItem{
				BookID:   item.BookID,
				Quantity: item.Quantity,
				Price:    book.Price,
			})
		}

		// Buat record Order
		orderToCreate := &model.Order{
			UserID:          userID,
			TotalPrice:      totalPrice,
			Status:          "pending",
			ShippingAddress: req.ShippingAddress,
			Notes:           req.Notes,
			OrderItems:      orderItems,
		}
		createdOrder, err := txOrderRepo.Create(orderToCreate)
		if err != nil {
			return err
		}
		order = *createdOrder

		// Buat record Payment
		paymentToCreate := &model.Payment{
			OrderID:    order.ID,
			Status:     "pending",
			TotalPrice: order.TotalPrice,
			Currency:   "IDR",
			ExpiresAt:  time.Now().Add(24 * time.Hour),
		}
		if _, err := txPaymentRepo.Create(paymentToCreate); err != nil {
			return err
		}

		return nil // Commit transaksi
	})

	return &order, totalPrice, err
}
