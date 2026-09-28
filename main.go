package main

import (
	echowebframework "intelligentBI/delivery/echo/router"
	"intelligentBI/repository/SQLServer"
	"log"
)

func main() {
	db, err := SQLServer.NewDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := echowebframework.Router(db); err != nil {
		log.Fatal(err)
	}
}
