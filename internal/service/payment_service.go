package service

import (
	"fmt"
	"ngabaca/internal/model"
	"ngabaca/internal/repository"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentService interface {
	UpdatePaymentStatus(payload map[string]interface{}) error
}

type paymentService struct {
	db          *gorm.DB // Dibutuhkan untuk transaksi
	orderRepo   repository.OrderRepository
	paymentRepo repository.PaymentRepository
}

func NewPaymentService(db *gorm.DB, orderRepo repository.OrderRepository, paymentRepo repository.PaymentRepository) PaymentService {
	return &paymentService{db, orderRepo, paymentRepo}
}

func (s *paymentService) UpdatePaymentStatus(payload map[string]interface{}) error {
	orderIDStr, ok := payload["order_id"].(string)
	if !ok {
		return fmt.Errorf("invalid notification payload: missing order_id")
	}

	parts := strings.Split(orderIDStr, "-")
	if len(parts) < 2 {
		return fmt.Errorf("invalid order_id format")
	}
	orderID, err := uuid.Parse(parts[1])
	if err != nil {
		return fmt.Errorf("invalid order_id format: not a valid UUID")
	}

	// Gunakan transaksi untuk memastikan update Order dan Payment konsisten
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Buat instance repo dengan 'tx' agar semua operasi masuk dalam transaksi
		txOrderRepo := repository.NewOrderRepository(tx)
		txPaymentRepo := repository.NewPaymentRepository(tx)

		order, err := txOrderRepo.FindByID(orderID)
		if err != nil {
			return fmt.Errorf("order with id %s not found", orderID)
		}

		payment, err := txPaymentRepo.FindByOrderID(order.ID)
		if err != nil {
			return fmt.Errorf("payment for order id %s not found", orderID)
		}

		// Logika update status (dipindahkan dari handler)
		transactionStatus, _ := payload["transaction_status"].(string)
		fraudStatus, _ := payload["fraud_status"].(string)
		paymentType, _ := payload["payment_type"].(string)
		transactionID, _ := payload["transaction_id"].(string)

		payment.PaymentMethod = paymentType
		payment.TransactionID = transactionID
		payment.PaymentGatewayResponse = model.JSONB(payload)

		if transactionStatus == "capture" || transactionStatus == "settlement" {
			if fraudStatus == "accept" || fraudStatus == "challenge" {
				payment.Status = "success"
				order.Status = "diproses"
				payment.VerifiedAt = time.Now()
			}
		} else if transactionStatus == "deny" || transactionStatus == "cancel" || transactionStatus == "expire" {
			payment.Status = "failed"
			order.Status = "batal"
		}

		// Simpan perubahan menggunakan repository
		if _, err := txPaymentRepo.Update(&payment); err != nil {
			return err
		}
		if _, err := txOrderRepo.Update(&order); err != nil {
			return err
		}

		return nil
	})
}
