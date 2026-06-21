package nethttp

import (
	"encoding/json"
	"fmt"
	"intelligentBI/repository/SQLServer"
	"intelligentBI/service"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func UserRegisterHandler(w http.ResponseWriter, r *http.Request) {
	var userRegReq service.UserRegisterRequest
	var userRegResp service.UserRegisterResponse
	var data = make([]byte, r.ContentLength)
	var userServ = service.UserService{
		Repository: SQLServer.SqlServer{},
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("error on HTTP request , wrong method")
		return
	}
	dataTemp, errTemp := io.ReadAll(r.Body)
	if errTemp != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("error on HTTP request , read body error")

		return
	} else {
		data = dataTemp
	}
	if err := json.Unmarshal(data, &userRegReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("error on HTTP request , unmarshal user register request error")

		return
	}
	userRegRespTemp, err := userServ.Register(userRegReq)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		log.Printf("error on HTTP request , register error : %v", err)

		return
	} else {
		userRegResp = userRegRespTemp
	}
	userRegRespData, err := json.Marshal(userRegResp)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		log.Printf("error on HTTP request , marshal user register response error")

		return
	}
	if _, err := w.Write(userRegRespData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		log.Printf("error on HTTP request , write response error")

		return
	}

}

func UserLoginHandler(w http.ResponseWriter, r *http.Request) {
	var userLoginReq service.UserLoginRequest
	var userLoginResp service.UserLoginResponse
	var data = make([]byte, r.ContentLength)
	var userServ = service.UserService{
		Repository: SQLServer.SqlServer{},
	}
	fmt.Println(r.Method)
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("error on HTTP request , wrong method")

		return
	}
	dataTemp, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("error on HTTP request , read body error : %v", err)

		return
	} else {
		data = dataTemp
	}
	if err := json.Unmarshal(data, &userLoginReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("error on HTTP request , unmarshal user login request error : %v", err)

		return
	}
	userLoginRespTemp, err := userServ.Login(userLoginReq)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		log.Printf("error on HTTP request , login error : %v", err)
		return
	} else {
		userLoginResp = userLoginRespTemp
	}
	userLoginRespData, err := json.Marshal(userLoginResp)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		log.Printf("error on HTTP request , marshal user login response error : %v", err)

		return
	}
	if _, err := w.Write(userLoginRespData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		log.Printf("error on HTTP request , write response error : %v", err)

		return
	}

}

func UserProfileHandler(w http.ResponseWriter, r *http.Request) {
	var userId int64
	var userProfileResp service.UserProfileResponse
	var userServ = service.UserService{
		Repository: SQLServer.SqlServer{},
	}
	params := mux.Vars(r)

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("error on HTTP request , wrong method")

		return
	}
	idStr := params["id"]
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("error on HTTP request , empty user id : %s", idStr)

		return
	}
	userIDTemp, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("error on HTTP request , convert to int error : %v", err)

		return
	} else {
		userId = int64(userIDTemp)
	}

	userProfileRequest := service.UserProfileRequest{
		UserID: userId,
	}
	userProfileRespTemp, err := userServ.Profile(userProfileRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("error on HTTP request , call service : %v", err)

		return
	} else {
		userProfileResp = userProfileRespTemp
	}
	userProfileRespData, err := json.Marshal(userProfileResp)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		log.Printf("error on HTTP request , marshal user profile response error : %v", err)

		return
	}
	if _, err := w.Write(userProfileRespData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		log.Printf("error on HTTP request , write response error : %v", err)

		return
	}
}
