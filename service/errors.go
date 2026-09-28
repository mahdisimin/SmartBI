package service

import (
	"errors"
	"strings"

	"intelligentBI/entity"
)

var (
	// ErrInvalidCredentials is the single error for every failed login, so a
	// caller cannot tell an unknown phone number from a wrong password.
	ErrInvalidCredentials = errors.New("invalid phone number or password")
	// ErrPhoneNumberExists is returned by Register when the phone number is
	// already registered.
	ErrPhoneNumberExists = entity.ErrPhoneNumberExists
	// ErrUserNotFound is returned by Profile when no user has the given ID.
	ErrUserNotFound = errors.New("user not found")
)

// ValidationError reports request fields that failed validation. Its message
// only describes the caller's own input, so it is safe to return to clients.
type ValidationError struct {
	Problems []string
}

func (e ValidationError) Error() string {
	return "invalid request: " + strings.Join(e.Problems, "; ")
}
