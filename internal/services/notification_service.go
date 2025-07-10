// internal/services/notification_service.go
package services

import (
	"consultation-booking/internal/models"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type NotificationService struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewNotificationService(db *gorm.DB, redis *redis.Client) *NotificationService {
	return &NotificationService{
		db:    db,
		redis: redis,
	}
}

func (s *NotificationService) CreateNotification(userID uint, title, message, notificationType string) error {
	notification := models.Notification{
		UserID:  userID,
		Title:   title,
		Message: message,
		Type:    notificationType,
	}

	return s.db.Create(&notification).Error
}

func (s *NotificationService) GetUserNotifications(userID uint, limit int) ([]models.Notification, error) {
	var notifications []models.Notification
	err := s.db.Where("user_id = ?", userID).
		Order("created_at desc").
		Limit(limit).
		Find(&notifications).Error
	return notifications, err
}

func (s *NotificationService) MarkAsRead(notificationID uint) error {
	return s.db.Model(&models.Notification{}).Where("id = ?", notificationID).Update("is_read", true).Error
}

func (s *NotificationService) GetUnreadCount(userID uint) (int64, error) {
	var count int64
	err := s.db.Model(&models.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&count).Error
	return count, err
}
