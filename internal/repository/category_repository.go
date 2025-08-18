package repository

import (
	"ngabaca/internal/model"

	"gorm.io/gorm"
)

type CategoryRepository interface {
	FindAll() ([]CategoryResponse, error)
}

type categoryRepository struct {
	db *gorm.DB
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

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}
func (r *categoryRepository) FindAll() ([]CategoryResponse, error) {
	var categories []model.Category
	err := r.db.Select("id", "name", "slug").Find(&categories).Error
	if err != nil {
		return nil, err
	}

	responses := make([]CategoryResponse, len(categories))
	for i, cat := range categories {
		responses[i] = CategoryResponse{
			ID:   cat.ID.String(),
			Name: cat.Name,
			Slug: cat.Slug,
		}
	}

	return responses, nil
}
