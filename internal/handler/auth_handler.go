package handler

import (
	"ngabaca/config"
	"ngabaca/database"
	"ngabaca/internal/model"
	"ngabaca/internal/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// RegisterRequest adalah struct untuk menampung data registrasi.
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=3"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginRequest adalah struct untuk menampung data login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Register untuk mendaftarkan pengguna baru.
func Register(c *fiber.Ctx) error {
	req := new(RegisterRequest)

	// Parse body request
	if err := c.BodyParser(req); err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Validasi input
	if errs := utils.ValidateStruct(req); errs != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": errs})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Failed to hash password")
	}

	// Buat user baru
	user := model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     "pelanggan", // Default role
	}

	// Simpan ke database
	if err := database.DB.Create(&user).Error; err != nil {
		return utils.GenericError(c, fiber.StatusConflict, "Email already exists")
	}

	// Hapus password dari response
	user.Password = ""

	return c.Status(fiber.StatusCreated).JSON(user)
}

// Login untuk mengautentikasi pengguna dan memberikan token JWT.
func Login(c *fiber.Ctx) error {
	req := new(LoginRequest)

	if err := c.BodyParser(req); err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if errs := utils.ValidateStruct(req); errs != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": errs})
	}

	var user model.User
	// Cari user berdasarkan email
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.GenericError(c, fiber.StatusUnauthorized, "Invalid email or password")
		}
		return utils.GenericError(c, fiber.StatusInternalServerError, "Database error")
	}

	// Bandingkan password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return utils.GenericError(c, fiber.StatusUnauthorized, "Invalid email or password")
	}

	// Muat konfigurasi untuk JWT Secret
	cfg, err := config.LoadConfig(".")
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Failed to load config")
	}

	// Buat JWT Claims
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 72).Unix(), // Token berlaku 72 jam
	}

	// Buat token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Failed to create token")
	}

	return c.JSON(fiber.Map{"token": t})
}

// GoogleLogin - Placeholder
func GoogleLogin(c *fiber.Ctx) error {
	// TODO: Implementasi redirect ke halaman login Google
	return c.SendString("Redirect to Google Login")
}

// GoogleCallback - Placeholder
func GoogleCallback(c *fiber.Ctx) error {
	// TODO: Implementasi callback dari Google, tukar kode dengan token,
	// lalu buat atau login user dan generate JWT lokal.
	return c.SendString("Handling Google Callback")
}
