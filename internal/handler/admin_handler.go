package handler

import (
	"io"
	"ngabaca/config"
	"ngabaca/internal/model"
	"ngabaca/internal/repository"
	"ngabaca/internal/utils"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BookData struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Slug          string  `json:"slug"`
	Author        string  `json:"author"`
	Price         float64 `json:"price"`
	Stock         int     `json:"stock"`
	PublishedYear int     `json:"published_year"`
	CoverImageURL string  `json:"cover_image_url"`
	CategoryID    string  `json:"category_id"`
}

type BookResponse struct {
	Message string   `json:"message"`
	Book    BookData `json:"book"`
}
type AdminHandler struct {
	bookRepo  repository.BookRepository
	userRepo  repository.UserRepository
	orderRepo repository.OrderRepository
	cfg       config.Config
}

func NewAdminHandler(bookRepo repository.BookRepository, userRepo repository.UserRepository, orderRepo repository.OrderRepository, cfg config.Config) *AdminHandler {
	return &AdminHandler{
		bookRepo:  bookRepo,
		userRepo:  userRepo,
		orderRepo: orderRepo,
		cfg:       cfg,
	}
}

// AdminGetBooks sekarang adalah method dari AdminHandler.
func (h *AdminHandler) AdminGetBooks(c *fiber.Ctx) error {
	books, err := h.bookRepo.FindAll()
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Could not fetch books")
	}
	return c.JSON(books)
}

func (h *AdminHandler) AdminGetBook(c *fiber.Ctx) error {
	bookID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid ID format")
	}

	book, err := h.bookRepo.FindByID(bookID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.GenericError(c, fiber.StatusNotFound, "Book not found")
		}
		return utils.GenericError(c, fiber.StatusInternalServerError, "Database error")
	}
	return c.JSON(book)
}

func (h *AdminHandler) AdminCreateBook(c *fiber.Ctx) error {
	// Variabel untuk menampung nilai input
	var (
		title, author, description, coverURL, categoryIDStr string
		price                                               float64
		stock, publishedYear                                int
		categoryUUID                                        uuid.UUID
		err                                                 error
	)

	// Logika parsing request tetap di handler, ini sudah benar.
	contentType := c.Get("Content-Type")
	if strings.Contains(contentType, "multipart/form-data") {
		// Parsing dari form-data
		title = c.FormValue("title")
		author = c.FormValue("author")
		description = c.FormValue("description")
		priceStr := c.FormValue("price")
		stockStr := c.FormValue("stock")
		publishedYearStr := c.FormValue("published_year")
		categoryIDStr = c.FormValue("category_id")

		// Validasi & konversi tipe data
		price, err = strconv.ParseFloat(priceStr, 64)
		if err != nil {
			return utils.GenericError(c, fiber.StatusBadRequest, "Invalid price format")
		}
		// ... (validasi dan konversi lain sama seperti kode Anda) ...
		stock, _ = strconv.Atoi(stockStr)
		publishedYear, _ = strconv.Atoi(publishedYearStr)

		// File upload opsional (Logika layanan eksternal tetap di handler)
		file, _ := c.FormFile("cover_image")
		if file != nil {
			openedFile, _ := file.Open()
			fileBytes, _ := io.ReadAll(openedFile)
			uploadedURL, upErr := utils.UploadToImageKit(h.cfg, fileBytes, file.Filename, "covers")
			if upErr != nil {
				return utils.GenericError(c, fiber.StatusInternalServerError, "Image upload failed: "+upErr.Error())
			}
			coverURL = uploadedURL
			openedFile.Close()
		}

	} else { // Anggap application/json
		req := new(model.Book) // Menggunakan model.Book sementara untuk parsing
		if err := c.BodyParser(req); err != nil {
			return utils.GenericError(c, fiber.StatusBadRequest, "Invalid request body")
		}
		title = req.Title
		author = req.Author
		description = req.Description
		price = req.Price
		stock = req.Stock
		publishedYear = req.PublishedYear
		categoryIDStr = req.CategoryID.String()
		coverURL = req.CoverImageURL
	}

	categoryUUID, err = uuid.Parse(categoryIDStr)
	if err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid category_id format")
	}

	// REFACTOR: Logika pengecekan slug sekarang memanggil repository
	baseSlug := utils.GenerateSlug(title)
	slug := baseSlug
	i := 1
	for {
		exists, err := h.bookRepo.IsSlugExist(slug, uuid.Nil)
		if err != nil {
			return utils.GenericError(c, fiber.StatusInternalServerError, "Error checking slug existence")
		}
		if !exists {
			break
		}
		slug = baseSlug + "-" + strconv.Itoa(i)
		i++
	}

	book := &model.Book{
		Title:         title,
		Slug:          slug,
		Author:        author,
		Description:   description,
		Price:         price,
		Stock:         stock,
		PublishedYear: publishedYear,
		CoverImageURL: coverURL,
		CategoryID:    categoryUUID,
	}

	// REFACTOR: Panggil repository untuk menyimpan ke DB
	createdBook, err := h.bookRepo.Create(book)
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Failed to create book in database")
	}

	return c.Status(fiber.StatusCreated).JSON(createdBook)
}

