package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// NewDB crea la conexion a la base de datos y realiza la migracion basica (creacion de tabla users)
func NewDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("error abriendo la base de datos: %w", err)
	}

	// limita conexiones simultáneas (SQLite no es para cargas altas)
	// db.SetMaxOpenConns(1)
	// db.SetMaxIdleConns(1)

	// Validacion de la conexión
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error conectando a la base: %w", err)
	}

	// Migración básica: crea la tabla users si no existe
	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("error en migración: %w", err)
	}

	log.Println("Conexion a la base de datos SQLite establecida y migracion OK")
	return db, nil
}

// migrate crea la tabla users si no existe
func migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

	CREATE TABLE IF NOT EXISTS bearer_tokens (
		id INTEGER PRIMARY KEY AUTOINCREMENT,       -- Identificador único del token
		token TEXT NOT NULL UNIQUE,                 -- El token (cadena aleatoria o JWT)
		user_id INTEGER NOT NULL,                   -- Llave foránea a la tabla users
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Fecha en que se generó el token
		expires_at DATETIME NOT NULL DEFAULT (DATETIME(CURRENT_TIMESTAMP, '+1 day')), 
		-- Fecha de expiración: 24 horas después de crearse

		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		-- Si el usuario se elimina, sus tokens también
	);
	`
	_, err := db.Exec(schema)
	return err
}

// INSERT INTO users (name, email, password_hash, role, created_at)
// VALUES ('Alan', 'alan@example.com', '$2a$10$.e2jTOtVHftDwmE5N2ig2eCvkMKzF3Y8UZu3Qg9t4NwzwLUlrh.Ou', 'admin', '2025-10-08 03:47:28');
