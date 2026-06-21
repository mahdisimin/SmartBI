package main

import (
	"intelligentBI/delivery/nethttp/router"
	"log"
)

func main() {
	err := nethttp.Router()
	if err != nil {
		log.Fatal(err)
	}
}
