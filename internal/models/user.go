// internal/models/user.go
package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Email       string         `json:"email" gorm:"unique;not null"`
	Password    string         `json:"-" gorm:"not null"`
	Name        string         `json:"name"`
	Phone       string         `json:"phone"`
	Avatar      string         `json:"avatar"`
	Description string         `json:"description"`
	Gender      string         `json:"gender"`
	Role        string         `json:"role" gorm:"default:user"` // user, expert, admin
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Bookings      []Booking      `json:"bookings,omitempty"`
	Notifications []Notification `json:"notifications,omitempty"`
}

type Expert struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	UserID      uint           `json:"user_id" gorm:"not null"`
	User        User           `json:"user"`
	Speciality  string         `json:"speciality"`
	Experience  int            `json:"experience"`
	Rating      float64        `json:"rating" gorm:"default:0"`
	IsAvailable bool           `json:"is_available" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	AvailableSlots []AvailableSlot `json:"available_slots,omitempty"`
	Bookings       []Booking       `json:"bookings,omitempty"`
}

type Booking struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	UserID       uint           `json:"user_id" gorm:"not null"`
	User         User           `json:"user"`
	ExpertID     uint           `json:"expert_id" gorm:"not null"`
	Expert       Expert         `json:"expert"`
	StartTime    time.Time      `json:"start_time" gorm:"not null"`
	EndTime      time.Time      `json:"end_time" gorm:"not null"`
	Status       string         `json:"status" gorm:"default:pending"` // pending, confirmed, rejected, cancelled, completed
	Notes        string         `json:"notes"`
	Format       string         `json:"format"` // online, offline
	CancelReason string         `json:"cancel_reason"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Feedback *Feedback `json:"feedback,omitempty"`
}

type AvailableSlot struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	ExpertID  uint           `json:"expert_id" gorm:"not null"`
	Expert    Expert         `json:"expert"`
	StartTime time.Time      `json:"start_time" gorm:"not null"`
	EndTime   time.Time      `json:"end_time" gorm:"not null"`
	IsBooked  bool           `json:"is_booked" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type Notification struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	User      User           `json:"user"`
	Title     string         `json:"title" gorm:"not null"`
	Message   string         `json:"message" gorm:"not null"`
	Type      string         `json:"type"` // booking, reminder, cancellation
	IsRead    bool           `json:"is_read" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type Feedback struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	BookingID uint           `json:"booking_id" gorm:"not null"`
	Booking   Booking        `json:"booking"`
	Rating    int            `json:"rating" gorm:"not null"`
	Comment   string         `json:"comment"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
