package main

import (
	"fmt"
	"intelligentBI/router"
	"log"
)

func main() {
	fmt.Println("hello world")
	err := router.Router()
	if err != nil {
		log.Fatal(err)
	}
}
