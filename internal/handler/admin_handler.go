package handler

import (
	"ngabaca/database"
	"ngabaca/internal/model"
	"ngabaca/internal/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type BookRequest struct {
	Title         string  `json:"title" validate:"required"`
	Author        string  `json:"author" validate:"required"`
	Description   string  `json:"description"`
	Price         float64 `json:"price" validate:"required,gt=0"`
	Stock         int     `json:"stock" validate:"required,gte=0"`
	PublishedYear int     `json:"published_year"`
	CoverImageURL string  `json:"cover_image_url" validate:"omitempty,url"`
	CategoryID    uint    `json:"category_id" validate:"required"`
}

// AdminGetBooks mengambil semua buku (untuk admin).
func AdminGetBooks(c *fiber.Ctx) error {
	return GetBooks(c) // Bisa menggunakan handler publik yang sama
}

// AdminGetBook mengambil detail satu buku (untuk admin).
func AdminGetBook(c *fiber.Ctx) error {
	id := c.Params("id")
	var book model.Book
	if err := database.DB.Preload("Category").First(&book, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.GenericError(c, fiber.StatusNotFound, "Book not found")
		}
		return utils.GenericError(c, fiber.StatusInternalServerError, "Database error")
	}
	return c.JSON(book)
}

// AdminCreateBook membuat buku baru.
func AdminCreateBook(c *fiber.Ctx) error {
	req := new(BookRequest)

	if err := c.BodyParser(req); err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid request body")
	}
	if errs := utils.ValidateStruct(req); errs != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": errs})
	}

	slug := utils.GenerateSlug(req.Title)
	// TODO: Cek keunikan slug, tambahkan angka jika sudah ada.

	book := model.Book{
		Title:         req.Title,
		Slug:          slug,
		Author:        req.Author,
		Description:   req.Description,
		Price:         req.Price,
		Stock:         req.Stock,
		PublishedYear: req.PublishedYear,
		CoverImageURL: req.CoverImageURL,
		CategoryID:    req.CategoryID,
	}

	if err := database.DB.Create(&book).Error; err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Failed to create book")
	}

	return c.Status(fiber.StatusCreated).JSON(book)
}

// AdminUpdateBook memperbarui data buku.
func AdminUpdateBook(c *fiber.Ctx) error {
	id := c.Params("id")
	req := new(BookRequest)

	if err := c.BodyParser(req); err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid request body")
	}
	if errs := utils.ValidateStruct(req); errs != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": errs})
	}

	var book model.Book
	if err := database.DB.First(&book, id).Error; err != nil {
		return utils.GenericError(c, fiber.StatusNotFound, "Book not found")
	}

	// Update data
	book.Title = req.Title
	book.Slug = utils.GenerateSlug(req.Title) // Regenerate slug jika judul berubah
	book.Author = req.Author
	book.Description = req.Description
	book.Price = req.Price
	book.Stock = req.Stock
	book.PublishedYear = req.PublishedYear
	book.CoverImageURL = req.CoverImageURL
	book.CategoryID = req.CategoryID

	if err := database.DB.Save(&book).Error; err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Failed to update book")
	}

	return c.JSON(book)
}

// AdminDeleteBook menghapus buku.
func AdminDeleteBook(c *fiber.Ctx) error {
	id := c.Params("id")
	result := database.DB.Delete(&model.Book{}, id)

	if result.Error != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Failed to delete book")
	}
	if result.RowsAffected == 0 {
		return utils.GenericError(c, fiber.StatusNotFound, "Book not found")
	}

	return c.SendStatus(fiber.StatusNoContent)
}
