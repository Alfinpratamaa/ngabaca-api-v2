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
