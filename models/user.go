package models

import "time"

// User es el modelo para la base de datos
type User struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Lastname     string    `json:"lastname"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`          // No se debe exponer nunca en respuestas
	Role         string    `json:"role"`       // "admin" o "user"
	CreatedAt    time.Time `json:"created_at"` // Fecha de creación
}

// CreateUserInput es lo que recibimos del cliente para crear el usuario
type CreateUserInput struct {
	Name     string `json:"name"`
	Lastname string `json:"lastname"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
