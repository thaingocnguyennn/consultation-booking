// main.go
package main

import (
	"consultation-booking/internal/config"
	"consultation-booking/internal/database"
	"consultation-booking/internal/middleware"
	"consultation-booking/internal/models"
	"consultation-booking/internal/routes"
	"consultation-booking/internal/services"
	"consultation-booking/internal/worker"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate database schema
	if err := db.AutoMigrate(
		&models.User{},
		&models.Expert{},
		&models.Booking{},
		&models.Notification{},
		&models.Feedback{},
		&models.AvailableSlot{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize Redis
	redisClient := database.InitRedis(cfg.RedisURL)

	// Initialize services
	userService := services.NewUserService(db, redisClient)
	expertService := services.NewExpertService(db, redisClient)
	bookingService := services.NewBookingService(db, redisClient)
	notificationService := services.NewNotificationService(db, redisClient)
	emailService := services.NewEmailService(cfg.SMTPConfig)

	// Initialize worker
	workerService := worker.NewWorker(db, redisClient, emailService, notificationService)
	go workerService.Start()

	// Initialize Gin router
	router := gin.Default()

	// Add middleware
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RateLimitMiddleware(redisClient))
	router.Use(middleware.LoggingMiddleware())

	// Setup routes
	routes.SetupRoutes(router, userService, expertService, bookingService, notificationService)

	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, router))
}
