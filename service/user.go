package service

import (
	"fmt"
	"intelligentBI/pkg"
)

type UserService struct {
	Repository Repo
}

type Repo interface {
	IsPhoneNumberExists(phoneNumber string) (bool, error)
	PersistUser(string, string, string) (int64, error)
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
