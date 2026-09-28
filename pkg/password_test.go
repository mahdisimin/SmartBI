package pkg

import (
	"strings"
	"testing"
)

func TestPasswordHashing(t *testing.T) {
	hash, err := HashPassword("p@ss' OR 1=1--")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(hash, "$2") || len(hash) != 60 {
		t.Fatalf("not a bcrypt hash: %q", hash)
	}
	if !CheckPassword(hash, "p@ss' OR 1=1--") {
		t.Error("correct password rejected")
	}
	if CheckPassword(hash, "p@ss' OR 1=1") || CheckPassword(hash, "") {
		t.Error("wrong password accepted")
	}

	// Salted: the same password never hashes the same way twice.
	again, _ := HashPassword("p@ss' OR 1=1--")
	if again == hash {
		t.Error("hashes are not salted")
	}

	// A legacy MD5 hex digest is not a valid bcrypt hash and never matches.
	if CheckPassword(HashStringMD5("secret"), "secret") {
		t.Error("MD5 hash accepted as a password hash")
	}
}
