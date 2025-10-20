package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"farmlands-backend/middleware"
	"farmlands-backend/models"
	"farmlands-backend/utils"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	DB *sql.DB
}

func NewUserHandler(db *sql.DB) *UserHandler {
	return &UserHandler{DB: db}
}

func (h *UserHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var input models.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		// http.Error(w, "JSON inválido", http.StatusBadRequest)
		utils.SendJSONError(w, http.StatusBadRequest, "INVALID_JSON", "JSON inválido")
		return
	}

	if input.Email == "" || input.Password == "" {
		// http.Error(w, "Email y contraseña requeridos", http.StatusBadRequest)
		utils.SendJSONError(w, http.StatusBadRequest, "MISSING_FIELDS", "Email y contraseña requeridos")
		return
	}

	stmt := `SELECT id, name, email, password_hash, role, created_at FROM users WHERE email = ? LIMIT 1`
	var u models.User
	var createdAtStr string

	email := strings.ToLower(strings.TrimSpace(input.Email))
	log.Println("Email recibido:", input.Email, "| Email limpio:", email, ".")

	err := h.DB.QueryRow(stmt, email).Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &createdAtStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// http.Error(w, "Credenciales incorrectas", http.StatusUnauthorized)
			utils.SendJSONError(w, http.StatusUnauthorized, "USER_NOT_FOUND", "Credenciales incorrectas")
			return
		}
		// http.Error(w, "Error interno", http.StatusInternalServerError)
		// utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error interno del servidor")
		// return
	}

	u.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(input.Password)); err != nil {
		// http.Error(w, "Error: Credenciales incorrectas", http.StatusUnauthorized)
		utils.SendJSONError(w, http.StatusUnauthorized, "INVALID_PASSWORD", "Credenciales incorrectas")
		return
	}

	token, err := middleware.GenerateJWT(u.ID, u.Email, u.Role)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "TOKEN_ERROR", "Error generando token")
		return
	}

	resp := struct {
		Token string      `json:"token"`
		User  models.User `json:"user"`
	}{
		Token: token,
		User:  u,
	}

	utils.SendJSONSuccess(w, resp)
}

func (h *UserHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var input models.CreateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if input.Name == "" || input.Email == "" || len(input.Password) < 6 {
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error al hashear contraseña", http.StatusInternalServerError)
		return
	}

	stmt := `
	INSERT INTO users (name, email, password_hash, role, created_at)
	VALUES (?, ?, ?, ?, ?)
	`
	role := input.Role
	if role == "" {
		role = "user"
	}

	createdAt := time.Now().Format("2006-01-02 15:04:05")
	res, err := h.DB.Exec(stmt, input.Name, input.Email, string(hash), role, createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || (err.Error() != "" && (contains(err.Error(), "UNIQUE") || contains(err.Error(), "unique"))) {
			http.Error(w, "El email ya está registrado", http.StatusConflict)
			return
		}
		http.Error(w, "Error al registrar usuario", http.StatusInternalServerError)
		return
	}

	id, _ := res.LastInsertId()
	user := models.User{
		ID:        id,
		Name:      input.Name,
		Email:     input.Email,
		Role:      role,
		CreatedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
