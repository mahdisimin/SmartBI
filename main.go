package main

import "log"

func main() {
	err := Router()
	if err != nil {
		log.Fatal(err)
	}
}
