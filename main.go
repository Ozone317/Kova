package main

import (
	"log"

	"github.com/Ozone317/Kova/server"
)

func main() {
	err := server.RunAsyncTCPServer()
	if err != nil {
		log.Fatal(err)
	}
}
