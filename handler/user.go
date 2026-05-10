package handler

import (
	"encoding/json"
	"intelligentBI/repository/SQLServer"
	"intelligentBI/service"
	"io"
	"log"
	"net/http"
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
