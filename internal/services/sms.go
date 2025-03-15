package services

import (
	"fmt"
	"log"
)

// SMSService définit l'interface pour un service d'envoi de SMS
type SMSService interface {
	SendSMS(phoneNumber, message string) error
	SendBulkSMS(phoneNumbers []string, message string) ([]string, []error)
}

// MockSMSService est une implémentation simulée du service SMS pour le développement
type MockSMSService struct{}

// NewMockSMSService crée une nouvelle instance de MockSMSService
func NewMockSMSService() *MockSMSService {
	return &MockSMSService{}
}

// SendSMS simule l'envoi d'un SMS à un numéro
func (s *MockSMSService) SendSMS(phoneNumber, message string) error {
	log.Printf("SIMULATION SMS: Envoi à %s: \"%s\"", phoneNumber, message)
	return nil
}

// SendBulkSMS simule l'envoi de SMS en masse
func (s *MockSMSService) SendBulkSMS(phoneNumbers []string, message string) ([]string, []error) {
	success := make([]string, 0)
	failures := make([]error, 0)

	for _, phone := range phoneNumbers {
		err := s.SendSMS(phone, message)
		if err != nil {
			failures = append(failures, fmt.Errorf("erreur pour %s: %w", phone, err))
		} else {
			success = append(success, phone)
		}
	}

	return success, failures
}

// Singleton instance du service SMS
var DefaultSMSService SMSService = NewMockSMSService()
