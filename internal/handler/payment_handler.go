package handler

import (
	"fmt"
	"ngabaca/internal/service"
	"ngabaca/internal/utils"

	"github.com/gofiber/fiber/v2"
)

type PaymentHandler struct {
	paymentService service.PaymentService
}

func NewPaymentHandler(paymentService service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

func (h *PaymentHandler) MidtransNotification(c *fiber.Ctx) error {
	var notificationPayload map[string]interface{}

	if err := c.BodyParser(&notificationPayload); err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Cannot parse notification")
	}

	// Panggil service untuk melakukan semua pekerjaan berat
	err := h.paymentService.UpdatePaymentStatus(notificationPayload)
	if err != nil {
		fmt.Printf("Error processing notification: %v\n", err)
		// Kembalikan error agar Midtrans mencoba lagi nanti
		return utils.GenericError(c, fiber.StatusInternalServerError, err.Error())
	}

	fmt.Printf("Successfully processed notification for order: %s\n", notificationPayload["order_id"])
	// Beri tahu Midtrans bahwa notifikasi sudah diterima
	return c.SendStatus(fiber.StatusOK)
}
