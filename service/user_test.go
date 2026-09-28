package service

import (
	"errors"
	"testing"

	"intelligentBI/entity"
)

// racingRepo models losing a concurrent registration race: the existence
// check passes, then the insert hits the unique index.
type racingRepo struct{ Repo }

func (racingRepo) IsPhoneNumberExists(string) (bool, error) { return false, nil }
func (racingRepo) PersistUser(string, string, string) (int64, error) {
	return 0, entity.ErrPhoneNumberExists
}

func TestRegister_LostRaceReportsPhoneExists(t *testing.T) {
	svc := UserService{Repository: racingRepo{}}
	_, err := svc.Register(UserRegisterRequest{UserName: "bob", PhoneNumber: "09120000002", Password: "pw"})
	if !errors.Is(err, ErrPhoneNumberExists) {
		t.Fatalf("want ErrPhoneNumberExists, got %v", err)
	}
}

func TestRegister_Validation(t *testing.T) {
	svc := UserService{Repository: racingRepo{}}
	for name, req := range map[string]UserRegisterRequest{
		"blank name":          {UserName: "  ", PhoneNumber: "0912", Password: "pw"},
		"phone with space":    {UserName: "bob", PhoneNumber: "0912 000", Password: "pw"},
		"non-ascii phone":     {UserName: "bob", PhoneNumber: "۰۹۱۲", Password: "pw"},
		"password > 72 bytes": {UserName: "bob", PhoneNumber: "0912", Password: string(make([]byte, 73))},
	} {
		t.Run(name, func(t *testing.T) {
			var ve ValidationError
			if _, err := svc.Register(req); !errors.As(err, &ve) {
				t.Fatalf("want ValidationError, got %v", err)
			}
		})
	}
}
