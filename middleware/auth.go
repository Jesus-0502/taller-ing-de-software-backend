package middleware

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("tu_clave_super_secreta") // cámbiala por una clave fuerte

// Generar token JWT
func GenerateJWT(userID int64, email string, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // expira en 24 horas
		"iat":     time.Now().Unix(),                     // emitido ahora
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
