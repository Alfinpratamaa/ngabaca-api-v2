package routes

import (
	"ngabaca/internal/middleware"
	"ngabaca/internal/server"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// SetupRoutes mendefinisikan semua rute untuk aplikasi.
func Setup(s *server.Server) {
	// Middleware umum
	s.App.Use(logger.New())
	s.App.Static("/", "./public")

	api := s.App.Group("/api/v2")
	api.Static("/docs", "./api-docs")
	api.Static("/css", "./api-docs/css")
	api.Static("/js", "./api-docs/js")
	api.Static("/content", "./api-docs/content")
	api.Static("/favicon.ico", "./api-docs/favicon.ico")

	// Redirect dari /api/v2 ke /api/v2/docs
	api.Get("/", func(c *fiber.Ctx) error {
		// Pastikan redirect ke path yang benar termasuk /docs
		return c.Redirect("/api/v2/docs")
	})
	// --- Rute Publik ---
	api.Get("/catalog", s.PublicHandler.GetBooks)
	api.Get("/book/:slug", s.PublicHandler.GetBookDetail)
	api.Get("/categories", s.PublicHandler.GetCategories)
	api.Get("/categories/:id", s.PublicHandler.GetCategoryByID)
	api.Get("/search", s.PublicHandler.SearchBooks)
	// Rute publik untuk melihat ulasan dipindahkan ke CustomerHandler
	api.Get("/books/:id/reviews", s.CustomerHandler.GetBookReviews)

	// --- Rute Otentikasi ---
	auth := api.Group("/auth")
	auth.Post("/login", s.AuthHandler.Login)
	auth.Post("/register", s.AuthHandler.Register)
	auth.Get("/google", s.AuthHandler.GoogleLogin)
	auth.Get("/google/callback", s.AuthHandler.GoogleCallback)

	// --- Rute Profil Pengguna (terproteksi) ---
	me := api.Group("/me", middleware.Protected())
	me.Get("/", s.UserHandler.GetMyProfile)
	me.Put("/", s.UserHandler.UpdateMyProfile)
	me.Post("/avatar", s.UserHandler.UploadMyAvatar)

	wishlist := me.Group("/wishlist")
	wishlist.Get("/", s.CustomerHandler.GetMyWishlist)
	wishlist.Post("/", s.CustomerHandler.AddToWishlist)
	wishlist.Delete("/:bookId", s.CustomerHandler.RemoveFromWishlist)

	// --- Rute Pelanggan (terproteksi) ---
	customer := api.Group("/customer", middleware.Protected())
	customer.Get("/orders", s.CustomerHandler.GetCustomerOrders)
	customer.Get("/orders/:id", s.CustomerHandler.GetCustomerOrderDetail)
	customer.Post("/checkout", s.CustomerHandler.Checkout)
	// PINDAHKAN RUTE CREATE REVIEW KE SINI
	customer.Post("/books/:id/reviews", s.CustomerHandler.CreateReview)

	// Rute untuk Admin (memerlukan login dengan role admin)
	admin := api.Group("/admin", middleware.Protected(), middleware.CheckRole("admin"))
	admin.Get("/books", s.AdminHandler.AdminGetBooks)
	admin.Post("/books", s.AdminHandler.AdminCreateBook)
	admin.Get("/books/:id", s.AdminHandler.AdminGetBook)
	admin.Put("/books/:id", s.AdminHandler.AdminUpdateBook)
	admin.Delete("/books/:id", s.AdminHandler.AdminDeleteBook)

	// --- Manajemen Pengguna ---
	admin.Get("/users", s.AdminHandler.AdminGetUsers)
	admin.Put("/users/:id", s.AdminHandler.AdminUpdateUser)

	// --- Manajemen Pesanan ---
	admin.Get("/orders", s.AdminHandler.AdminGetOrders)
	admin.Get("/orders/:id", s.AdminHandler.AdminGetOrderDetail)
	admin.Put("/orders/:id/status", s.AdminHandler.AdminUpdateOrderStatus)

	// Rute untuk webhook
	s.App.Post("/midtrans/notification", s.PaymentHandler.MidtransNotification)

}
