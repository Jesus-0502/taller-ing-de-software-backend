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
	router.HandleFunc("/proyectos/update", handler.HandleUpdateProject).Methods("POST")
	router.HandleFunc("/proyectos/update-status", handler.HandleUpdateProjectStatus).Methods("POST")
	router.HandleFunc("/proyectos/download", handler.HandleDownloadProjects).Methods("GET")
	router.HandleFunc("/proyectos/userProjects", handler.HandleGetUserProjects).Methods("POST")
	router.HandleFunc("/proyectos/availableForUser", handler.HandleGetAvailableProjectsForUser).Methods("POST")
	router.HandleFunc("/proyectos/asignar", handler.HandleAssignUserToProject).Methods("POST")
	router.HandleFunc("/proyectos/changeAssociation", handler.HandleChangeAssociation).Methods("POST")
	router.HandleFunc("/proyectos/removeAssociation", handler.HandleRemoveAssociation).Methods("POST")
}
