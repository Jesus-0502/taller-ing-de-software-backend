package main

import (
	"log"
	"taller-ing-de-software-backend/cmd/api"
)

func main() {
	server := api.NewAPIServer(":8080", nil)
	if err := server.Run(); err != nil {
		log.Fatal()
	}

}
