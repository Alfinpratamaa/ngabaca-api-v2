package server

import (
	"log"
	"ngabaca/config"
	"ngabaca/database"
	"ngabaca/internal/handler"
	"ngabaca/internal/repository"
	"ngabaca/internal/service"
	"ngabaca/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"gorm.io/gorm"
)

// Server adalah struct utama yang menampung semua dependency aplikasi.
type Server struct {
	App             *fiber.App
	DB              *gorm.DB
	Cfg             config.Config
	AdminHandler    *handler.AdminHandler
	AuthHandler     *handler.AuthHandler
	PublicHandler   *handler.PublicHandler
	CustomerHandler *handler.CustomerHandler
	PaymentHandler  *handler.PaymentHandler
	UserHandler     *handler.UserHandler
}

// NewServer adalah constructor yang merakit semua komponen aplikasi.
func NewServer() *Server {
	// Load config
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("Tidak dapat memuat konfigurasi:", err)
	}

	// Setup utilitas
	utils.SetupGoogleOAuthConfig(cfg)

	// Hubungkan DB
	db := database.ConnectDB(cfg)
	database.ConnectRedis(cfg)

	// seed database
	if err := database.SeedCategories(db); err != nil {
		log.Fatal("Gagal melakukan seeding kategori:", err)
	}

	// Inisialisasi semua repository
	bookRepo := repository.NewBookRepository(db)
	userRepo := repository.NewUserRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	categoryRepo := repository.NewCategoryRepository(db, database.RDB)
	reviewRepo := repository.NewReviewRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)

	orderService := service.NewOrderService(db, bookRepo, orderRepo, paymentRepo)
	paymentService := service.NewPaymentService(db, orderRepo, paymentRepo)
	// Inisialisasi semua handler
	adminHandler := handler.NewAdminHandler(bookRepo, userRepo, orderRepo, cfg)
	authHandler := handler.NewAuthHandler(userRepo, cfg)
	publicHandler := handler.NewPublicHandler(bookRepo, categoryRepo)
	customerHandler := handler.NewCustomerHandler(orderRepo, userRepo, orderService, reviewRepo, cfg)
	paymentHandler := handler.NewPaymentHandler(paymentService)
	userHandler := handler.NewUserHandler(userRepo, cfg)

	// Buat instance Fiber
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Kembalikan Server struct yang sudah lengkap
	return &Server{
		App:             app,
		DB:              db,
		Cfg:             cfg,
		AdminHandler:    adminHandler,
		AuthHandler:     authHandler,
		PublicHandler:   publicHandler,
		CustomerHandler: customerHandler,
		PaymentHandler:  paymentHandler,
		UserHandler:     userHandler,
	}
}
