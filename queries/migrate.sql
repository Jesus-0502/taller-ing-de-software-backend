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

INSERT INTO users (name, email, password_hash, role)
VALUES ('root', 'root@example.com', '$2a$10$.e2jTOtVHftDwmE5N2ig2eCvkMKzF3Y8UZu3Qg9t4NwzwLUlrh.Ou', 'admin');