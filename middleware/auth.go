package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Usa la misma clave secreta que usas para firmar los tokens
var jwtSecret = []byte("mi_super_clave_secreta") // ⚠️ cámbiala por algo fuerte

// VerifyToken es un middleware que valida el token JWT enviado por el cliente
func VerifyToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1️⃣ Obtenemos el header Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Falta el token", http.StatusUnauthorized)
			return
		}

		// 2️⃣ Extraemos el token después de "Bearer "
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		tokenString = strings.TrimSpace(tokenString)

		if tokenString == "" {
			http.Error(w, "Token inválido", http.StatusUnauthorized)
			return
		}

		// 3️⃣ Parseamos y verificamos el token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validar el método de firma
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de firma inválido")
			}
			return jwtSecret, nil
		})

		if err != nil {
			http.Error(w, "Error al procesar token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			http.Error(w, "Token inválido o expirado", http.StatusUnauthorized)
			return
		}

		// 4️⃣ Si el token es válido, pasamos al siguiente handler
		next.ServeHTTP(w, r)
	})
}
