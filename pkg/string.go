package pkg

import (
	"crypto/md5"
	"encoding/hex"
)

// Deprecated: MD5 is not a password hash — use HashPassword/CheckPassword.
// Kept only until the QA handler test fixture stops seeding with it.
func HashStringMD5(string string) string {
	hash := md5.Sum([]byte(string))
	hashString := hex.EncodeToString(hash[:])
	return hashString
}
