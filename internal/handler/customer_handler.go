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
	userRepo     repository.UserRepository
	orderService service.OrderService
	reviewRepo   repository.ReviewRepository
	wishlistRepo repository.WishlistRepository
	cartRepo     repository.CartRepository
	cfg          config.Config
}

// NewCustomerHandler adalah constructor untuk CustomerHandler.
func NewCustomerHandler(
	orderRepo repository.OrderRepository,
	userRepo repository.UserRepository,
	orderService service.OrderService,
	reviewRepo repository.ReviewRepository,
	wishlistRepo repository.WishlistRepository,
	cartRepo repository.CartRepository,
	cfg config.Config,
) *CustomerHandler {
	return &CustomerHandler{
		orderRepo:    orderRepo,
		userRepo:     userRepo,
		reviewRepo:   reviewRepo,
		orderService: orderService,
		wishlistRepo: wishlistRepo,
		cartRepo:     cartRepo,
		cfg:          cfg,
	}
}

// Struct untuk request body AddToWishlist
type AddToWishlistRequest struct {
	BookID uuid.UUID `json:"book_id" validate:"required"`
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

// Method baru untuk menambahkan buku ke wishlist
func (h *CustomerHandler) AddToWishlist(c *fiber.Ctx) error {
	req := new(AddToWishlistRequest)
	if err := c.BodyParser(req); err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid request body")
	}
	if errs := utils.ValidateStruct(req); errs != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": errs})
	}

	userClaims := c.Locals("user").(jwt.MapClaims)
	userID, _ := uuid.Parse(userClaims["user_id"].(string))

	// Cek apakah sudah ada di wishlist
	exists, err := h.wishlistRepo.Check(userID, req.BookID)
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Error checking wishlist")
	}
	if exists {
		return utils.GenericError(c, fiber.StatusConflict, "Book already in wishlist")
	}

	wishlistItem := &model.Wishlist{
		UserID: userID,
		BookID: req.BookID,
	}

	if err := h.wishlistRepo.Create(wishlistItem); err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Failed to add book to wishlist")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Book added to wishlist successfully"})
}

// Method baru untuk mendapatkan isi wishlist
func (h *CustomerHandler) GetMyWishlist(c *fiber.Ctx) error {
	userClaims := c.Locals("user").(jwt.MapClaims)
	userID, _ := uuid.Parse(userClaims["user_id"].(string))

	wishlistItems, err := h.wishlistRepo.FindByUserID(userID)
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Could not fetch wishlist")
	}

	// Buat DTO agar respons lebih bersih (hanya menampilkan data buku)
	type BookResponse struct {
		ID            string `json:"ID"`
		Title         string `json:"title"`
		Author        string `json:"author"`
		CoverImageURL string `json:"cover_image_url"`
	}

	response := make([]BookResponse, len(wishlistItems))
	for i, item := range wishlistItems {
		response[i] = BookResponse{
			ID:            item.Book.ID.String(),
			Title:         item.Book.Title,
			Author:        item.Book.Author,
			CoverImageURL: item.Book.CoverImageURL,
		}
	}

	return c.JSON(response)
}

// Method baru untuk menghapus buku dari wishlist
func (h *CustomerHandler) RemoveFromWishlist(c *fiber.Ctx) error {
	bookID, err := uuid.Parse(c.Params("bookId"))
	if err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid Book ID format")
	}

	userClaims := c.Locals("user").(jwt.MapClaims)
	userID, _ := uuid.Parse(userClaims["user_id"].(string))

	if err := h.wishlistRepo.Delete(userID, bookID); err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Failed to remove book from wishlist")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// DTO untuk response Cart
type CartResponse struct {
	ID     uuid.UUID          `json:"id"`
	UserID uuid.UUID          `json:"userId"`
	Items  []CartItemResponse `json:"items"`
}

type BookResponses struct {
	ID            uuid.UUID `json:"id"`
	Title         string    `json:"title"`
	Author        string    `json:"author"`
	Price         float64   `json:"price"`
	CoverImageURL string    `json:"cover_image_url"`
	Stock         int       `json:"stock"`
}

type CartItemResponse struct {
	ID       uuid.UUID     `json:"id"`
	Quantity int           `json:"quantity"`
	Book     BookResponses `json:"book"`
}

