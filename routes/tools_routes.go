package routes

import (
	"database/sql"
	"farmlands-backend/handlers"

	"github.com/gorilla/mux"
)

func RegisterToolsRoutes(router *mux.Router, db *sql.DB) {
	handler := handlers.NewToolsHandler(db)

	router.HandleFunc("/tools", handler.HandleListTools).Methods("GET")
}
