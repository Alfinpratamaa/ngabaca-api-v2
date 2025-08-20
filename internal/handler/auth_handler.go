package handler

import (
	"context"
	"ngabaca/config"
	"ngabaca/internal/model"
	"ngabaca/internal/repository"
	"ngabaca/internal/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	oauth2v2 "google.golang.org/api/oauth2/v2"
	"gorm.io/gorm"
)

type AuthHandler struct {
	userRepo repository.UserRepository
	cfg      config.Config
}

var expiredTime = time.Now().Add(time.Hour * 24).Unix()

func NewAuthHandler(userRepo repository.UserRepository, cfg config.Config) *AuthHandler {
	return &AuthHandler{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=3"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	req := new(RegisterRequest)
	if err := c.BodyParser(req); err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid request body")
	}
	if errs := utils.ValidateStruct(req); errs != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": errs})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Failed to hash password")
	}

	user := &model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     "pelanggan",
	}

	// REFACTOR: Panggil repository untuk membuat user
	createdUser, err := h.userRepo.Create(user)
	if err != nil {
		return utils.GenericError(c, fiber.StatusConflict, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(createdUser)
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	req := new(LoginRequest)
	if err := c.BodyParser(req); err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid request body")
	}
	if errs := utils.ValidateStruct(req); errs != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": errs})
	}

	// REFACTOR: Panggil repository untuk mencari user
	user, err := h.userRepo.FindByEmail(req.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.GenericError(c, fiber.StatusUnauthorized, "Invalid email or password")
		}
		return utils.GenericError(c, fiber.StatusInternalServerError, "Database error")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return utils.GenericError(c, fiber.StatusUnauthorized, "Invalid email or password")
	}

	// Buat JWT Claims
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"role":    user.Role,
		"exp":     expiredTime,
	}

	// Buat token menggunakan config dari handler
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Failed to create token")
	}

	return c.JSON(fiber.Map{"token": t})
}

func (h *AuthHandler) GoogleLogin(c *fiber.Ctx) error {
	url := utils.GoogleOAuthConfig.AuthCodeURL("state-string")
	return c.Redirect(url, fiber.StatusTemporaryRedirect)
}

func (h *AuthHandler) GoogleCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return utils.GenericError(c, fiber.StatusBadRequest, "Authorization code not provided")
	}

	token, err := utils.GoogleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Failed to exchange token: "+err.Error())
	}

	client := utils.GoogleOAuthConfig.Client(context.Background(), token)
	service, err := oauth2v2.New(client)
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Failed to create Google service client")
	}
	userInfo, err := service.Userinfo.Get().Do()
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Failed to get user info")
	}

	// REFACTOR: Panggil repository untuk logika find-or-create
	user, err := h.userRepo.FindOrCreateByGoogle(userInfo)
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Database transaction error: "+err.Error())
	}

	// Buat token JWT lokal (logika sama seperti Login)
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := jwtToken.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Failed to create local JWT")
	}

	return c.JSON(fiber.Map{"token": t})
}
