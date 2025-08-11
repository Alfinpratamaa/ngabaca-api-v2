package handler

import (
	"fmt"
	"ngabaca/config"
	"ngabaca/internal/model"
	"ngabaca/internal/repository"
	"ngabaca/internal/service"
	"ngabaca/internal/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"gorm.io/gorm"
)

// CustomerHandler menampung semua dependency yang dibutuhkan untuk fitur-fitur pelanggan.
type CustomerHandler struct {
	orderRepo    repository.OrderRepository
	userRepo     repository.UserRepository // Dibutuhkan untuk mengambil detail user saat checkout
	orderService service.OrderService
	reviewRepo   repository.ReviewRepository
	cfg          config.Config
}

// NewCustomerHandler adalah constructor untuk CustomerHandler.
func NewCustomerHandler(
	orderRepo repository.OrderRepository,
	userRepo repository.UserRepository,
	orderService service.OrderService,
	reviewRepo repository.ReviewRepository,
	cfg config.Config,
) *CustomerHandler {
	return &CustomerHandler{
		orderRepo:    orderRepo,
		userRepo:     userRepo,
		reviewRepo:   reviewRepo,
		orderService: orderService,
		cfg:          cfg,
	}
}

// GetCustomerOrders mengambil semua riwayat pesanan milik pengguna yang sedang login.
func (h *CustomerHandler) GetCustomerOrders(c *fiber.Ctx) error {
	userClaims := c.Locals("user").(jwt.MapClaims)
	userID, _ := uuid.Parse(userClaims["user_id"].(string))

	orders, err := h.orderRepo.FindByUserID(userID)
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Could not fetch orders")
	}

	return c.JSON(orders)
}

// GetCustomerOrderDetail mengambil detail satu pesanan spesifik milik pengguna yang sedang login.
func (h *CustomerHandler) GetCustomerOrderDetail(c *fiber.Ctx) error {
	userClaims := c.Locals("user").(jwt.MapClaims)
	userID, _ := uuid.Parse(userClaims["user_id"].(string))
	orderID, _ := uuid.Parse(c.Params("id"))

	order, err := h.orderRepo.FindByIDAndUserID(orderID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.GenericError(c, fiber.StatusNotFound, "Order not found or you don't have permission to view it")
		}
		return utils.GenericError(c, fiber.StatusInternalServerError, "Database error")
	}

	return c.JSON(order)
}

// Checkout memproses permintaan checkout dari pengguna.
func (h *CustomerHandler) Checkout(c *fiber.Ctx) error {
	// 1. Parse dan validasi request body menggunakan struct dari paket service.
	req := new(service.CreateOrderRequest)
	if err := c.BodyParser(req); err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid request body")
	}
	if errs := utils.ValidateStruct(req); errs != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": errs})
	}

	// 2. Ambil ID pengguna dari token.
	userClaims := c.Locals("user").(jwt.MapClaims)
	userID, _ := uuid.Parse(userClaims["user_id"].(string))

	// 3. Panggil "Kepala Koki" (OrderService) untuk memproses semua logika bisnis yang kompleks.
	order, totalPrice, err := h.orderService.CreateOrder(userID, req)
	if err != nil {
		// Error dari service bisa jadi karena stok tidak cukup, dll.
		return utils.GenericError(c, fiber.StatusBadRequest, err.Error())
	}

	// 4. Siapkan dan panggil Midtrans Snap (interaksi dengan layanan eksternal).
	var s = snap.Client{}
	s.New(h.cfg.MidtransServerKey, midtrans.Sandbox)

	// Ambil detail pengguna untuk Midtrans
	user, _ := h.userRepo.FindByID(userID)

	snapReq := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  fmt.Sprintf("NGABACA-%s-%d", order.ID.String(), time.Now().Unix()),
			GrossAmt: int64(totalPrice),
		},
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

	// 5. Kembalikan respons dari Midtrans ke frontend.
	return c.JSON(snapRes)

}

type CreateReviewRequest struct {
	Rating  int    `json:"rating" validate:"required,min=1,max=5"`
	Comment string `json:"comment" validate:"omitempty,max=500"`
}

func (h *CustomerHandler) CreateReview(c *fiber.Ctx) error {
	req := new(CreateReviewRequest) // Anda mungkin perlu memindahkan struct ini juga
	if err := c.BodyParser(req); err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid request body")
	}
	if errs := utils.ValidateStruct(req); errs != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": errs})
	}

	bookID, _ := uuid.Parse(c.Params("id"))
	userClaims := c.Locals("user").(jwt.MapClaims)
	userID, _ := uuid.Parse(userClaims["user_id"].(string))

	exists, err := h.reviewRepo.CheckExisting(userID, bookID)
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Error checking review")
	}
	if exists {
		return utils.GenericError(c, fiber.StatusConflict, "You have already reviewed this book")
	}

	review := model.Review{
		Rating:  req.Rating,
		Comment: req.Comment,
		BookID:  bookID,
		UserID:  userID,
	}

	if err := h.reviewRepo.Create(&review); err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Failed to create review")
	}
	return c.Status(fiber.StatusCreated).JSON(review)
}

// GetBookReviews untuk mengambil semua ulasan dari sebuah buku.
func (h *CustomerHandler) GetBookReviews(c *fiber.Ctx) error {
	bookID, _ := uuid.Parse(c.Params("id"))
	reviews, err := h.reviewRepo.FindByBookID(bookID)
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Could not fetch reviews")
	}
	return c.JSON(reviews)
}

// Anda juga perlu memindahkan struct CreateReviewRequest ke file ini
