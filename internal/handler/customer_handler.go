package handler

import (
	"fmt"
	"ngabaca/config"
	"ngabaca/database"
	"ngabaca/internal/model"
	"ngabaca/internal/utils"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"gorm.io/gorm"
)

type CheckoutItemRequest struct {
	BookID   uint `json:"book_id" validate:"required"`
	Quantity int  `json:"quantity" validate:"required,gt=0"`
}

// CheckoutRequest merepresentasikan seluruh permintaan checkout.
type CheckoutRequest struct {
	Items           []CheckoutItemRequest `json:"items" validate:"required,min=1,dive"`
	ShippingAddress string                `json:"shipping_address" validate:"required"`
	Notes           string                `json:"notes"`
}

func GetCustomerOrders(c *fiber.Ctx) error {
	userClaims := c.Locals("user").(jwt.MapClaims)

	userID := uint(userClaims["user_id"].(float64))

	var orders []model.Order

	err := database.DB.Where("user_id = ?", userID).Preload("OrderItems.Book").Find(&orders).Error

	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Database error")
	}

	return c.JSON(orders)
}

func GetCustomerOrderDetail(c *fiber.Ctx) error {
	userClaims := c.Locals("user").(jwt.MapClaims)
	userID := uint(userClaims["user_id"].(float64))
	orderID := c.Params("id")

	var order model.Order
	// Pastikan pesanan yang diambil adalah milik user yang sedang login
	err := database.DB.Where("id = ? AND user_id = ?", orderID, userID).
		Preload("OrderItems.Book").
		Preload("Payment").
		First(&order).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.GenericError(c, fiber.StatusNotFound, "Order not found")
		}
		return utils.GenericError(c, fiber.StatusInternalServerError, "Database error")
	}

	return c.JSON(order)
}

func Checkout(c *fiber.Ctx) error {
	req := new(CheckoutRequest)

	userClaims := c.Locals("user").(jwt.MapClaims)
	userID := uint(userClaims["user_id"].(float64))

	if err := c.BodyParser(req); err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if errs := utils.ValidateStruct(req); errs != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": errs})
	}

	var totalPrice float64
	var orderItems []model.OrderItem

	var midtransItems []midtrans.ItemDetails

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		for _, item := range req.Items {
			var book model.Book

			if err := tx.First(&book, item.BookID).Error; err != nil {
				return fmt.Errorf("Book with ID %d not found", item.BookID)
			}

			if book.Stock < item.Quantity {
				return fmt.Errorf("Insufficient stock for book %s", book.Title)
			}

			newStock := book.Stock - item.Quantity
			if err := tx.Model(&book).Update("stock", newStock).Error; err != nil {
				return err
			}

			itemPrice := book.Price * float64(item.Quantity)
			totalPrice += itemPrice
			orderItems = append(orderItems, model.OrderItem{
				BookID:   item.BookID,
				Quantity: item.Quantity,
				Price:    book.Price,
			})

			midtransItems = append(midtransItems, midtrans.ItemDetails{
				ID:    strconv.Itoa(int(book.ID)),
				Name:  book.Title,
				Price: int64(book.Price),
				Qty:   int32(item.Quantity),
			})
		}

		order := model.Order{
			UserID:          userID,
			TotalPrice:      totalPrice,
			Status:          "pending", // Status awal
			ShippingAddress: req.ShippingAddress,
			Notes:           req.Notes,
			OrderItems:      orderItems,
		}

		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		c.Locals("order_id", order.ID)

		payment := model.Payment{
			OrderID:    order.ID,
			Status:     "pending",
			TotalPrice: order.TotalPrice,
			Currency:   "IDR",
			ExpiresAt:  time.Now().Add(24 * time.Hour),
		}
		if err := tx.Create(&payment).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, err.Error())
	}

	cfg, _ := config.LoadConfig(".")
	var s = snap.Client{}
	s.New(cfg.MidtransServerKey, midtrans.Sandbox)

	orderID, _ := c.Locals("order_id").(uint)
	var user model.User
	database.DB.First(&user, userID)

	snapReq := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  fmt.Sprintf("NGABACA-%d-%d", orderID, time.Now().Unix()),
			GrossAmt: int64(totalPrice),
		},
		Items: &midtransItems,
		CustomerDetail: &midtrans.CustomerDetails{
			FName: user.Name,
			Email: user.Email,
			Phone: user.PhoneNumber,
		},
	}

	snapRes, errSnap := s.CreateTransaction(snapReq)

	if errSnap != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Failed to create Midtrans transaction")
	}

	return c.JSON(snapRes)

}
