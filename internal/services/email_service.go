// internal/services/email_service.go
package services

import (
	"consultation-booking/internal/config"
	"fmt"
	"net/smtp"
)

type EmailService struct {
	smtpConfig config.SMTPConfig
}

func NewEmailService(smtpConfig config.SMTPConfig) *EmailService {
	return &EmailService{
		smtpConfig: smtpConfig,
	}
}

func (s *EmailService) SendEmail(to, subject, body string) error {
	auth := smtp.PlainAuth("", s.smtpConfig.Username, s.smtpConfig.Password, s.smtpConfig.Host)

	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to, subject, body)

	return smtp.SendMail(
		s.smtpConfig.Host+":"+s.smtpConfig.Port,
		auth,
		s.smtpConfig.Username,
		[]string{to},
		[]byte(msg),
	)
}

func (s *EmailService) SendWelcomeEmail(to, name string) error {
	subject := "Welcome to Consultation Booking System"
	body := fmt.Sprintf(`
		Dear %s,

		Welcome to our Consultation Booking System!

		You can now:
		- Browse available experts
		- Book consultations
		- Manage your appointments
		- Receive notifications

		Thank you for joining us!

		Best regards,
		Consultation Booking Team
	`, name)

	return s.SendEmail(to, subject, body)
}

func (s *EmailService) SendBookingConfirmation(to, expertName, startTime string) error {
	subject := "Booking Confirmation"
	body := fmt.Sprintf(`
		Your consultation booking has been confirmed!

		Expert: %s
		Time: %s

		Please be ready 5 minutes before the scheduled time.

		Best regards,
		Consultation Booking Team
	`, expertName, startTime)

	return s.SendEmail(to, subject, body)
}

func (s *EmailService) SendReminder(to, expertName, startTime string) error {
	subject := "Consultation Reminder"
	body := fmt.Sprintf(`
		Reminder: Your consultation is starting soon!

		Expert: %s
		Time: %s

		Please be ready for your consultation.

		Best regards,
		Consultation Booking Team
	`, expertName, startTime)

	return s.SendEmail(to, subject, body)
}
