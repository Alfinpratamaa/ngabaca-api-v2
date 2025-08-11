package handler

import (
	"fmt"
	"ngabaca/config"
	"ngabaca/internal/repository"
	"ngabaca/internal/utils"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserHandler menampung semua dependency yang dibutuhkan untuk fitur profil pengguna.
type UserHandler struct {
	userRepo repository.UserRepository
	cfg      config.Config
}

// NewUserHandler adalah constructor untuk UserHandler.
func NewUserHandler(userRepo repository.UserRepository, cfg config.Config) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

// GetMyProfile mengambil data profil lengkap dari pengguna yang sedang login.
func (h *UserHandler) GetMyProfile(c *fiber.Ctx) error {
	userClaims := c.Locals("user").(jwt.MapClaims)
	userID, err := uuid.Parse(userClaims["user_id"].(string))
	if err != nil {
		return utils.GenericError(c, fiber.StatusUnauthorized, "Invalid user ID in token")
	}

	// REFACTOR: Panggil repository untuk mencari user
	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.GenericError(c, fiber.StatusNotFound, "User not found")
		}
		return utils.GenericError(c, fiber.StatusInternalServerError, "Database error")
	}

	return c.JSON(user)
}

// UpdateProfileRequest adalah struct untuk validasi update profil.
type UpdateProfileRequest struct {
	Name        string `json:"name" validate:"omitempty,min=3"`
	PhoneNumber string `json:"phone_number" validate:"omitempty,min=10"`
}

// UpdateMyProfile memperbarui data pengguna yang sedang login.
func (h *UserHandler) UpdateMyProfile(c *fiber.Ctx) error {
	req := new(UpdateProfileRequest)
	if err := c.BodyParser(req); err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid request body")
	}
	if errs := utils.ValidateStruct(req); errs != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"errors": errs})
	}

	userClaims := c.Locals("user").(jwt.MapClaims)
	userID, _ := uuid.Parse(userClaims["user_id"].(string))

	// REFACTOR: Ambil data user dari repository
	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		return utils.GenericError(c, fiber.StatusNotFound, "User not found")
	}

	// Update field hanya jika diisi
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.PhoneNumber != "" {
		user.PhoneNumber = req.PhoneNumber
	}

	// REFACTOR: Simpan perubahan melalui repository
	updatedUser, err := h.userRepo.Update(&user)
	if err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Failed to update profile")
	}

	return c.JSON(updatedUser)
}

// UploadMyAvatar menangani upload file avatar untuk pengguna.
func (h *UserHandler) UploadMyAvatar(c *fiber.Ctx) error {
	userClaims := c.Locals("user").(jwt.MapClaims)
	userID, _ := uuid.Parse(userClaims["user_id"].(string))

	file, err := c.FormFile("avatar")
	if err != nil {
		return utils.GenericError(c, fiber.StatusBadRequest, "Avatar file is required")
	}

	ext := filepath.Ext(file.Filename)
	allowedExt := map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
	if !allowedExt[strings.ToLower(ext)] {
		return utils.GenericError(c, fiber.StatusBadRequest, "Invalid file type. Only jpg, jpeg, png are allowed.")
	}

	newFileName := fmt.Sprintf("%s%s", userID.String(), ext)
	savePath := filepath.Join("./public/uploads/avatars", newFileName)

	// REFACTOR: Ambil data user dari repository
	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		return utils.GenericError(c, fiber.StatusNotFound, "User not found")
	}

	// Hapus avatar lama jika ada
	if user.Avatar != "" && !strings.HasPrefix(user.Avatar, "http") {
		oldPath := filepath.Join("./public", user.Avatar)
		os.Remove(oldPath)
	}

	if err := c.SaveFile(file, savePath); err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Failed to save file")
	}

	// Update path avatar di objek user
	avatarURL := "/uploads/avatars/" + newFileName
	user.Avatar = avatarURL

	// REFACTOR: Simpan perubahan melalui repository
	if _, err := h.userRepo.Update(&user); err != nil {
		return utils.GenericError(c, fiber.StatusInternalServerError, "Failed to update avatar path in database")
	}

	return c.JSON(fiber.Map{
		"message":    "Avatar updated successfully",
		"avatar_url": avatarURL,
	})
}
