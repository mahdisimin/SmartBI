package pkg

import (
	"sync"

	"golang.org/x/crypto/bcrypt"
)

// bcryptCost is the work factor for password hashes (~250ms per hash on
// current hardware — slow enough to make offline guessing expensive).
const bcryptCost = 12

// HashPassword returns a salted bcrypt hash of password, suitable for
// storing. bcrypt only reads the first 72 bytes; callers must reject longer
// passwords rather than silently truncating them.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword reports whether password matches a hash from HashPassword.
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// dummyHash is compared against when there is no real hash to check (e.g.
// an unknown phone number at login), so that path costs as much time as a
// real check and response timing does not reveal whether an account exists.
var dummyHash = sync.OnceValue(func() string {
	hash, _ := HashPassword("not-a-real-password")
	return hash
})

// SpendPasswordCheckTime performs a password comparison that always fails,
// taking the same time as CheckPassword.
func SpendPasswordCheckTime(password string) {
	CheckPassword(dummyHash(), password)
}
