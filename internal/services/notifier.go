package services

import "log"

type EmailNotifier struct{}

func NewEmailNotifier() *EmailNotifier {
	return &EmailNotifier{}
}

func (n *EmailNotifier) SendSecurityAlert(userID, message string) error {
	log.Printf("Email alert for user %s: %s", userID, message)
	return nil
}
