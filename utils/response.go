package utils

import (
	"encoding/json"
	"net/http"
)

// JSONError estructura estándar para errores del backend
type JSONError struct {
	Code    string `json:"code"`    // código interno (ej: "USER_NOT_FOUND")
	Message string `json:"message"` // descripción legible para el usuario
}

// JSONResponse estructura genérica para cualquier respuesta
type JSONResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`  // omite si es nil
	Error   *JSONError  `json:"error,omitempty"` // omite si no hay error
}

// SendJSONError envía un error en formato JSON uniforme
func SendJSONError(w http.ResponseWriter, code int, errCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(JSONResponse{
		Success: false,
		Error: &JSONError{
			Code:    errCode,
			Message: message,
		},
	})
}

// SendJSONSuccess envía una respuesta exitosa
func SendJSONSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(JSONResponse{
		Success: true,
		Data:    data,
	})
}
