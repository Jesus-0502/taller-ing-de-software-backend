package main

import (
	"farmlands-backend/api"
	"farmlands-backend/db"
	"log"
)

func main() {
	// Crear la conexión con la base de datos
	database, err := db.NewDB("app.db")
	if err != nil {
		log.Fatal(err)
	}

	// Crear y lanzar el servidor
	server := api.NewAPIServer(":8080", database)
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
