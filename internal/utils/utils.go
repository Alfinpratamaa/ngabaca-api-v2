package utils

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// GenerateSlug membuat slug yang URL-friendly dari sebuah string.
func GenerateSlug(title string) string {
	// Ganti semua karakter non-alphanumeric dengan tanda hubung
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug := re.ReplaceAllString(strings.ToLower(title), "-")
	// Hapus tanda hubung di awal atau akhir
	slug = strings.Trim(slug, "-")
	return slug
}

// ErrorResponse merepresentasikan format error validasi.
type ErrorResponse struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Value string `json:"value,omitempty"`
}

// ValidateStruct memvalidasi struct dan mengembalikan error jika ada.
func ValidateStruct(s interface{}) []*ErrorResponse {
	var errors []*ErrorResponse
	validate := validator.New()
	err := validate.Struct(s)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element ErrorResponse
			element.Field = err.StructNamespace()
			element.Tag = err.Tag()
			element.Value = err.Param()
			errors = append(errors, &element)
		}
	}
	return errors
}

// GenericError adalah helper untuk mengirim response error standar.
func GenericError(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{"error": message})
}

func DefaultString(val, fallback string) string {
	if val != "" {
		return val
	}
	return fallback
}
