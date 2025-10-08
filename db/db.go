package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// NewDB crea la conexion a la base de datos y realiza la migracion basica (creación de tabla users)
func NewDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("error abriendo la base de datos: %w", err)
	}

	// Recomendado: limita conexiones simultáneas (SQLite no es para cargas altas)
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
`
	_, err := db.Exec(schema)
	return err
}