// Struct untuk request body AddToCart
type AddToCartRequest struct {
	BookID   uuid.UUID `json:"book_id" validate:"required"`
	Quantity int       `json:"quantity" validate:"required,min=1"`
}

// Method baru untuk mendapatkan isi keranjang
func (h *CustomerHandler) GetMyCart(c *fiber.Ctx) error {
	userClaims := c.Locals("user").(jwt.MapClaims)
	userID, _ := uuid.Parse(userClaims["user_id"].(string))

	cart, err := h.cartRepo.GetCartByUserID(userID)
	if err != nil {
		return utils.GenericError(c, fiber.StatusNotFound, "Could not fetch cart")
	}

	// Mapping ke DTO
	response := CartResponse{
		ID:     cart.ID,
		UserID: cart.UserID,
		Items:  []CartItemResponse{},
	}

	for _, item := range cart.CartItems {
		book := item.Book

		bookResp := BookResponses{
			ID:            book.ID,
			Title:         book.Title,
			Author:        book.Author,
			Price:         book.Price,
			CoverImageURL: book.CoverImageURL,
			Stock:         book.Stock,
		}

		itemResp := CartItemResponse{
			ID:       item.ID,
			Quantity: item.Quantity,
			Book:     bookResp,
		}

		response.Items = append(response.Items, itemResp)
	}

	return c.JSON(response)
}

// Method baru untuk menambahkan item ke keranjang
func (h *CustomerHandler) AddToCart(c *fiber.Ctx) error {
	req := new(AddToCartRequest)
	if err := c.BodyParser(req); err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid request body")
	}
	if errs := utils.ValidateStruct(req); errs != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": errs})
	}

	userClaims := c.Locals("user").(jwt.MapClaims)
	userID, _ := uuid.Parse(userClaims["user_id"].(string))

	err := h.cartRepo.AddItem(userID, req.BookID, req.Quantity)
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Item added to cart successfully"})
}

// Method baru untuk memperbarui item di keranjang
func (h *CustomerHandler) UpdateCartItem(c *fiber.Ctx) error {
	itemID, err := uuid.Parse(c.Params("itemId"))
	if err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid Item ID format")
	}

	req := new(AddToCartRequest) // Kita bisa pakai struct yang sama untuk update
	if err := c.BodyParser(req); err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid request body")
	}
	if errs := utils.ValidateStruct(req); errs != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": errs})
	}

	userClaims := c.Locals("user").(jwt.MapClaims)
	userID, _ := uuid.Parse(userClaims["user_id"].(string))

	err = h.cartRepo.UpdateCartItem(userID, itemID, req.Quantity)
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Failed to update cart item")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Cart item updated successfully"})
}

func (h *CustomerHandler) RemoveFromCart(c *fiber.Ctx) error {
	bookID, err := uuid.Parse(c.Params("bookId"))
	if err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid Book ID format")
	}

	userClaims := c.Locals("user").(jwt.MapClaims)
	userID, _ := uuid.Parse(userClaims["user_id"].(string))

	if err := h.cartRepo.RemoveItem(userID, bookID); err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Failed to remove item from cart")
	}

	return c.SendStatus(fiber.StatusNoContent)

}

type SyncCartRequest struct {
	Items []struct {
		BookID   uuid.UUID `json:"book_id" validate:"required"`
		Quantity int       `json:"quantity" validate:"required,min=1"`
	} `json:"items" validate:"required,dive"`
}

// Method baru untuk sinkronisasi keranjang
func (h *CustomerHandler) SyncCart(c *fiber.Ctx) error {
	req := new(SyncCartRequest)
	if err := c.BodyParser(req); err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid request body")
	}
	if errs := utils.ValidateStruct(req); errs != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": errs})
	}

	userClaims := c.Locals("user").(jwt.MapClaims)
	userID, _ := uuid.Parse(userClaims["user_id"].(string))

	// Looping setiap item dari cart lokal dan tambahkan ke cart di database
	// Repository AddItem kita sudah pintar menangani item yang sudah ada (akan menambah quantity)
	for _, item := range req.Items {
		err := h.cartRepo.AddItem(userID, item.BookID, item.Quantity)
		if err != nil {
			// Lanjutkan meski ada error di satu item, atau bisa juga dibatalkan semua
			fmt.Printf("Warning: could not sync item %s for user %s: %v\n", item.BookID, userID, err)
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Cart synchronized successfully"})
}
