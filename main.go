package main

import (
	echowebframework "intelligentBI/delivery/echo/router"
	"log"
)

func main() {
	//err := nethttp.Router()
	//if err != nil {
	//	log.Fatal(err)
	//}

	if err := echowebframework.Router(); err != nil {
		log.Fatal(err)
	}
}
