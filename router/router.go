package router

import (
	"encoding/json"
	"intelligentBI/handler"
	"log"
	"net/http"
	"os"
)

func Router() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	mux.HandleFunc("/config.js", configHandler)
	mux.HandleFunc("/user/register", handler.UserRegisterHandler)
	mux.HandleFunc("/user/login", handler.UserLoginHandler)
	mux.Handle("/", http.FileServer(http.Dir("web")))

	log.Println("Server is running on Localhost:8090 ...")
	if err := http.ListenAndServe("localhost:8090", mux); err != nil {
		log.Printf("error on HTTP request , %v", err)
		return err
	}
	return nil
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	baseURL, err := json.Marshal(os.Getenv("SMARTBI_API_BASE_URL"))
	if err != nil {
		baseURL = []byte(`""`)
	}
	w.Write([]byte("window.SMARTBI_API_BASE_URL = "))
	w.Write(baseURL)
	w.Write([]byte(";"))
}
