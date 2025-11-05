package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"farmlands-backend/middleware"
	"farmlands-backend/models"
	"farmlands-backend/utils"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

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
	rows, err := h.DB.Query("SELECT id, name, lastname, username, email, role FROM users")
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error interno del servidor")
		return
	}
	defer rows.Close()

	var users []models.User

	// Iterar sobre los resultados
	for rows.Next() {
		var u models.User

		if err := rows.Scan(&u.ID, &u.Name, &u.Lastname, &u.Username, &u.Email, &u.Role); err != nil {
			utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error leyendo los datos")
			return
		}

		// Parsear string a time.Time
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
	stmt := `SELECT id, name, lastname, username, email, password_hash, role FROM users WHERE email = ? LIMIT 1`
	var u models.User

	email := strings.ToLower(strings.TrimSpace(input.Email))

	err := h.DB.QueryRow(stmt, email).Scan(
		&u.ID, &u.Name, &u.Lastname, &u.Username, &u.Email, &u.PasswordHash, &u.Role,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// http.Error(w, "Credenciales incorrectas", http.StatusUnauthorized)
			utils.SendJSONError(w, http.StatusUnauthorized, "USER_NOT_FOUND", "Usuario no encontrado")
			return
		}
		// http.Error(w, "Error interno", http.StatusInternalServerError)
		// utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error interno del servidor")
		// return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(input.Password)); err != nil {
		// http.Error(w, "Error: Credenciales incorrectas", http.StatusUnauthorized)
		utils.SendJSONError(w, http.StatusUnauthorized, "INVALID_PASSWORD", "Contraseña incorrecta")
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
		INSERT INTO users (name, lastname, username, email, password_hash, role)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	role := input.Role
	if role == "" {
		role = "user"
	}

	res, err := h.DB.Exec(stmt, input.Name, input.Lastname, input.Username, input.Email, string(hash), role)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error al registrar usuario")
		return
	}

	id, _ := res.LastInsertId()
	user := models.User{
		ID:       id,
		Name:     input.Name,
		Lastname: input.Lastname,
		Username: input.Username,
		Email:    input.Email,
		Role:     role,
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

func (h *UserHandler) HandleUserQuery(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q") // texto buscado

	var rows *sql.Rows
	var err error

	if query == "" {
		rows, err = h.DB.Query(`
			SELECT id, name, lastname, username, email, role, password_hash
			FROM users
		`)
	} else {
		pattern := "%" + query + "%"
		rows, err = h.DB.Query(`
			SELECT id, name, lastname, username, email, role, password_hash
			FROM users
			WHERE UPPER(name) LIKE UPPER(?)
			   OR UPPER(lastname) LIKE UPPER(?)
			   OR UPPER(username) LIKE UPPER(?)
		`, pattern, pattern, pattern)
	}

	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error en la búsqueda")
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Lastname, &u.Username, &u.Email, &u.Role, &u.PasswordHash); err != nil {
			utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error leyendo resultados")
			return
		}
		users = append(users, u)
	}

	if users == nil {
		users = []models.User{}
	}

	utils.SendJSONSuccess(w, users)
}

// HandleListRoles obtiene todos los roles registrados
func (h *UserHandler) HandleListRoles(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query("SELECT id, role FROM roles")
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error al consultar los roles")
		return
	}
	defer rows.Close()

	roles := make(map[int]string)

	for rows.Next() {
		var id int
		var role string
		if err := rows.Scan(&id, &role); err != nil {
			utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error leyendo los datos de roles")
			return
		}
		roles[id] = role
	}

	utils.SendJSONSuccess(w, roles)
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func (h *UserHandler) HandleEditUser(w http.ResponseWriter, r *http.Request) {
	var input models.User

	// Decodificar JSON de entrada
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.SendJSONError(w, http.StatusBadRequest, "INVALID_JSON", "JSON inválido")
		return
	}

	// Validar que venga el ID
	if input.ID == 0 {
		utils.SendJSONError(w, http.StatusBadRequest, "MISSING_ID", "El ID del usuario es obligatorio")
		return
	}

	// Buscar usuario actual en BD
	var dbUser models.User
	err := h.DB.QueryRow(`
		SELECT id, name, lastname, username, email, password_hash, role
		FROM users WHERE id = ?`, input.ID,
	).Scan(
		&dbUser.ID, &dbUser.Name, &dbUser.Lastname, &dbUser.Username,
		&dbUser.Email, &dbUser.PasswordHash, &dbUser.Role,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.SendJSONError(w, http.StatusNotFound, "USER_NOT_FOUND", "Usuario no encontrado")
			return
		}
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error consultando usuario")
		return
	}

	// Comparar y construir dinámicamente la sentencia UPDATE
	updateFields := []string{}
	args := []interface{}{}

	if input.Name != "" && input.Name != dbUser.Name {
		updateFields = append(updateFields, "name = ?")
		args = append(args, input.Name)
	}
	if input.Lastname != "" && input.Lastname != dbUser.Lastname {
		updateFields = append(updateFields, "lastname = ?")
		args = append(args, input.Lastname)
	}
	if input.Username != "" && input.Username != dbUser.Username {
		updateFields = append(updateFields, "username = ?")
		args = append(args, input.Username)
	}
	if input.Email != "" && input.Email != dbUser.Email {
		updateFields = append(updateFields, "email = ?")
		args = append(args, input.Email)
	}
	if input.Role != "" && input.Role != dbUser.Role {
		updateFields = append(updateFields, "role = ?")
		args = append(args, input.Role)
	}

	// Si no hay cambios, devolver mensaje
	if len(updateFields) == 0 {
		utils.SendJSONSuccess(w, map[string]string{"message": "No se realizaron cambios"})
		return
	}

	// Ejecutar actualización dinámica
	query := fmt.Sprintf("UPDATE users SET %s WHERE id = ?", strings.Join(updateFields, ", "))
	args = append(args, input.ID)

	_, err = h.DB.Exec(query, args...)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_UPDATE_ERROR", "Error actualizando usuario")
		return
	}

	// Devolver usuario actualizado (merge entre original y cambios)
	if input.Name != "" {
		dbUser.Name = input.Name
	}
	if input.Lastname != "" {
		dbUser.Lastname = input.Lastname
	}
	if input.Username != "" {
		dbUser.Username = input.Username
	}
	if input.Email != "" {
		dbUser.Email = input.Email
	}
	if input.Role != "" {
		dbUser.Role = input.Role
	}

	utils.SendJSONSuccess(w, dbUser)
}
