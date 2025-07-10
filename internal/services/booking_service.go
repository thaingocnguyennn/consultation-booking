// internal/services/booking_service.go
package services

import (
	"consultation-booking/internal/models"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type BookingService struct {
	db    *gorm.DB
	redis *redis.Client
}

type CreateBookingRequest struct {
	ExpertID  uint      `json:"expert_id" binding:"required"`
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" binding:"required"`
	Notes     string    `json:"notes"`
	Format    string    `json:"format"` // online, offline
}

func NewBookingService(db *gorm.DB, redis *redis.Client) *BookingService {
	return &BookingService{
		db:    db,
		redis: redis,
	}
}

func (s *BookingService) CreateBooking(userID uint, req CreateBookingRequest) (*models.Booking, error) {
	// Check if expert exists and is available
	var expert models.Expert
	if err := s.db.First(&expert, req.ExpertID).Error; err != nil {
		return nil, errors.New("expert not found")
	}

	if !expert.IsAvailable {
		return nil, errors.New("expert is not available")
	}

	// Check for conflicts - user shouldn't have overlapping bookings
	var userConflictCount int64
	s.db.Model(&models.Booking{}).Where(
		"user_id = ? AND status NOT IN (?, ?) AND ((start_time <= ? AND end_time > ?) OR (start_time < ? AND end_time >= ?))",
		userID, "cancelled", "completed", req.StartTime, req.StartTime, req.EndTime, req.EndTime,
	).Count(&userConflictCount)

	if userConflictCount > 0 {
		return nil, errors.New("you have a conflicting booking at this time")
	}

	// Check for expert conflicts
	var expertConflictCount int64
	s.db.Model(&models.Booking{}).Where(
		"expert_id = ? AND status NOT IN (?, ?) AND ((start_time <= ? AND end_time > ?) OR (start_time < ? AND end_time >= ?))",
		req.ExpertID, "cancelled", "completed", req.StartTime, req.StartTime, req.EndTime, req.EndTime,
	).Count(&expertConflictCount)

	if expertConflictCount > 0 {
		return nil, errors.New("expert has a conflicting booking at this time")
	}

	// Check if the time slot is available
	var availableSlot models.AvailableSlot
	if err := s.db.Where(
		"expert_id = ? AND start_time <= ? AND end_time >= ? AND is_booked = ?",
		req.ExpertID, req.StartTime, req.EndTime, false,
	).First(&availableSlot).Error; err != nil {
		return nil, errors.New("time slot is not available")
	}

	// Create booking
	booking := models.Booking{
		UserID:    userID,
		ExpertID:  req.ExpertID,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Status:    "pending",
		Notes:     req.Notes,
		Format:    req.Format,
	}

	if err := s.db.Create(&booking).Error; err != nil {
		return nil, err
	}

	// Mark slot as booked
	s.db.Model(&models.AvailableSlot{}).Where("id = ?", availableSlot.ID).Update("is_booked", true)

	// Load relationships
	s.db.Preload("User").Preload("Expert").Preload("Expert.User").First(&booking, booking.ID)

	return &booking, nil
}

func (s *BookingService) GetBooking(bookingID uint) (*models.Booking, error) {
	var booking models.Booking
	err := s.db.Preload("User").Preload("Expert").Preload("Expert.User").First(&booking, bookingID).Error
	return &booking, err
}

func (s *BookingService) CancelBooking(bookingID uint, userID uint, reason string) error {
	var booking models.Booking
	if err := s.db.First(&booking, bookingID).Error; err != nil {
		return errors.New("booking not found")
	}

	// Check if user owns the booking or is the expert
	if booking.UserID != userID && booking.Expert.UserID != userID {
		return errors.New("unauthorized to cancel this booking")
	}

	// Check if cancellation is allowed (at least 1 hour before)
	if time.Now().Add(time.Hour).After(booking.StartTime) {
		return errors.New("cannot cancel booking less than 1 hour before start time")
	}

	// Update booking status
	updates := map[string]interface{}{
		"status":        "cancelled",
		"cancel_reason": reason,
	}

	if err := s.db.Model(&booking).Updates(updates).Error; err != nil {
		return err
	}

	// Free up the slot
	s.db.Model(&models.AvailableSlot{}).Where(
		"expert_id = ? AND start_time <= ? AND end_time >= ?",
		booking.ExpertID, booking.StartTime, booking.EndTime,
	).Update("is_booked", false)

	return nil
}

func (s *BookingService) UpdateBookingStatus(bookingID uint, status string) error {
	return s.db.Model(&models.Booking{}).Where("id = ?", bookingID).Update("status", status).Error
}

func (s *BookingService) GetUpcomingBookings() ([]models.Booking, error) {
	var bookings []models.Booking
	err := s.db.Where("start_time > ? AND start_time < ? AND status IN (?)", 
		time.Now(), time.Now().Add(time.Hour*2), []string{"pending", "confirmed"}).
		Preload("User").
		Preload("Expert").
		Preload("Expert.User").
		Find(&bookings).Error
	return bookings, err
}// internal/services/booking_service.go
package services

import (
	"consultation-booking/internal/models"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type BookingService struct {
	db    *gorm.DB
	redis *redis.Client
}

type CreateBookingRequest struct {
	ExpertID  uint      `json:"expert_id" binding:"required"`
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" binding:"required"`
	Notes     string    `json:"notes"`
	Format    string    `json:"format"` // online, offline
}

func NewBookingService(db *gorm.DB, redis *redis.Client) *BookingService {
	return &BookingService{
		db:    db,
		redis: redis,
	}
}

func (s *BookingService) CreateBooking(userID uint, req CreateBookingRequest) (*models.Booking, error) {
	// Check if expert exists and is available
	var expert models.Expert
	if err := s.db.First(&expert, req.ExpertID).Error; err != nil {
		return nil, errors.New("expert not found")
	}

	if !expert.IsAvailable {
		return nil, errors.New("expert is not available")
	}

	// Check for conflicts - user shouldn't have overlapping bookings
	var userConflictCount int64
	s.db.Model(&models.Booking{}).Where(
		"user_id = ? AND status NOT IN (?, ?) AND ((start_time <= ? AND end_time > ?) OR (start_time < ? AND end_time >= ?))",
		userID, "cancelled", "completed", req.StartTime, req.StartTime, req.EndTime, req.EndTime,
	).Count(&userConflictCount)

	if userConflictCount > 0 {
		return nil, errors.New("you have a conflicting booking at this time")
	}

	// Check for expert conflicts
	var expertConflictCount int64
	s.db.Model(&models.Booking{}).Where(
		"expert_id = ? AND status NOT IN (?, ?) AND ((start_time <= ? AND end_time > ?) OR (start_time < ? AND end_time >= ?))",
		req.ExpertID, "cancelled", "completed", req.StartTime, req.StartTime, req.EndTime, req.EndTime,
	).Count(&expertConflictCount)

	if expertConflictCount > 0 {
		return nil, errors.New("expert has a conflicting booking at this time")
	}

	// Check if the time slot is available
	var availableSlot models.AvailableSlot
	if err := s.db.Where(
		"expert_id = ? AND start_time <= ? AND end_time >= ? AND is_booked = ?",
		req.ExpertID, req.StartTime, req.EndTime, false,
	).First(&availableSlot).Error; err != nil {
		return nil, errors.New("time slot is not available")
	}

	// Create booking
	booking := models.Booking{
		UserID:    userID,
		ExpertID:  req.ExpertID,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Status:    "pending",
		Notes:     req.Notes,
		Format:    req.Format,
	}

	if err := s.db.Create(&booking).Error; err != nil {
		return nil, err
	}

	// Mark slot as booked
	s.db.Model(&models.AvailableSlot{}).Where("id = ?", availableSlot.ID).Update("is_booked", true)

	// Load relationships
	s.db.Preload("User").Preload("Expert").Preload("Expert.User").First(&booking, booking.ID)

	return &booking, nil
}

func (s *BookingService) GetBooking(bookingID uint) (*models.Booking, error) {
	var booking models.Booking
	err := s.db.Preload("User").Preload("Expert").Preload("Expert.User").First(&booking, bookingID).Error
	return &booking, err
}

func (s *BookingService) CancelBooking(bookingID uint, userID uint, reason string) error {
	var booking models.Booking
	if err := s.db.First(&booking, bookingID).Error; err != nil {
		return errors.New("booking not found")
	}

	// Check if user owns the booking or is the expert
	if booking.UserID != userID && booking.Expert.UserID != userID {
		return errors.New("unauthorized to cancel this booking")
	}

	// Check if cancellation is allowed (at least 1 hour before)
	if time.Now().Add(time.Hour).After(booking.StartTime) {
		return errors.New("cannot cancel booking less than 1 hour before start time")
	}

	// Update booking status
	updates := map[string]interface{}{
		"status":        "cancelled",
		"cancel_reason": reason,
	}

	if err := s.db.Model(&booking).Updates(updates).Error; err != nil {
		return err
	}

	// Free up the slot
	s.db.Model(&models.AvailableSlot{}).Where(
		"expert_id = ? AND start_time <= ? AND end_time >= ?",
		booking.ExpertID, booking.StartTime, booking.EndTime,
	).Update("is_booked", false)

	return nil
}

func (s *BookingService) UpdateBookingStatus(bookingID uint, status string) error {
	return s.db.Model(&models.Booking{}).Where("id = ?", bookingID).Update("status", status).Error
}

func (s *BookingService) GetUpcomingBookings() ([]models.Booking, error) {
	var bookings []models.Booking
	err := s.db.Where("start_time > ? AND start_time < ? AND status IN (?)", 
		time.Now(), time.Now().Add(time.Hour*2), []string{"pending", "confirmed"}).
		Preload("User").
		Preload("Expert").
		Preload("Expert.User").
		Find(&bookings).Error
	return bookings, err
}