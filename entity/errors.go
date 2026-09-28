package entity

import "errors"

// ErrPhoneNumberExists is returned by a user repository when an insert hits
// the unique phone-number index — i.e. another account already uses it
// (including one created concurrently, after any existence check).
var ErrPhoneNumberExists = errors.New("phone number already registered")

// ErrSessionNotFound is returned by a session repository when a session token
// is unknown or has expired.
var ErrSessionNotFound = errors.New("session not found")
