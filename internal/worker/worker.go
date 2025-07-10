// internal/worker/worker.go
package worker

import (
	"consultation-booking/internal/models"
	"consultation-booking/internal/services"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Worker struct {
	db                  *gorm.DB
	redis               *redis.Client
	emailService        *services.EmailService
	notificationService *services.NotificationService
}

func NewWorker(db *gorm.DB, redis *redis.Client, emailService *services.EmailService, notificationService *services.NotificationService) *Worker {
	return &Worker{
		db:                  db,
		redis:               redis,
		emailService:        emailService,
		notificationService: notificationService,
	}
}

func (w *Worker) Start() {
	// Run reminder job every 10 minutes
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	log.Println("Worker started")

	for {
		select {
		case <-ticker.C:
			w.processReminders()
			w.processExpiredBookings()
			w.cleanupOldNotifications()
		}
	}
}

func (w *Worker) processReminders() {
	// Get bookings that need reminders (60 minutes before)
	reminderTime := time.Now().Add(60 * time.Minute)
	var bookings []models.Booking

	w.db.Where("start_time BETWEEN ? AND ? AND status = ?",
		reminderTime, reminderTime.Add(10*time.Minute), "confirmed").
		Preload("User").
		Preload("Expert").
		Preload("Expert.User").
		Find(&bookings)

	for _, booking := range bookings {
		// Send email reminder
		err := w.emailService.SendReminder(
			booking.User.Email,
			booking.Expert.User.Name,
			booking.StartTime.Format("2006-01-02 15:04"),
		)
		if err != nil {
			log.Printf("Failed to send reminder email: %v", err)
		}

		// Create notification
		w.notificationService.CreateNotification(
			booking.UserID,
			"Consultation Reminder",
			fmt.Sprintf("Your consultation with %s is starting in 1 hour", booking.Expert.User.Name),
			"reminder",
		)

		// Send notification to expert
		w.notificationService.CreateNotification(
			booking.Expert.UserID,
			"Consultation Reminder",
			fmt.Sprintf("Your consultation with %s is starting in 1 hour", booking.User.Name),
			"reminder",
		)
	}

	log.Printf("Processed %d reminders", len(bookings))
}

func (w *Worker) processExpiredBookings() {
	// Mark bookings as missed if they're past their time and still pending
	var expiredBookings []models.Booking
	w.db.Where("end_time < ? AND status = ?", time.Now(), "pending").Find(&expiredBookings)

	for _, booking := range expiredBookings {
		w.db.Model(&booking).Update("status", "missed")

		// Free up the slot
		w.db.Model(&models.AvailableSlot{}).Where(
			"expert_id = ? AND start_time <= ? AND end_time >= ?",
			booking.ExpertID, booking.StartTime, booking.EndTime,
		).Update("is_booked", false)
	}

	log.Printf("Processed %d expired bookings", len(expiredBookings))
}

func (w *Worker) cleanupOldNotifications() {
	// Delete notifications older than 30 days
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	result := w.db.Where("created_at < ?", thirtyDaysAgo).Delete(&models.Notification{})

	if result.Error != nil {
		log.Printf("Failed to cleanup old notifications: %v", result.Error)
	} else {
		log.Printf("Cleaned up %d old notifications", result.RowsAffected)
	}
}
