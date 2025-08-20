package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"ngabaca/internal/model"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	FindAll() ([]CategoryResponse, error)
	FindByID(id string) (CategoryResponse, error)
}

type categoryRepository struct {
	db  *gorm.DB
	rdb *redis.Client
}

type CategoryResponse struct {
	ID    string        `json:"id"`
	Name  string        `json:"name"`
	Slug  string        `json:"slug"`
	Books []BookSummary `json:"books,omitempty"`
}

type BookSummary struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Author     string `json:"author"`
	CategoryID string `json:"category_id"`
}

func NewCategoryRepository(db *gorm.DB, rdb *redis.Client) CategoryRepository {
	return &categoryRepository{db: db, rdb: rdb}
}
func (r *categoryRepository) FindAll() ([]CategoryResponse, error) {
	ctx := context.Background()
	cacheKey := "categories"
	var responses []CategoryResponse
	cachedCategories, err := r.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		fmt.Println("CACHE HIT: Mengambil kategori dari Redis.")
		err = json.Unmarshal([]byte(cachedCategories), &responses)
		return responses, err
	}

	if err != redis.Nil {
		return nil, err
	}

	fmt.Println("CACHE MISS: Mengambil kategori dari Database.")
	var categories []model.Category
	dbErr := r.db.Select("id", "name", "slug").Find(&categories).Error
	if dbErr != nil {
		return nil, dbErr
	}

	responses = make([]CategoryResponse, len(categories))
	for i, cat := range categories {
		responses[i] = CategoryResponse{
			ID:   cat.ID.String(),
			Name: cat.Name,
			Slug: cat.Slug,
		}
	}
	data, err := json.Marshal(responses)
	if err != nil {
		return nil, err
	}
	err = r.rdb.Set(ctx, cacheKey, data, 24*time.Hour).Err()
	if err != nil {
		fmt.Println("Gagal menyimpan kategori ke cache:", err)
	}

	return responses, nil
}

// find category by id
func (r *categoryRepository) FindByID(id string) (CategoryResponse, error) {
	ctx := context.Background()
	cacheKey := "category:by_key:" + id

	var categoryResp CategoryResponse

	// Coba ambil dari Redis (key khusus per ID)
	cached, err := r.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		fmt.Println("CACHE HIT: Mengambil kategori dari Redis (by ID).")
		if uErr := json.Unmarshal([]byte(cached), &categoryResp); uErr != nil {
			return CategoryResponse{}, uErr
		}
		return categoryResp, nil
	}
	if err != redis.Nil {
		return CategoryResponse{}, err
	}

	// Coba ambil dari cache "categories"
	categoriesCache, err := r.rdb.Get(ctx, "categories").Result()
	if err == nil {
		var categories []CategoryResponse
		if uErr := json.Unmarshal([]byte(categoriesCache), &categories); uErr == nil {
			for _, cat := range categories {
				if cat.ID == id {
					fmt.Println("CACHE HIT: Mengambil kategori dari Redis (categories).")

					// Simpan juga ke cache per ID biar lebih cepat diakses nanti
					data, _ := json.Marshal(cat)
					_ = r.rdb.Set(ctx, cacheKey, data, 24*time.Hour).Err()

					return cat, nil
				}
			}
		}
	}

	// Fallback terakhir: ambil dari DB
	fmt.Println("CACHE MISS: Mengambil kategori dari Database.")
	var category model.Category
	if dbErr := r.db.Where("id = ?", id).First(&category).Error; dbErr != nil {
		return CategoryResponse{}, dbErr
	}

	categoryResp = CategoryResponse{
		ID:   category.ID.String(),
		Name: category.Name,
		Slug: category.Slug,
	}

	// Simpan ke cache
	data, _ := json.Marshal(categoryResp)
	_ = r.rdb.Set(ctx, cacheKey, data, 24*time.Hour).Err()

	return categoryResp, nil
}
