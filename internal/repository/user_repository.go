package repository

import (
	"ngabaca/internal/model"

	"github.com/google/uuid"
	oauth2v2 "google.golang.org/api/oauth2/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GoogleClaims struct {
	Sub            string
	Email          string
	Name           string
	Picture        string
	EmailVerified  bool
	DefaultAppRole string
}

// UserRepository mendefinisikan kontrak untuk data pengguna.
type UserRepository interface {
	FindAll() ([]model.User, error)
	FindByID(id uuid.UUID) (model.User, error)
	Update(user *model.User) (*model.User, error)
	// Kita tambahkan ini untuk refactor handler auth nanti
	FindByEmail(email string) (model.User, error)
	Create(user *model.User) (*model.User, error)
	FindOrCreateByGoogle(googleUser *oauth2v2.Userinfo) (model.User, error)
	FindOrCreateFromGoogleClaims(claims GoogleClaims) (model.User, error)
}

type userRepository struct {
	db *gorm.DB
}

// FindOrCreateFromGoogleClaims implements UserRepository.
func (r *userRepository) FindOrCreateFromGoogleClaims(claims GoogleClaims) (model.User, error) {
	var user model.User

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Coba cari user berdasarkan Google ID
		if err := tx.Where("google_id = ?", claims.Sub).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Jika tidak ada, coba cari berdasarkan email
				if err := tx.Where("email = ?", claims.Email).First(&user).Error; err == nil {
					// Email ada, user pernah daftar manual. Update Google ID-nya.
					user.GoogleID = &claims.Sub
					user.Avatar = claims.Picture // Update avatar juga
					return tx.Save(&user).Error
				}
				// Jika email juga tidak ada, buat user baru
				newUser := model.User{
					Name:     claims.Name,
					Email:    claims.Email,
					GoogleID: &claims.Sub,
					Avatar:   claims.Picture,
					Role:     "pelanggan",
				}
				if err := tx.Create(&newUser).Error; err != nil {
					return err
				}
				user = newUser
			} else {
				// Error lain saat query
				return err
			}
		}
		return nil
	})

	return user, err
}

// NewUserRepository adalah constructor untuk userRepository.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindAll() ([]model.User, error) {
	var users []model.User
	err := r.db.Find(&users).Error
	return users, err
}

func (r *userRepository) FindByID(id uuid.UUID) (model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	return user, err
}

func (r *userRepository) Update(user *model.User) (*model.User, error) {
	err := r.db.Save(user).Error
	return user, err
}

func (r *userRepository) FindByEmail(email string) (model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return user, err
}
func (r *userRepository) Create(user *model.User) (*model.User, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Returning{}).Create(user).Error; err != nil {
			return err
		}
		cart := model.Cart{UserID: user.ID}
		if err := tx.Create(&cart).Error; err != nil {
			return err
		}
		return nil
	})
	return user, err
}

func (r *userRepository) FindOrCreateByGoogle(userInfo *oauth2v2.Userinfo) (model.User, error) {
	var user model.User

	// Gunakan transaksi untuk memastikan konsistensi data
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Coba cari user berdasarkan Google ID
		if err := tx.Where("google_id = ?", userInfo.Id).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Jika tidak ada, coba cari berdasarkan email
				if err := tx.Where("email = ?", userInfo.Email).First(&user).Error; err == nil {
					// Email ada, user pernah daftar manual. Update Google ID-nya.
					user.GoogleID = &userInfo.Id
					user.Avatar = userInfo.Picture // Update avatar juga
					return tx.Save(&user).Error
				}

				// Jika email juga tidak ada, buat user baru
				newUser := model.User{
					Name:     userInfo.Name,
					Email:    userInfo.Email,
					GoogleID: &userInfo.Id,
					Avatar:   userInfo.Picture,
					Role:     "pelanggan",
				}
				if err := tx.Create(&newUser).Error; err != nil {
					return err
				}
				user = newUser
			} else {
				// Error lain saat query
				return err
			}
		}
		return nil
	})

	return user, err
}
