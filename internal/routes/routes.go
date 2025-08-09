package routes

import (
	"ngabaca/internal/handler" // Kita akan buat handler nanti
	"ngabaca/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// SetupRoutes mendefinisikan semua rute untuk aplikasi.
func SetupRoutes(app *fiber.App) {
	// Middleware umum
	app.Use(logger.New()) // Logger untuk setiap request

	// Grup rute API utama
	api := app.Group("/api/v2")

	// Rute Publik (tidak perlu login)
	api.Get("/", handler.Home) // Sama seperti Route::get('/')
	api.Get("/catalog", handler.GetBooks)
	api.Get("/book/:slug", handler.GetBookDetail)
	api.Get("/categories", handler.GetCategories)

	// Rute Otentikasi
	auth := api.Group("/auth")
	auth.Post("/login", handler.Login)
	auth.Post("/register", handler.Register)
	// Rute Google OAuth
	auth.Get("/google", handler.GoogleLogin)
	auth.Get("/google/callback", handler.GoogleCallback)

	// Rute untuk Pelanggan (memerlukan login)
	customer := api.Group("/customer", middleware.Protected(), middleware.CheckRole("pelanggan", "admin"))
	customer.Get("/orders", handler.GetCustomerOrders)
	customer.Get("/orders/:id", handler.GetCustomerOrderDetail)
	customer.Post("/checkout", handler.Checkout)
	// ... rute pelanggan lainnya ...

	// Rute untuk Admin (memerlukan login dengan role admin)
	admin := api.Group("/admin", middleware.Protected(), middleware.CheckRole("admin"))

	// CRUD Books
	admin.Get("/books", handler.AdminGetBooks)
	admin.Post("/books", handler.AdminCreateBook)
	admin.Get("/books/:id", handler.AdminGetBook)
	admin.Put("/books/:id", handler.AdminUpdateBook)
	admin.Delete("/books/:id", handler.AdminDeleteBook)

	// ... rute admin lainnya untuk user, order, dll ...

	// Rute untuk webhook
	// app.Post("/midtrans/notification", handler.MidtransNotification)

}
