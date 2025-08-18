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
		{"Bisnis", "bisnis"},
		{"Ekonomi", "ekonomi"},
		{"Politik", "politik"},
		{"Seni", "seni"},
		{"Musik", "musik"},
		{"Film", "film"},
		{"Literasi", "literasi"},
		{"Fiksi", "fiksi"},
		{"Non Fiksi", "non-fiksi"},
		{"Sejarah", "sejarah"},
		{"Budaya", "budaya"},
		{"Agama", "agama"},
		{"Filsafat", "filsafat"},
		{"Psikologi", "psikologi"},
		{"Sosial", "sosial"},
		{"Hukum", "hukum"},
		{"Kriminal", "kriminal"},
		{"Lingkungan", "lingkungan"},
		{"Sains", "sains"},
		{"Matematika", "matematika"},
		{"Fisika", "fisika"},
		{"Kimia", "kimia"},
		{"Biologi", "biologi"},
		{"Astronomi", "astronomi"},
		{"Geografi", "geografi"},
		{"Pertanian", "pertanian"},
		{"Peternakan", "peternakan"},
		{"Perikanan", "perikanan"},
		{"Teknik", "teknik"},
		{"Arsitektur", "arsitektur"},
		{"Desain", "desain"},
		{"Fotografi", "fotografi"},
		{"Jurnalistik", "jurnalistik"},
		{"Komunikasi", "komunikasi"},
		{"Transportasi", "transportasi"},
		{"Pariwisata", "pariwisata"},
		{"Kuliner", "kuliner"},
		{"Resep", "resep"},
		{"Travel", "travel"},
		{"Gaya Hidup", "gaya-hidup"},
		{"Fashion", "fashion"},
		{"Kecantikan", "kecantikan"},
		{"Keluarga", "keluarga"},
		{"Pernikahan", "pernikahan"},
		{"Parenting", "parenting"},
		{"Anak", "anak"},
		{"Remaja", "remaja"},
		{"Dewasa", "dewasa"},
		{"Lansia", "lansia"},
		{"Komputer", "komputer"},
		{"Internet", "internet"},
		{"AI", "ai"},
		{"Blockchain", "blockchain"},
		{"Kripto", "kripto"},
		{"Startup", "startup"},
		{"Manajemen", "manajemen"},
		{"Marketing", "marketing"},
		{"Investasi", "investasi"},
		{"Saham", "saham"},
		{"Properti", "properti"},
		{"Perbankan", "perbankan"},
		{"Asuransi", "asuransi"},
		{"Pajak", "pajak"},
		{"Kerja", "kerja"},
		{"Karier", "karier"},
		{"Freelance", "freelance"},
		{"Motivasi", "motivasi"},
		{"Pengembangan Diri", "pengembangan-diri"},
		{"Produktivitas", "produktivitas"},
		{"Keterampilan", "keterampilan"},
		{"Bahasa", "bahasa"},
		{"Inggris", "inggris"},
		{"Jepang", "jepang"},
		{"Korea", "korea"},
		{"Mandarin", "mandarin"},
		{"Selebriti", "selebriti"},
		{"Game", "game"},
		{"E-Sport", "e-sport"},
		{"Anime", "anime"},
		{"Manga", "manga"},
		{"Komik", "komik"},
		{"Novel", "novel"},
		{"Puisi", "puisi"},
		{"Cerpen", "cerpen"},
		{"Opini", "opini"},
		{"Review", "review"},
		{"Tutorial", "tutorial"},
		{"Tips", "tips"},
		{"Inspirasi", "inspirasi"},
		{"Berita", "berita"},
		{"Trend", "trend"},
		{"Viral", "viral"},
		{"Random", "random"},
		{"Umum", "umum"},
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