// AdminUpdateBook memperbarui data buku.
func (h *AdminHandler) AdminUpdateBook(c *fiber.Ctx) error {
	bookID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid ID format")
	}

	// REFACTOR: Ambil data lama dari repository
	book, err := h.bookRepo.FindByID(bookID)
	if err != nil {
		return utils.GenericError(c, fiber.StatusNotFound, "Book not found")
	}

	// FIX: Logika update harusnya mirip dengan create, bisa menerima multipart/form-data
	title := c.FormValue("title")
	// ... (lakukan parsing semua field dari c.FormValue seperti di AdminCreateBook)

	// Jika judul berubah, cek slug baru
	if book.Title != title {
		// ... (logika slug check menggunakan h.bookRepo.IsSlugExist(newSlug, bookID))
	}

	// Cek apakah ada file cover baru
	file, _ := c.FormFile("cover_image")
	if file != nil {
		// Hapus cover lama dari ImageKit jika ada
		if book.CoverImageURL != "" {
			_ = utils.DeleteFromImageKit(h.cfg, book.CoverImageURL)
		}
		// Upload cover baru
		// ... (logika upload seperti di AdminCreateBook)
		// book.CoverImageURL = uploadedURL
	}

	// Update data lain
	book.Title = title
	// ...

	// REFACTOR: Panggil repository untuk update
	updatedBook, err := h.bookRepo.Update(&book)
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Failed to update book")
	}

	return c.JSON(updatedBook)
}

// AdminDeleteBook menghapus buku.
func (h *AdminHandler) AdminDeleteBook(c *fiber.Ctx) error {
	bookID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid ID format")
	}

	// REFACTOR & FIX: Ambil data buku DULU sebelum menghapus
	book, err := h.bookRepo.FindByID(bookID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.GenericError(c, fiber.StatusNotFound, "Book not found")
		}
		return utils.GenericError(c, fiber.StatusInternalServerError, "Database error")
	}

	// REFACTOR: Panggil repository untuk menghapus dari DB
	if err := h.bookRepo.Delete(&book); err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Failed to delete book from database")
	}

	// Hapus file dari ImageKit setelah berhasil hapus dari DB
	if book.CoverImageURL != "" {
		_ = utils.DeleteFromImageKit(h.cfg, book.CoverImageURL)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AdminHandler) AdminGetUsers(c *fiber.Ctx) error {
	users, err := h.userRepo.FindAll()
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Could not fetch users")
	}
	return c.JSON(users)
}

// UpdateUserRequest adalah struct untuk validasi permintaan update user.
type UpdateUserRequest struct {
	Role string `json:"role" validate:"required,oneof=admin pelanggan"`
}

// AdminUpdateUser untuk memperbarui data pengguna (misal: mengubah role).
func (h *AdminHandler) AdminUpdateUser(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid ID format")
	}

	req := new(UpdateUserRequest)
	if err := c.BodyParser(req); err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid request body")
	}
	if errs := utils.ValidateStruct(req); errs != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": errs})
	}

	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.GenericError(c, fiber.StatusNotFound, "User not found")
		}
		return utils.GenericError(c, fiber.StatusInternalServerError, "Database error")
	}

	user.Role = req.Role

	updatedUser, err := h.userRepo.Update(&user)
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Failed to update user")
	}

	return c.JSON(updatedUser)
}

// =====================================================================
// MANAJEMEN PESANAN UNTUK ADMIN
// =====================================================================

func (h *AdminHandler) AdminGetOrders(c *fiber.Ctx) error {
	status := c.Query("status")
	orders, err := h.orderRepo.FindAll(status)
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Could not fetch orders")
	}
	return c.JSON(orders)
}

// AdminGetOrderDetail sekarang memanggil repository.
func (h *AdminHandler) AdminGetOrderDetail(c *fiber.Ctx) error {
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid ID format")
	}

	order, err := h.orderRepo.FindByID(orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.GenericError(c, fiber.StatusNotFound, "Order not found")
		}
		return utils.GenericError(c, fiber.StatusInternalServerError, "Database error")
	}
	return c.JSON(order)
}

// UpdateOrderStatusRequest adalah struct untuk validasi permintaan update status.
type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending diproses dikirim selesai batal challenge"`
}

// AdminUpdateOrderStatus untuk mengubah status sebuah pesanan.
func (h *AdminHandler) AdminUpdateOrderStatus(c *fiber.Ctx) error {
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid ID format")
	}

	req := new(UpdateOrderStatusRequest)
	if err := c.BodyParser(req); err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid request body")
	}
	if errs := utils.ValidateStruct(req); errs != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": errs})
	}

	order, err := h.orderRepo.FindByID(orderID)
	if err != nil {
		return utils.GenericError(c, fiber.StatusNotFound, "Order not found")
	}

	order.Status = req.Status

	updatedOrder, err := h.orderRepo.Update(&order)
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Failed to update order status")
	}
	return c.JSON(updatedOrder)
}
