package models

// User es el modelo para la base de datos
type User struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Lastname     string `json:"lastname"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`    // No se debe exponer nunca en respuestas
	Role         string `json:"role"` // "admin" o "user"
}

type UserFullData struct {
	ID           int64  `json:"id"`
	CI 			 string `json:"ci"`
	Name         string `json:"name"`
	Lastname     string `json:"lastname"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`    // No se debe exponer nunca en respuestas
	Role         string `json:"role"` // "admin" o "user"
}

// CreateUserInput es lo que recibimos del cliente para crear el usuario
type CreateUserInput struct {
	Name     string `json:"name"`
	CI 	     string `json:"ci"`
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
