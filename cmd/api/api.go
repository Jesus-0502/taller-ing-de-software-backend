package api

import (
	"database/sql"
	"log"
	"net/http"
	"taller-ing-de-software-backend/middleware"
	"taller-ing-de-software-backend/service/user"

	"github.com/gorilla/mux"
)

type APIServer struct {
	addr string
	db   *sql.DB
}

func NewAPIServer(addr string, database *sql.DB) *APIServer {
	return &APIServer{
		addr: addr,
		db:   database,
	}
}

func (s *APIServer) Run() error {
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/api/usuario").Subrouter()

	userHandler := user.NewHandler(s.db)
	userHandler.RegisterRoutes(subrouter)

	log.Println("Listening on", s.addr)
	handler := middleware.CorsMiddleware(router)

	return http.ListenAndServe(s.addr, handler)

}
