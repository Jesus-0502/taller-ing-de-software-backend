package main

import (
	"log"
	"taller-ing-de-software-backend/cmd/api"
	"taller-ing-de-software-backend/db"
)

func main() {

	// Crea la conexión a la base de datos y realiza migración
	database, err := db.NewDB("app.db")
	if err != nil {
		log.Fatalf("No se pudo conectar a la base de datos: %v", err)
	}

	server := api.NewAPIServer(":8080", database)
	if err := server.Run(); err != nil {
		log.Fatal()
	}

}
