package pkg

import (
	"crypto/md5"
	"encoding/hex"
)

func HashStringMD5(string string) string {
	hash := md5.Sum([]byte(string))
	hashString := hex.EncodeToString(hash[:])
	return hashString
}
