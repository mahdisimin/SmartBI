package service

import (
	"fmt"
	"intelligentBI/entity"
	"intelligentBI/pkg"
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
	UserId int64
}

func (u UserService) Register(request UserRegisterRequest) (response UserRegisterResponse, err error) {
	var userID int64

	if isExist, err := u.Repository.IsPhoneNumberExists(request.PhoneNumber); err != nil || isExist {
		if err != nil {
			return response, fmt.Errorf("database error %v", err)
		}
		return response, fmt.Errorf("phone number %v already exists", request.PhoneNumber)
	}

	// TODO - Validate Password base on company policies

	hashPassword := pkg.HashStringMD5(request.Password)

	if userIDTemp, err := u.Repository.PersistUser(request.UserName, request.PhoneNumber, hashPassword); err != nil {
		return response, fmt.Errorf("database error: %v", err)
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

func (u UserService) Login(request UserLoginRequest) (response UserLoginResponse, err error) {
	var password string
	var userId int64
	if isExists, err := u.Repository.IsPhoneNumberExists(request.PhoneNumber); err != nil || !isExists {
		if err != nil {
			return response, fmt.Errorf("database error %v", err)
		} else if !isExists {
			return response, fmt.Errorf("phone number %s does NOT exists", request.PhoneNumber)
		}
	}

	passwordTemp, err := u.Repository.GetPasswordByPhoneNumber(request.PhoneNumber)
	if err != nil {
		return response, fmt.Errorf("database error: %v", err)
	} else {
		password = passwordTemp
	}
	hashedPassword := pkg.HashStringMD5(request.Password)
	if hashedPassword != password {
		return response, fmt.Errorf("password is NOT correct")
	}

	if userIdTemp, err := u.Repository.GetUserIDByPhoneNumber(request.PhoneNumber); err != nil {
		return response, fmt.Errorf("database error: %v", err)
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

func (u UserService) Profile(request UserProfileRequest) (response UserProfileResponse, err error) {
	var userName string
	var userLinkList []entity.WebAppList
	var userPhoneNumber string
	userTemp, err := u.Repository.GetUserByUserID(request.UserID)
	if err != nil {
		return response, fmt.Errorf("database error: %v", err)
	} else {
		userName = userTemp.UserName
		userLinkList = userTemp.WebAppList
		userPhoneNumber = userTemp.PhoneNumber
	}
	response = UserProfileResponse{
		UserName:     userName,
		UserLinkList: userLinkList,
		UserPhone:    userPhoneNumber,
	}
	return response, nil
}
