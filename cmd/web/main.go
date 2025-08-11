package main

import (
	"log"
	"ngabaca/internal/routes" // <-- Import routes
	"ngabaca/internal/server" // <-- Import server

	"github.com/robfig/cron/v3"
)

func main() {
	// 1. Buat server yang sudah terkonfigurasi lengkap dari paket 'server'
	server := server.NewServer()

	// 2. Jalankan scheduler (jika ada)
	c := cron.New()
	// c.AddFunc(...)
	go c.Start()
	defer c.Stop()

	// 3. Daftarkan semua rute dari paket 'routes'
	routes.Setup(server)

	// 4. Jalankan server Fiber
	log.Printf("Server berjalan di %s", server.Cfg.AppURL)
	log.Fatal(server.App.Listen(":3000"))
}