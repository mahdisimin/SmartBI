package router

import (
	"intelligentBI/handler"
	"log"
	"net/http"
)

func Router() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	mux.HandleFunc("/user/register", handler.UserRegisterHandler)
	mux.HandleFunc("/user/login", handler.UserLoginHandler)

	log.Println("Server is running on Localhost:8090 ...")
	if err := http.ListenAndServe("localhost:8090", mux); err != nil {
		log.Printf("error on HTTP request , %v", err)
		return err
	}
	return nil
}
