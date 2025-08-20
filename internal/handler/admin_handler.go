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

	// Ambil data lama dari DB
	book, err := h.bookRepo.FindByID(bookID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.GenericError(c, fiber.StatusNotFound, "Book not found")
		}
		return utils.GenericError(c, fiber.StatusInternalServerError, "Database error")
	}

	var (
		title, author, description, coverURL, categoryIDStr string
		price                                               float64
		stock, publishedYear                                int
		categoryUUID                                        uuid.UUID
	)

	contentType := c.Get("Content-Type")
	if strings.Contains(contentType, "multipart/form-data") {
		// Parsing dari form-data
		title = c.FormValue("title", book.Title)
		author = c.FormValue("author", book.Author)
		description = c.FormValue("description", book.Description)

		priceStr := c.FormValue("price")
		if priceStr != "" {
			price, err = strconv.ParseFloat(priceStr, 64)
			if err != nil {
				return utils.GenericError(c, fiber.StatusBadRequest, "Invalid price format")
			}
		} else {
			price = book.Price
		}

		stockStr := c.FormValue("stock")
		if stockStr != "" {
			stock, _ = strconv.Atoi(stockStr)
		} else {
			stock = book.Stock
		}

		publishedYearStr := c.FormValue("published_year")
		if publishedYearStr != "" {
			publishedYear, _ = strconv.Atoi(publishedYearStr)
		} else {
			publishedYear = book.PublishedYear
		}

		categoryIDStr = c.FormValue("category_id", book.CategoryID.String())

		// File cover opsional
		file, _ := c.FormFile("cover_image")
		if file != nil {
			openedFile, _ := file.Open()
			fileBytes, _ := io.ReadAll(openedFile)
			uploadedURL, upErr := utils.UploadToImageKit(h.cfg, fileBytes, file.Filename, "covers")
			if upErr != nil {
				return utils.GenericError(c, fiber.StatusInternalServerError, "Image upload failed: "+upErr.Error())
			}
			// Hapus cover lama kalau ada
			if book.CoverImageURL != "" {
				_ = utils.DeleteFromImageKit(h.cfg, book.CoverImageURL)
			}
			coverURL = uploadedURL
			openedFile.Close()
		} else {
			coverURL = book.CoverImageURL
		}

	} else { // application/json
		req := new(model.Book)
		if err := c.BodyParser(req); err != nil {
			return utils.GenericError(c, fiber.StatusBadRequest, "Invalid request body")
		}
		title = utils.DefaultString(req.Title, book.Title)
		author = utils.DefaultString(req.Author, book.Author)
		description = utils.DefaultString(req.Description, book.Description)
		if req.Price != 0 {
			price = req.Price
		} else {
			price = book.Price
		}
		if req.Stock != 0 {
			stock = req.Stock
		} else {
			stock = book.Stock
		}
		if req.PublishedYear != 0 {
			publishedYear = req.PublishedYear
		} else {
			publishedYear = book.PublishedYear
		}
		if req.CategoryID != uuid.Nil {
			categoryIDStr = req.CategoryID.String()
		} else {
			categoryIDStr = book.CategoryID.String()
		}
		if req.CoverImageURL != "" {
			coverURL = req.CoverImageURL
		} else {
			coverURL = book.CoverImageURL
		}
	}

	categoryUUID, err = uuid.Parse(categoryIDStr)
	if err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid category_id format")
	}

	// Cek slug kalau judul berubah
	if title != book.Title {
		baseSlug := utils.GenerateSlug(title)
		slug := baseSlug
		i := 1
		for {
			exists, err := h.bookRepo.IsSlugExist(slug, book.ID)
			if err != nil {
				return utils.GenericError(c, fiber.StatusInternalServerError, "Error checking slug existence")
			}
			if !exists {
				book.Slug = slug
				break
			}
			slug = baseSlug + "-" + strconv.Itoa(i)
			i++
		}
	}

	// Update field buku
	book.Title = title
	book.Author = author
	book.Description = description
	book.Price = price
	book.Stock = stock
	book.PublishedYear = publishedYear
	book.CoverImageURL = coverURL
	book.CategoryID = categoryUUID

	// Simpan ke DB
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
