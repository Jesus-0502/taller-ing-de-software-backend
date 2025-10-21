package routes

import (
	"database/sql"
	"farmlands-backend/handlers"

	"github.com/gorilla/mux"
)

func RegisterProjectRoutes(router *mux.Router, db *sql.DB) {
	handler := handlers.NewProjectHandler(db)

	router.HandleFunc("/proyectos", handler.HandleListProjects).Methods("GET")
	router.HandleFunc("/proyectos", handler.HandleCreateProject).Methods("POST")
	router.HandleFunc("/proyectos/search", handler.HandleProjectQuery).Methods("GET")
	router.HandleFunc("/proyectos/delete", handler.HandleDeleteProject).Methods("POST")
}
