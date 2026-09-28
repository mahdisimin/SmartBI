package service

import (
	"database/sql"
	"errors"
	"fmt"
	"intelligentBI/entity"
	"intelligentBI/pkg"
	"strings"
	"unicode"
	"unicode/utf16"
)

type UserService struct {
	Repository Repo
}

type Repo interface {
	IsPhoneNumberExists(phoneNumber string) (bool, error)
	PersistUser(string, string, string) (int64, error)
	GetPasswordByPhoneNumber(phoneNumber string) (string, error)
	GetUserIDByPhoneNumber(phoneNumber string) (int64, error)
	GetUserByUserID(userID int64) (entity.User, error)
}

func NewUserService(repo Repo) *UserService {
	return &UserService{
		Repository: repo,
	}
}

type UserRegisterRequest struct {
	UserName    string `json:"user_name"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
}

type UserRegisterResponse struct {
	UserId int64 `json:"user_id"`
}

// Field limits, matching the APP.[USER] columns.
const (
	maxUserNameLen    = 500 // NVARCHAR(500), in UTF-16 code units
	maxPhoneNumberLen = 100 // VARCHAR(100), in bytes
	// maxPasswordBytes is bcrypt's input limit, enforced now so no password
	// accepted today becomes unusable after the planned move to bcrypt.
	maxPasswordBytes = 72
)

// validate trims the request's text fields in place and checks them.
func (r *UserRegisterRequest) validate() error {
	r.UserName = strings.TrimSpace(r.UserName)
	r.PhoneNumber = strings.TrimSpace(r.PhoneNumber)

	var problems []string
	if r.UserName == "" {
		problems = append(problems, "user_name is required")
	} else if len(utf16.Encode([]rune(r.UserName))) > maxUserNameLen {
		problems = append(problems, fmt.Sprintf("user_name must be at most %d characters", maxUserNameLen))
	}

	switch {
	case r.PhoneNumber == "":
		problems = append(problems, "phone_number is required")
	case len(r.PhoneNumber) > maxPhoneNumberLen:
		problems = append(problems, fmt.Sprintf("phone_number must be at most %d characters", maxPhoneNumberLen))
	case !isPrintableASCIINoSpace(r.PhoneNumber):
		problems = append(problems, "phone_number contains invalid characters")
	}

	// TODO - Validate Password base on company policies
	if r.Password == "" {
		problems = append(problems, "password is required")
	} else if len(r.Password) > maxPasswordBytes {
		problems = append(problems, fmt.Sprintf("password must be at most %d bytes", maxPasswordBytes))
	}

	if len(problems) > 0 {
		return ValidationError{Problems: problems}
	}
	return nil
}

// isPrintableASCIINoSpace reports whether s holds only visible ASCII — what a
// VARCHAR phone number column can store losslessly.
func isPrintableASCIINoSpace(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII || !unicode.IsPrint(r) || unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

func (u UserService) Register(request UserRegisterRequest) (response UserRegisterResponse, err error) {
	var userID int64

	if err := request.validate(); err != nil {
		return response, err
	}

	if isExist, err := u.Repository.IsPhoneNumberExists(request.PhoneNumber); err != nil || isExist {
		if err != nil {
			return response, fmt.Errorf("check phone number: %w", err)
		}
		return response, ErrPhoneNumberExists
	}

	hashPassword, err := pkg.HashPassword(request.Password)
	if err != nil {
		return response, fmt.Errorf("hash password: %w", err)
	}

	// The check above is only a fast path: two concurrent registrations can
	// both pass it. The unique index decides, and the repository reports the
	// loser as ErrPhoneNumberExists.
	if userIDTemp, err := u.Repository.PersistUser(request.UserName, request.PhoneNumber, hashPassword); err != nil {
		if errors.Is(err, ErrPhoneNumberExists) {
			return response, ErrPhoneNumberExists
		}
		return response, fmt.Errorf("persist user: %w", err)
	} else {
		userID = userIDTemp
	}

	return UserRegisterResponse{
		UserId: userID,
	}, nil
}

type UserLoginRequest struct {
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
}
type UserLoginResponse struct {
	UserId int64 `json:"user_id"`
}

// Login returns ErrInvalidCredentials for an unknown phone number and for a
// wrong password alike, so responses never reveal which accounts exist.
func (u UserService) Login(request UserLoginRequest) (response UserLoginResponse, err error) {
	var password string
	var userId int64
	request.PhoneNumber = strings.TrimSpace(request.PhoneNumber)
	if request.PhoneNumber == "" || request.Password == "" {
		return response, ErrInvalidCredentials
	}

	if isExists, err := u.Repository.IsPhoneNumberExists(request.PhoneNumber); err != nil || !isExists {
		if err != nil {
			return response, fmt.Errorf("check phone number: %w", err)
		}
		// Spend a bcrypt comparison anyway, so an unknown phone number takes
		// as long as a wrong password and timing cannot enumerate accounts.
		pkg.SpendPasswordCheckTime(request.Password)
		return response, ErrInvalidCredentials
	}

	passwordTemp, err := u.Repository.GetPasswordByPhoneNumber(request.PhoneNumber)
	if err != nil {
		return response, fmt.Errorf("get password: %w", err)
	} else {
		password = passwordTemp
	}
	if !pkg.CheckPassword(password, request.Password) {
		return response, ErrInvalidCredentials
	}

	if userIdTemp, err := u.Repository.GetUserIDByPhoneNumber(request.PhoneNumber); err != nil {
		return response, fmt.Errorf("get user id: %w", err)
	} else {
		userId = userIdTemp
	}

	response = UserLoginResponse{
		UserId: userId,
	}
	return response, nil
}

type UserProfileRequest struct {
	UserID int64
}
type UserProfileResponse struct {
	UserName     string              `json:"user_name"`
	UserPhone    string              `json:"user_phone"`
	UserLinkList []entity.WebAppList `json:"user_link_list"`
}

// Profile returns ErrUserNotFound when no user has the requested ID.
func (u UserService) Profile(request UserProfileRequest) (response UserProfileResponse, err error) {
	var userName string
	var userLinkList []entity.WebAppList
	var userPhoneNumber string
	userTemp, err := u.Repository.GetUserByUserID(request.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response, ErrUserNotFound
		}
		return response, fmt.Errorf("get user: %w", err)
	} else {
		userName = userTemp.UserName
		userLinkList = userTemp.WebAppList
		userPhoneNumber = userTemp.PhoneNumber
	}
	if userLinkList == nil {
		userLinkList = []entity.WebAppList{} // serialize as [], not null
	}
	response = UserProfileResponse{
		UserName:     userName,
		UserLinkList: userLinkList,
		UserPhone:    userPhoneNumber,
	}
	return response, nil
}
