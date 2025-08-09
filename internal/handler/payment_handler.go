package handler

import (
	"fmt"
	"ngabaca/database"
	"ngabaca/internal/model"
	"ngabaca/internal/utils"
	"strings"
	"time" // Tambahkan import time

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// MidtransNotification menangani notifikasi (webhook) dari Midtrans.
func MidtransNotification(c *fiber.Ctx) error {
	var notificationPayload map[string]interface{}

	if err := c.BodyParser(&notificationPayload); err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Cannot parse notification")
	}

	orderIDStr, ok := notificationPayload["order_id"].(string)
	if !ok {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid notification payload")
	}

	parts := strings.Split(orderIDStr, "-")
	if len(parts) < 2 {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid order_id format")
	}
	orderID := parts[1]

	// Gunakan transaksi untuk memastikan update Order dan Payment konsisten
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var order model.Order
		var payment model.Payment

		// 1. Cari pesanan di database
		if err := tx.First(&order, orderID).Error; err != nil {
			return fmt.Errorf("order not found")
		}

		// 2. Cari data pembayaran yang terkait dengan pesanan
		if err := tx.Where("order_id = ?", order.ID).First(&payment).Error; err != nil {
			return fmt.Errorf("payment record not found")
		}

		// 3. Update status dan data pembayaran berdasarkan notifikasi
		transactionStatus, _ := notificationPayload["transaction_status"].(string)
		fraudStatus, _ := notificationPayload["fraud_status"].(string)
		paymentType, _ := notificationPayload["payment_type"].(string)
		transactionID, _ := notificationPayload["transaction_id"].(string)

		payment.PaymentMethod = paymentType
		payment.TransactionID = transactionID
		// Simpan seluruh payload untuk audit
		payment.PaymentGatewayResponse = model.JSONB(notificationPayload)

		paymentStatus := ""
		orderStatus := order.Status // Default ke status yang ada

		if transactionStatus == "capture" {
			if fraudStatus == "challenge" {
				paymentStatus = "challenge"
				orderStatus = "challenge"
			} else if fraudStatus == "accept" {
				paymentStatus = "success"
				orderStatus = "diproses"
				payment.VerifiedAt = time.Now()
			}
		} else if transactionStatus == "settlement" {
			paymentStatus = "success"
			orderStatus = "diproses"
			payment.VerifiedAt = time.Now()
		} else if transactionStatus == "deny" || transactionStatus == "cancel" || transactionStatus == "expire" {
			paymentStatus = "failed"
			orderStatus = "batal"
		} else if transactionStatus == "pending" {
			paymentStatus = "pending"
			orderStatus = "pending"
		}

		payment.Status = paymentStatus
		order.Status = orderStatus

		// 4. Simpan perubahan ke database
		if err := tx.Save(&payment).Error; err != nil {
			return err
		}
		if err := tx.Save(&order).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		// Jika ada error dalam transaksi, kirim response error
		// Midtrans akan mencoba mengirim notifikasi lagi nanti
		fmt.Printf("Error processing notification for order %s: %v\n", orderID, err)
		return utils.GenericError(c, fiber.StatusInternalServerError, err.Error())
	}

	fmt.Printf("Successfully processed notification for order %s\n", orderID)
	return c.SendStatus(fiber.StatusOK)
}
