// internal/services/expert_service.go
package services

import (
	"consultation-booking/internal/models"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type ExpertService struct {
	db    *gorm.DB
	redis *redis.Client
}

type CreateSlotRequest struct {
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" binding:"required"`
}

func NewExpertService(db *gorm.DB, redis *redis.Client) *ExpertService {
	return &ExpertService{
		db:    db,
		redis: redis,
	}
}

func (s *ExpertService) CreateExpert(userID uint, speciality string, experience int) error {
	expert := models.Expert{
		UserID:     userID,
		Speciality: speciality,
		Experience: experience,
	}

	if err := s.db.Create(&expert).Error; err != nil {
		return err
	}

	// Update user role
	return s.db.Model(&models.User{}).Where("id = ?", userID).Update("role", "expert").Error
}

func (s *ExpertService) GetExperts() ([]models.Expert, error) {
	var experts []models.Expert
	err := s.db.Preload("User").Where("is_available = ?", true).Find(&experts).Error
	return experts, err
}

func (s *ExpertService) GetExpertByID(expertID uint) (*models.Expert, error) {
	var expert models.Expert
	err := s.db.Preload("User").First(&expert, expertID).Error
	return &expert, err
}

func (s *ExpertService) CreateAvailableSlot(expertID uint, req CreateSlotRequest) error {
	// Check if expert exists
	var expert models.Expert
	if err := s.db.First(&expert, expertID).Error; err != nil {
		return err
	}

	// Check for conflicts
	var conflictCount int64
	s.db.Model(&models.AvailableSlot{}).Where(
		"expert_id = ? AND ((start_time <= ? AND end_time > ?) OR (start_time < ? AND end_time >= ?))",
		expertID, req.StartTime, req.StartTime, req.EndTime, req.EndTime,
	).Count(&conflictCount)

	if conflictCount > 0 {
		return errors.New("time slot conflicts with existing slot")
	}

	slot := models.AvailableSlot{
		ExpertID:  expertID,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}

	if err := s.db.Create(&slot).Error; err != nil {
		return err
	}

	// Update cache
	s.updateAvailableSlotsCache(expertID)
	return nil
}

func (s *ExpertService) GetAvailableSlots(expertID uint) ([]models.AvailableSlot, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("available_slots:%d", expertID)
	cached, err := s.redis.Get(context.Background(), cacheKey).Result()
	if err == nil {
		var slots []models.AvailableSlot
		if err := json.Unmarshal([]byte(cached), &slots); err == nil {
			return slots, nil
		}
	}

	// Get from database
	var slots []models.AvailableSlot
	err = s.db.Where("expert_id = ? AND is_booked = ? AND start_time > ?",
		expertID, false, time.Now()).
		Order("start_time").
		Find(&slots).Error

	if err != nil {
		return nil, err
	}

	// Cache the result
	if data, err := json.Marshal(slots); err == nil {
		s.redis.Set(context.Background(), cacheKey, string(data), time.Hour)
	}

	return slots, nil
}

func (s *ExpertService) GetExpertBookings(expertID uint) ([]models.Booking, error) {
	var bookings []models.Booking
	err := s.db.Where("expert_id = ?", expertID).
		Preload("User").
		Order("start_time").
		Find(&bookings).Error
	return bookings, err
}

func (s *ExpertService) UpdateBookingStatus(bookingID uint, status string) error {
	return s.db.Model(&models.Booking{}).Where("id = ?", bookingID).Update("status", status).Error
}

func (s *ExpertService) updateAvailableSlotsCache(expertID uint) {
	var slots []models.AvailableSlot
	s.db.Where("expert_id = ? AND is_booked = ? AND start_time > ?",
		expertID, false, time.Now()).
		Order("start_time").
		Find(&slots)

	cacheKey := fmt.Sprintf("available_slots:%d", expertID)
	if data, err := json.Marshal(slots); err == nil {
		s.redis.Set(context.Background(), cacheKey, string(data), time.Hour)
	}
}
