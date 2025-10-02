package handler

import (
	"context"
	"encoding/json"
	"errors"
	"ngabaca/config"
	"ngabaca/internal/model"
	"ngabaca/internal/repository"
	"ngabaca/internal/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
	oauth2v2 "google.golang.org/api/oauth2/v2"
	"gorm.io/gorm"
)

type AuthHandler struct {
	userRepo repository.UserRepository
	cfg      config.Config
}

func NewAuthHandler(userRepo repository.UserRepository, cfg config.Config) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, cfg: cfg}
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

// ========== UTIL ==========

func (h *AuthHandler) signAppJWT(user *model.User, dur time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":     user.ID.String(),
		"user_id": user.ID.String(),
		"role":    user.Role,
		"iat":     now.Unix(),
		"exp":     now.Add(dur).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.cfg.JWTSecret))
}

// Verifikasi idToken dari Google untuk beberapa audience (Android/iOS/Web)
func (h *AuthHandler) verifyGoogleIDToken(ctx context.Context, idTok string) (*idtoken.Payload, error) {
	// daftar clientID yang sah. Simpan di config kamu.
	auds := []string{
		h.cfg.GoogleAndroidClientID,
		h.cfg.GoogleIOSClientID,
		h.cfg.GoogleClientID,
		h.cfg.GoogleExpoClientID,
	}
	var lastErr error
	for _, aud := range auds {
		if aud == "" {
			continue
		}
		pl, err := idtoken.Validate(ctx, idTok, aud)
		if err == nil {
			return pl, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = errors.New("no google audiences configured")
	}
	return nil, lastErr
}

// ========== HANDLERS ==========

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

	t, err := h.signAppJWT(&user, 24*time.Hour)
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Failed to create token")
	}
	return c.JSON(fiber.Map{"token": t})
}

// ======= FLOW WEB (tetap ada) =======
// GoogleLogin & GoogleCallback kamu biarkan seperti semula…

// ======= FLOW MOBILE/EXPO =======

type googleMobileReq struct {
	IDToken string `json:"idToken"` // dari Expo (response.authentication.idToken)
}

func (h *AuthHandler) GoogleMobileSignIn(c *fiber.Ctx) error {
	var req googleMobileReq
	if err := json.Unmarshal(c.Body(), &req); err != nil || req.IDToken == "" {
		return utils.GenericError(c, fiber.StatusBadRequest, "idToken is required")
	}

	// 1) Verifikasi idToken ke Google
	pl, err := h.verifyGoogleIDToken(c.Context(), req.IDToken)
	if err != nil {
		return utils.GenericError(c, fiber.StatusUnauthorized, "Invalid Google ID token: "+err.Error())
	}

	// 2) Ambil claims penting
	// Catatan: kunci claim umum: "sub", "email", "name", "picture", "email_verified"
	sub, _ := pl.Claims["sub"].(string)
	email, _ := pl.Claims["email"].(string)
	name, _ := pl.Claims["name"].(string)
	picture, _ := pl.Claims["picture"].(string)
	emailVerified, _ := pl.Claims["email_verified"].(bool)

	if email == "" {
		// sebagian akun enterprise bisa restrict email. Kamu bisa fallback ke sub sebagai identifier.
		return utils.GenericError(c, fiber.StatusBadRequest, "Email not provided by Google")
	}

	// 3) Find or create user
	user, err := h.userRepo.FindOrCreateFromGoogleClaims(repository.GoogleClaims{
		Sub:            sub,
		Email:          email,
		Name:           name,
		Picture:        picture,
		EmailVerified:  emailVerified,
		DefaultAppRole: "pelanggan",
	})
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Database transaction error: "+err.Error())
	}

	// 4) Buat JWT lokal
	t, err := h.signAppJWT(&user, 72*time.Hour)
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Failed to create local JWT")
	}

	// (Opsional) kirim juga info user dasar buat ditampilkan
	return c.JSON(fiber.Map{
		"token": t,
		"user": fiber.Map{
			"id":      user.ID.String(),
			"name":    user.Name,
			"email":   user.Email,
			"role":    user.Role,
			"picture": picture,
		},
	})
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
