// internal/routes/routes.go
package routes

import (
	"consultation-booking/internal/handlers"
	"consultation-booking/internal/middleware"
	"consultation-booking/internal/services"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	router *gin.Engine,
	userService *services.UserService,
	expertService *services.ExpertService,
	bookingService *services.BookingService,
	notificationService *services.NotificationService,
) {
	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	expertHandler := handlers.NewExpertHandler(expertService)
	bookingHandler := handlers.NewBookingHandler(bookingService)
	notificationHandler := handlers.NewNotificationHandler(notificationService)

	// Public routes
	api := router.Group("/api/v1")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
			auth.POST("/refresh", userHandler.RefreshToken)
		}

		// Public expert routes
		experts := api.Group("/experts")
		{
			experts.GET("", expertHandler.GetExperts)
			experts.GET("/:id", expertHandler.GetExpertByID)
			experts.GET("/:id/slots", expertHandler.GetAvailableSlots)
		}
	}

	// Protected routes
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware("your-secret-key"))
	{
		// User routes
		user := protected.Group("/user")
		{
			user.GET("/profile", userHandler.GetProfile)
			user.PUT("/profile", userHandler.UpdateProfile)
			user.GET("/bookings", userHandler.GetBookingHistory)
		}

		// Expert routes
		expert := protected.Group("/expert")
		expert.Use(middleware.RoleMiddleware("expert", "admin"))
		{
			expert.POST("/slots", expertHandler.CreateAvailableSlot)
			expert.GET("/bookings", expertHandler.GetExpertBookings)
			expert.PUT("/bookings/:id/status", expertHandler.UpdateBookingStatus)
		}

		// Booking routes
		booking := protected.Group("/bookings")
		{
			booking.POST("", bookingHandler.CreateBooking)
			booking.GET("/:id", bookingHandler.GetBooking)
			booking.PUT("/:id/cancel", bookingHandler.CancelBooking)
		}

		// Notification routes
		notification := protected.Group("/notifications")
		{
			notification.GET("", notificationHandler.GetNotifications)
			notification.PUT("/:id/read", notificationHandler.MarkAsRead)
			notification.GET("/unread-count", notificationHandler.GetUnreadCount)
		}

		// Admin routes
		admin := protected.Group("/admin")
		admin.Use(middleware.RoleMiddleware("admin"))
		{
			admin.POST("/experts", expertHandler.CreateExpert)
			admin.GET("/bookings", bookingHandler.GetAllBookings)
			admin.GET("/stats", bookingHandler.GetStats)
		}
	}
}
