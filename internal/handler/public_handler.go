package handler

import (
	"ngabaca/database"
	"ngabaca/internal/model"
	"ngabaca/internal/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Home adalah handler untuk halaman utama.
func Home(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Selamat datang di API Ngabaca!",
	})
}

// GetCategories mengambil semua kategori.
func GetCategories(c *fiber.Ctx) error {
	var categories []model.Category
	if err := database.DB.Find(&categories).Error; err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Could not fetch categories")
	}
	return c.JSON(categories)
}

// GetBooks mengambil daftar buku dengan paginasi.
func GetBooks(c *fiber.Ctx) error {
	var books []model.Book
	// Ambil semua buku dan preload kategori terkait
	result := database.DB.Preload("Category").Find(&books)
	if result.Error != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Could not fetch books")
	}
	return c.JSON(books)
}

// GetBookDetail mengambil detail satu buku berdasarkan slug.
func GetBookDetail(c *fiber.Ctx) error {
	slug := c.Params("slug")
	var book model.Book

	// Cari buku berdasarkan slug dan preload data kategori
	err := database.DB.Preload("Category").Where("slug = ?", slug).First(&book).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.GenericError(c, fiber.StatusNotFound, "Book not found")
		}
		return utils.GenericError(c, fiber.StatusInternalServerError, "Database error")
	}
	return c.JSON(book)
}
