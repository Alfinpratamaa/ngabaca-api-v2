package database

import (
	"fmt"
	"log"
	"ngabaca/config"
	"ngabaca/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// ConnectDB menginisialisasi koneksi ke database dan melakukan auto-migrate.
func ConnectDB(cfg config.Config) *gorm.DB {
	var err error
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost, cfg.DBUsername, cfg.DBPassword, cfg.DBDatabase, cfg.DBPort, cfg.DBSSLMode)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal terhubung ke database:", err)
	}

	fmt.Println("Koneksi database berhasil.")

	// AutoMigrate akan membuat tabel berdasarkan struct model
	fmt.Println("Menjalankan migrasi database...")
	err = DB.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.Book{},
		&model.Order{},
		&model.OrderItem{},
		&model.Payment{},
		&model.Review{},
	)
	if err != nil {
		log.Fatal("Gagal melakukan migrasi:", err)
	}
	fmt.Println("Migrasi database selesai.")
	return DB
}
