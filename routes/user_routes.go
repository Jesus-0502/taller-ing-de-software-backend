package routes

import (
	"database/sql"
	"farmlands-backend/handlers"

	"github.com/gorilla/mux"
)

func RegisterUserRoutes(router *mux.Router, db *sql.DB) {
	handler := handlers.NewUserHandler(db)
	router.HandleFunc("/usuario/login", handler.HandleLogin).Methods("POST")
	router.HandleFunc("/usuario/register", handler.HandleRegister).Methods("POST")
	router.HandleFunc("/usuario/listUsers", handler.HandleListUsers).Methods("GET")
	router.HandleFunc("/usuario/search", handler.HandleUserQuery).Methods("GET")
	router.HandleFunc("/usuario/delete", handler.HandleDeleteUser).Methods("POST")
	router.HandleFunc("/usuario/edit", handler.HandleEditUser).Methods("POST")
}
