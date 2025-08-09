package main

import (
	"log"
	"ngabaca/config"
	"ngabaca/database"
	"ngabaca/internal/routes"
	"ngabaca/internal/scheduler"

	"github.com/gofiber/fiber/v2"
	"github.com/robfig/cron/v3"
)

func main() {
	// 1. Muat Konfigurasi
	cfg, err := config.LoadConfig(".") // "." berarti cari di folder root
	if err != nil {
		log.Fatal("Tidak dapat memuat konfigurasi:", err)
	}

	// 2. Hubungkan ke Database
	database.ConnectDB(cfg)

	c := cron.New()

	c.AddFunc("@every 5m", scheduler.CancelExpiredOrders)

	go c.Start()

	defer c.Stop()

	// 3. Inisialisasi Fiber App
	app := fiber.New()

	// 4. Setup Routes
	routes.SetupRoutes(app)

	// 5. Jalankan Server
	log.Printf("Server berjalan di %s", cfg.AppURL)
	err = app.Listen(":3000") // Port bisa diambil dari config
	if err != nil {
		log.Fatal("Gagal menjalankan server:", err)
	}
}
