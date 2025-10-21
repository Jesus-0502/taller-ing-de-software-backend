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
    role TEXT NOT NULL DEFAULT 'user'
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
	