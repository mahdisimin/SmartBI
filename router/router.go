package router

import (
	"intelligentBI/handler"
	"intelligentBI/middlewear"
	"log"
	"net/http"

	muxlib "github.com/gorilla/mux"
)

func Router() error {
	mux := muxlib.NewRouter()
	mux.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	mux.HandleFunc("/user/register", handler.UserRegisterHandler)
	mux.HandleFunc("/user/login", handler.UserLoginHandler)
	mux.HandleFunc("/user/user_profile/{id}", handler.UserProfileHandler)

	log.Println("Server is running on Localhost:8091 ...")
	if err := http.ListenAndServe("localhost:8091", middlewear.CorsMiddleware(mux)); err != nil {
		log.Printf("error on HTTP request , %v", err)
		return err
	}
	return nil
}
