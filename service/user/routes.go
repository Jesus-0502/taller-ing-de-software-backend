package user

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/login", h.handleLogin).Methods("POST")
	router.HandleFunc("/register", h.handleRegister).Methods("POST")

}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var input LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if input.Email == "" || input.Password == "" {
		http.Error(w, "Email y contraseña requeridos", http.StatusBadRequest)
		return
	}

	// Se hace la query para buscar el usuario en la base de datos por email
	stmt := `SELECT id, name, email, password_hash, role, created_at FROM users WHERE email = ? LIMIT 1`
	var u User
	var createdAtStr string
	err := h.db.QueryRow(stmt, input.Email).Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &createdAtStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Credenciales incorrectas", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}
	u.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)

	// Compara el password recibido con el guardado
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(input.Password)); err != nil {
		http.Error(w, "Credenciales incorrectas", http.StatusUnauthorized)
		return
	}

	// Si todo está bien, responde con los datos del usuario
	resp := struct {
		ID        int64     `json:"id"`
		Name      string    `json:"name"`
		Email     string    `json:"email"`
		Role      string    `json:"role"`
		CreatedAt time.Time `json:"created_at"`
	}{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var input CreateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validaciones básicas
	if input.Name == "" || input.Email == "" || len(input.Password) < 6 {
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	// Hashear la contraseña
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error al hashear contraseña", http.StatusInternalServerError)
		return
	}

	// Insertar en la base de datos
	stmt := `
	INSERT INTO users (name, email, password_hash, role, created_at)
	VALUES (?, ?, ?, ?, ?)
	`
	role := input.Role
	if role == "" {
		role = "user"
	}

	createdAt := time.Now().Format("2006-01-02 15:04:05")
	res, err := h.db.Exec(stmt, input.Name, input.Email, string(hash), role, createdAt)
	if err != nil {
		// Si el email existe, SQLite devuelve error UNIQUE
		if errors.Is(err, sql.ErrNoRows) || (err.Error() != "" && (contains(err.Error(), "UNIQUE") || contains(err.Error(), "unique"))) {
			http.Error(w, "El email ya está registrado", http.StatusConflict)
			return
		}
		http.Error(w, "Error al registrar usuario", http.StatusInternalServerError)
		return
	}

	id, _ := res.LastInsertId()

	// Crear respuesta sin password
	user := User{
		ID:        id,
		Name:      input.Name,
		Email:     input.Email,
		Role:      role,
		CreatedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// contains es para buscar si el error contiene "UNIQUE"
func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
