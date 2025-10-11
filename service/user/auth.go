package user

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("tu_clave_super_secreta") // cámbiala por una clave fuerte

// Estructura del payload
type Claims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

// Generar token JWT
func GenerateJWT(userID int64, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID, // ahora almacena int64 directamente
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
