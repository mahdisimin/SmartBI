package main

import (
	"intelligentBI/router"
	"log"
)

func main() {
	err := router.Router()
	if err != nil {
		log.Fatal(err)
	}
}
