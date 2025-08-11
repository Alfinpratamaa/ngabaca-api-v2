package handler

import (
	"ngabaca/internal/model"
	"ngabaca/internal/repository"
	"ngabaca/internal/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

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

func (h *PublicHandler) GetBooks(c *fiber.Ctx) error {
	books, err := h.bookRepo.FindAll()
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Could not fetch books")
	}
	return c.JSON(books)
}

func (h *PublicHandler) GetBookDetail(c *fiber.Ctx) error {
	slug := c.Params("slug")
	book, err := h.bookRepo.FindBySlug(slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.GenericError(c, fiber.StatusNotFound, "Book not found")
		}
		return utils.GenericError(c, fiber.StatusInternalServerError, "Database error")
	}
	return c.JSON(book)
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
