package handler

import (
	"ngabaca/internal/model"
	"ngabaca/internal/repository"
	"ngabaca/internal/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// =====================================================================
//             DTO (Data Transfer Objects) for Book Detail
// =====================================================================

// UserSummary hanya berisi data user yang relevan untuk ditampilkan di ulasan.
type UserSummary struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Avatar string    `json:"avatar"`
}

// ReviewDetail adalah format ulasan yang akan ditampilkan di response.
type ReviewDetail struct {
	ID        uuid.UUID   `json:"id"`
	CreatedAt time.Time   `json:"created_at"`
	Rating    int         `json:"rating"`
	Comment   string      `json:"comment"`
	User      UserSummary `json:"user"`
}

// CategorySummary adalah format kategori yang disederhanakan.
type CategorySummary struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

// BookDetailResponse adalah struct utama untuk respons JSON.
type BookDetailResponse struct {
	ID            uuid.UUID       `json:"ID"`
	Title         string          `json:"title"`
	Slug          string          `json:"slug"`
	PublishedYear int             `json:"published_year"`
	CoverImageURL string          `json:"cover_image_url"`
	Author        string          `json:"author"`
	Description   string          `json:"description"`
	Price         float64         `json:"price"`
	Stock         int             `json:"stock"`
	AvgRating     float64         `json:"avg_rating"`
	ReviewCount   int             `json:"review_count"`
	Category      CategorySummary `json:"category"`
	Reviews       []ReviewDetail  `json:"reviews"`
}

type PublicHandler struct {
	bookRepo     repository.BookRepository
	categoryRepo repository.CategoryRepository
}

func NewPublicHandler(bookRepo repository.BookRepository, categoryRepo repository.CategoryRepository) *PublicHandler {
	return &PublicHandler{
		bookRepo:     bookRepo,
		categoryRepo: categoryRepo,
	}
}

func (h *PublicHandler) Home(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Selamat datang di API Ngabaca!",
	})
}

func (h *PublicHandler) GetCategories(c *fiber.Ctx) error {
	categories, err := h.categoryRepo.FindAll()
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Could not fetch categories")
	}
	return c.JSON(categories)
}

func (h *PublicHandler) GetCategoryByID(c *fiber.Ctx) error {
	id := c.Params("id")
	category, err := h.categoryRepo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.GenericError(c, fiber.StatusNotFound, "Category not found")
		}
		return utils.GenericError(c, fiber.StatusInternalServerError, "Database error")
	}
	return c.JSON(category)
}

func (h *PublicHandler) GetBooks(c *fiber.Ctx) error {
	books, err := h.bookRepo.FindAll()
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Could not fetch books")
	}
	return c.JSON(books)
}

func (h *PublicHandler) GetBookDetail(c *fiber.Ctx) error {
	slug := c.Params("slug")

	// 1. Panggil repository untuk mendapatkan data lengkap dari DB (tidak ada perubahan di sini)
	book, err := h.bookRepo.FindBySlug(slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.GenericError(c, fiber.StatusNotFound, "Book not found")
		}
		return utils.GenericError(c, fiber.StatusInternalServerError, "Database error")
	}

	// 2. Buat slice untuk menampung data review yang sudah diformat
	reviewResponses := make([]ReviewDetail, 0)

	// 3. Looping melalui hasil dari database dan map ke struct ReviewDetail
	for _, review := range book.Reviews {
		reviewResponses = append(reviewResponses, ReviewDetail{
			ID:        review.ID,
			CreatedAt: review.CreatedAt,
			Rating:    review.Rating,
			Comment:   review.Comment,
			User: UserSummary{
				ID:     review.User.ID,
				Name:   review.User.Name,
				Avatar: review.User.Avatar,
			},
		})
	}

	// 4. Susun respons akhir menggunakan struct BookDetailResponse
	response := BookDetailResponse{
		ID:            book.ID,
		Title:         book.Title,
		Slug:          book.Slug,
		Author:        book.Author,
		Description:   book.Description,
		Price:         book.Price,
		Stock:         book.Stock,
		PublishedYear: book.PublishedYear,
		CoverImageURL: book.CoverImageURL,
		AvgRating:     book.AvgRating,
		ReviewCount:   book.ReviewCount,
		Category: CategorySummary{
			ID:   book.Category.ID,
			Name: book.Category.Name,
			Slug: book.Category.Slug,
		},
		Reviews: reviewResponses, // Gunakan slice yang sudah kita format
	}

	// 5. Kirim DTO sebagai JSON, bukan model GORM asli
	return c.JSON(response)
}

func (h *PublicHandler) SearchBooks(c *fiber.Ctx) error {
	keyword := c.Query("q")
	if keyword == "" {
		return c.JSON([]model.Book{})
	}

	books, err := h.bookRepo.Search(keyword)
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Database query error")
	}
	return c.JSON(books)
}
