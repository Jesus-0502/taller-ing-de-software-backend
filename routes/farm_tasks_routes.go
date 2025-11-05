package routes

import (
	"database/sql"
	"farmlands-backend/handlers"

	"github.com/gorilla/mux"
)

func RegisterFarmTasksRoutes(router *mux.Router, db *sql.DB) {
	handler := handlers.NewFarmTasksHandlerHandler(db)

	router.HandleFunc("/farm_tasks", handler.HandleListFarmTasks).Methods("GET")

}
