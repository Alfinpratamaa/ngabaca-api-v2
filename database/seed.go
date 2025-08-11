package database

import (
	"context"
	"fmt"
	"ngabaca/internal/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedCategories inserts default categories if they do not yet exist.
// Call this after AutoMigrate.
func SeedCategories(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}

	// Ensure table exists
	if err := db.AutoMigrate(&model.Category{}); err != nil {
		return fmt.Errorf("automigrate Category: %w", err)
	}

	ctx := context.Background()

	// Define the categories you want to seed.
	// Add / modify as needed.
	defaults := []struct {
		Name string
		Slug string
	}{
		{"Teknologi", "teknologi"},
		{"Pendidikan", "pendidikan"},
		{"Kesehatan", "kesehatan"},
		{"Olahraga", "olahraga"},
		{"Hiburan", "hiburan"},
	}

	for _, c := range defaults {
		var existing model.Category
		err := db.WithContext(ctx).
			Where("slug = ?", c.Slug).
			First(&existing).Error

		if err == nil {
			// Already exists, skip
			continue
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return fmt.Errorf("query category %s: %w", c.Slug, err)
		}

		newCat := model.Category{
			Basemodel: model.Basemodel{
				ID:        uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name: c.Name,
			Slug: c.Slug,
		}

		if err := db.WithContext(ctx).Create(&newCat).Error; err != nil {
			return fmt.Errorf("create category %s: %w", c.Slug, err)
		}
	}

	return nil
}

// Optionally a helper that runs all seeds (expand later).
func SeedAll(db *gorm.DB) error {
	if err := SeedCategories(db); err != nil {
		return err
	}
	return nil
}
