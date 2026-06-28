package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	println("Hello World")

	Body, _ := json.Marshal(map[string]string{
		"name":  "Toby",
		"email": "mahdisimin@gmail.com",
	})

	client := &http.Client{
		Timeout: time.Second * 5,
	}

	reqBody := bytes.NewBuffer(Body)
	req, _ := http.NewRequest(http.MethodPost, "https://postman-echo.com/post", reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("HeaderTest", "Test")
	if respTemp, err := client.Do(req); err != nil {
		fmt.Println(err)
	} else {
		defer respTemp.Body.Close()
		body, _ := io.ReadAll(respTemp.Body)
		fmt.Println(string(body))
	}

}
