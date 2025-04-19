package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmailNotifier_SendSecurityAlert(t *testing.T) {
	notifier := NewEmailNotifier()
	err := notifier.SendSecurityAlert("user1", "test message")
	assert.NoError(t, err)
}
