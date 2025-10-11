package user

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// helper para construir router con una ruta distinta
func buildRouter(h *Handler) http.Handler {
	r := mux.NewRouter()
	sub := r.PathPrefix("/api/usuario").Subrouter()
	h.RegisterRoutes(sub)
	return r
}

/***************************************
*********** Register Test **************
****************************************/

// Caso 1: Login exitoso
func TestHandleRegister_Success(t *testing.T) {

	// Se crea la base de datos simulada
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// Se inyecta la base de datos y se registran las rutas
	h := NewHandler(db)
	router := buildRouter(h)

	// Se crea el JSON que simula el cuerpo del POST.
	// Este cuerpo se envia como si fuera un usuario registrandose.
	body := map[string]string{
		"name":     "Juan",
		"email":    "juan@example.com",
		"password": "secreto123",
	}
	b, _ := json.Marshal(body)

	// Se espera que se ejecute un comando SQL (Exec) que contenga la cadena "INSERT INTO users".
	mock.ExpectExec("INSERT INTO users").
		WithArgs("Juan", "juan@example.com", sqlmock.AnyArg(), "user", sqlmock.AnyArg()). // Define los valores esperados que el codigo deberia pasar al Exec:
		WillReturnResult(sqlmock.NewResult(1, 1))                                         // Simula la respuesta que la DB devolveria después del INSERT.

	// Se crea una peticion HTTP falsa hacia el router.
	req := httptest.NewRequest(http.MethodPost, "/api/usuario/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	// Esto actua como el response writer, donde se guardara la respuesta.
	rec := httptest.NewRecorder()

	// Se ejecuta el endpoint real (handleRegister).
	router.ServeHTTP(rec, req)

	// Se comprueba que el código de respuesta HTTP sea 201 Created
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Caso 2: Email ya existente
func TestHandleRegister_DuplicateEmail(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	h := NewHandler(db)
	router := buildRouter(h)

	body := map[string]string{
		"name":     "Juan",
		"email":    "juan@example.com",
		"password": "secreto123",
	}
	b, _ := json.Marshal(body)

	// Simula violacion UNIQUE
	mock.ExpectExec("INSERT INTO users").
		WithArgs("Juan", "juan@example.com", sqlmock.AnyArg(), "user", sqlmock.AnyArg()).
		WillReturnError(&sqliteUniqueErr{msg: "UNIQUE constraint failed: users.email"})

	req := httptest.NewRequest(http.MethodPost, "/api/usuario/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Verificamos que respondio con error 409
	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Caso 3: Campos vacios
func TestHandleRegister_BadRequest_WhenMissingFields(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	h := NewHandler(db)
	router := buildRouter(h)

	// Cuerpo del JSON con datos faltantes
	body := map[string]string{
		"email": "sin-nombre@example.com",
		// falta name y password
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/usuario/register", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Verifica si se recibe el error 400
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// No debe ejecutar ningun INSERT
	assert.NoError(t, mock.ExpectationsWereMet())
}

/***************************************
*********** Login Test *****************
****************************************/

// Caso 1: Login exitoso
func TestHandleLogin_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	h := NewHandler(db)
	router := buildRouter(h)

	// Generamos el hash de la clave que vamos a ingresar para la prueba
	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte("secreto123"), bcrypt.DefaultCost)
	createdAt := "2025-10-11 10:00:00"

	mock.ExpectQuery("SELECT id, name, email, password_hash, role, created_at FROM users WHERE email = ?").
		WithArgs("juan@example.com").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "email", "password_hash", "role", "created_at",
		}).AddRow(1, "Juan", "juan@example.com", string(hashedPwd), "user", createdAt))

	mock.ExpectExec("INSERT INTO bearer_tokens").
		WithArgs(sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	body := map[string]string{
		"email":    "juan@example.com",
		"password": "secreto123",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/usuario/login", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"token"`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Caso 2: Email no encontrado
func TestHandleLogin_EmailNotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	h := NewHandler(db)
	router := buildRouter(h)

	mock.ExpectQuery("SELECT id, name, email, password_hash, role, created_at FROM users WHERE email = ?").
		WithArgs("noexiste@example.com").
		WillReturnError(sql.ErrNoRows)

	body := map[string]string{
		"email":    "noexiste@example.com",
		"password": "loquesea",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/usuario/login", bytes.NewReader(b))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Caso 3: Contraseña incorrecta
func TestHandleLogin_WrongPassword(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	h := NewHandler(db)
	router := buildRouter(h)

	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte("secreto123"), bcrypt.DefaultCost)
	createdAt := time.Now().Format("2006-01-02 15:04:05")

	mock.ExpectQuery("SELECT id, name, email, password_hash, role, created_at FROM users WHERE email = ?").
		WithArgs("juan@example.com").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "email", "password_hash", "role", "created_at",
		}).AddRow(1, "Juan", "juan@example.com", string(hashedPwd), "user", createdAt))

	body := map[string]string{
		"email":    "juan@example.com",
		"password": "mala123",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/usuario/login", bytes.NewReader(b))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Caso 4: Campos vacios
func TestHandleLogin_EmptyFields(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	h := NewHandler(db)
	router := buildRouter(h)

	body := map[string]string{
		"email":    "",
		"password": "",
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/usuario/login", bytes.NewReader(b))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// sqliteUniqueErr es un tipo para simular un error UNIQUE del driver
type sqliteUniqueErr struct{ msg string }

func (e *sqliteUniqueErr) Error() string { return e.msg }
