package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// NewDB crea la conexión a la base de datos y ejecuta la migración
// solo si el archivo no existe aún.
func NewDB(dataSourceName string) (*sql.DB, error) {
	// Verificar si ya existe el archivo de base de datos
	dbExists := fileExists(dataSourceName)

	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("error abriendo la base de datos: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error conectando a la base de datos: %w", err)
	}

	// Solo ejecuta la migración si el archivo no existía
	if !dbExists {
		log.Println("Base de datos no encontrada. Ejecutando migración inicial...")
		if err := migrate(db); err != nil {
			return nil, fmt.Errorf("error en migración: %w", err)
		}
		log.Println("Migración completada exitosamente")
	} else {
		log.Println("Base de datos existente detectada. No se ejecuta migración.")
	}

	log.Println("Conexión a SQLite establecida.")
	return db, nil
}

// fileExists verifica si un archivo existe en el sistema
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// migrate crea las tablas iniciales y el usuario admin
func migrate(db *sql.DB) error {
	schema := `
-- =============================================
-- TABLA DE ADMINISTRADORES
-- =============================================

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    lastname TEXT NOT NULL,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
);

-- =============================================
-- TABLA DE BEARER TOKENS
-- =============================================

CREATE TABLE IF NOT EXISTS bearer_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    token TEXT NOT NULL UNIQUE,
    user_id INTEGER NOT NULL,
    created_at TEXT,
    expires_at TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- =============================================
-- TABLA DE PROYECTOS AGRÍCOLAS
-- =============================================
CREATE TABLE IF NOT EXISTS projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    descripcion TEXT NOT NULL,
    fecha_inicio TEXT NOT NULL,
    fecha_cierre TEXT NOT NULL,
    estado TEXT NOT NULL DEFAULT 'abierto',
    created_at TEXT
);

-- =============================================
-- RELACIÓN USUARIOS - PROYECTOS
-- =============================================
CREATE TABLE IF NOT EXISTS user_projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    project_id INTEGER NOT NULL,
    role_in_project TEXT DEFAULT 'colaborador',
    assigned_at TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    UNIQUE(user_id, project_id)
);

-- =============================================
-- USUARIO ADMINISTRADOR INICIAL
-- =============================================
INSERT INTO users (name, lastname, username, email, password_hash, role)
VALUES ('root', 'root', 'root','root@example.com', '$2a$10$.e2jTOtVHftDwmE5N2ig2eCvkMKzF3Y8UZu3Qg9t4NwzwLUlrh.Ou', 'admin');
	`
	_, err := db.Exec(schema)
	return err
}
