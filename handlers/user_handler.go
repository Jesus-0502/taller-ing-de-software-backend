package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"farmlands-backend/middleware"
	"farmlands-backend/models"
	"farmlands-backend/utils"
	"io"
	"log"
	"net/http"
	"regexp"
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

func (h *UserHandler) HandleListUsers(w http.ResponseWriter, r *http.Request) {

	// Consulta a la base de datos
	rows, err := h.DB.Query("SELECT id, name, email, role, created_at FROM users")
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error interno del servidor")
		return
	}
	defer rows.Close()

	var users []models.User

	// Iterar sobre los resultados
	for rows.Next() {
		var u models.User
		var createdAtStr sql.NullString // temporal para leer la fecha como string
		layout := "2006-01-02 15:04:05"

		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &createdAtStr); err != nil {
			utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error leyendo los datos")
			return
		}

		// Parsear string a time.Time
		u.CreatedAt, _ = time.Parse(layout, createdAtStr.String)
		// if err != nil {
		// 	utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error parseando fecha")
		// 	return
		// }

		users = append(users, u)
	}

	// Verificar errores en la iteración
	if err := rows.Err(); err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error iterando resultados")
		return
	}

	if users == nil {
		users = []models.User{}
	}

	utils.SendJSONSuccess(w, users)
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
		utils.SendJSONError(w, http.StatusBadRequest, "INVALID_JSON", "JSON inválido")
		return
	}

	// === Validaciones básicas ===
	if input.Name == "" || input.Lastname == "" || input.Username == "" || input.Email == "" || len(input.Password) < 6 {
		utils.SendJSONError(w, http.StatusBadRequest, "INVALID_DATA", "Todos los campos son obligatorios y la contraseña debe tener al menos 6 caracteres")
		return
	}

	// === Validar formato de username ===
	validUsername := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !validUsername.MatchString(input.Username) {
		utils.SendJSONError(w, http.StatusBadRequest, "INVALID_USERNAME", "El nombre de usuario solo puede contener letras, números y guiones bajos")
		return
	}

	// === Validar formato de lastname ===
	validLastname := regexp.MustCompile(`^[a-zA-ZÀ-ÿ\s]+$`)
	if !validLastname.MatchString(input.Lastname) {
		utils.SendJSONError(w, http.StatusBadRequest, "INVALID_LASTNAME", "El apellido solo puede contener letras y espacios")
		return
	}

	// === Verificación de duplicados ===
	var userExists, emailExists bool

	// Verificar username
	err := h.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)`, input.Username).Scan(&userExists)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error verificando nombre de usuario")
		return
	}

	// Verificar email
	err = h.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)`, input.Email).Scan(&emailExists)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error verificando correo electrónico")
		return
	}

	// === Responder según el tipo de duplicado encontrado ===
	if userExists && emailExists {
		utils.SendJSONError(w, http.StatusConflict, "USER_AND_EMAIL_EXIST", "El nombre de usuario y el correo electrónico ya existen")
		return
	} else if userExists {
		utils.SendJSONError(w, http.StatusConflict, "USERNAME_EXISTS", "El nombre de usuario ya está en uso")
		return
	} else if emailExists {
		utils.SendJSONError(w, http.StatusConflict, "EMAIL_EXISTS", "El correo electrónico ya está registrado")
		return
	}

	// === Hashear contraseña ===
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "HASH_ERROR", "Error al hashear la contraseña")
		return
	}

	// === Insertar nuevo usuario ===
	stmt := `
		INSERT INTO users (name, lastname, username, email, password_hash, role, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	role := input.Role
	if role == "" {
		role = "user"
	}

	layout := "2006-01-02"
	createdAt := time.Now().Format(layout)

	res, err := h.DB.Exec(stmt, input.Name, input.Lastname, input.Username, input.Email, string(hash), role, createdAt)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error al registrar usuario")
		return
	}

	id, _ := res.LastInsertId()
	user := models.User{
		ID:        id,
		Name:      input.Name,
		Lastname:  input.Lastname,
		Username:  input.Username,
		Email:     input.Email,
		Role:      role,
		CreatedAt: time.Now(),
	}

	utils.SendJSONSuccess(w, user)
}

func (h *UserHandler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {

	bodyBytes, _ := io.ReadAll(r.Body)

	var input struct {
		ID int64 `json:"id"`
	}

	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.SendJSONError(w, http.StatusBadRequest, "INVALID_JSON", "JSON inválido")
		return
	}

	res, err := h.DB.Exec("DELETE FROM users WHERE id = ?", input.ID)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error eliminando proyecto")
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		utils.SendJSONError(w, http.StatusNotFound, "NOT_FOUND", "Usuario no encontrado")
		return
	}

	utils.SendJSONSuccess(w, "Usuario eliminado correctamente")
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
