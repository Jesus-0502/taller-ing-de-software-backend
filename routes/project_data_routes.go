package routes

import (
	"database/sql"
	"farmlands-backend/handlers"

	"github.com/gorilla/mux"
)

func RegisterProjectDataRoutes(router *mux.Router, db *sql.DB) {
	handler := handlers.NewProjectDataHandler(db)

	router.HandleFunc("/project_data", handler.HandleNewProjectData).Methods("POST")
	// router.HandleFunc("/project_data", handler.HandleSearchTool).Methods("GET")
	// router.HandleFunc("/project_data/edit", handler.HandleEditTool).Methods("POST")
	router.HandleFunc("/project_data/delete", handler.HandleDeleteProjectData).Methods("POST")
}
