package database

import (
	"context"
	"fmt"
	"log"
	"ngabaca/config"
	"ngabaca/internal/model"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

var RDB *redis.Client

func ConnectRedis(cfg config.Config) {
	RDB = redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Tes koneksi
	_, err := RDB.Ping(context.Background()).Result()
	if err != nil {
		panic(fmt.Sprintf("Gagal terhubung ke Redis: %v", err))
	}

	fmt.Println("Koneksi Redis berhasil.")
}

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
		&model.Wishlist{},
	)
	if err != nil {
		log.Fatal("Gagal melakukan migrasi:", err)
	}
	fmt.Println("Migrasi database selesai.")
	return DB
}
