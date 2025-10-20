package api

import (
	"database/sql"
	"farmlands-backend/middleware"
	"farmlands-backend/routes"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type APIServer struct {
	Addr string
	DB   *sql.DB
}

func NewAPIServer(addr string, db *sql.DB) *APIServer {
	return &APIServer{
		Addr: addr,
		DB:   db,
	}
}

func (s *APIServer) Run() error {
	router := mux.NewRouter()

	// 👉 Rutas base de tu API
	apiRouter := router.PathPrefix("/api").Subrouter()

	// 👉 Registrar rutas específicas
	routes.RegisterUserRoutes(apiRouter, s.DB)
	//routes.RegisterProjectRoutes(apiRouter, s.DB) // cuando lo tengas

	// 👉 Middlewares globales
	handler := middleware.CorsMiddleware(router)

	log.Println("Servidor corriendo en", s.Addr)
	return http.ListenAndServe(s.Addr, handler)
}
