package handlers

import (
	"database/sql"
	"farmlands-backend/models"
	"farmlands-backend/utils"
	"net/http"
)

type ToolsHandler struct {
	DB *sql.DB
}

func NewToolsHandler(db *sql.DB) *ToolsHandler {
	return &ToolsHandler{DB: db}
}

func (h *ToolsHandler) HandleListTools(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`
		SELECT id, descripcion FROM tools
	`)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error al obtener equipos e implementos")
		return
	}
	defer rows.Close()

	var tools []models.Tool
	for rows.Next() {
		var p models.Tool

		if err := rows.Scan(&p.ID, &p.Descripcion); err != nil {
			utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error leyendo los datos")
			return
		}
		tools = append(tools, p)
	}

	if err := rows.Err(); err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error iterando resultados")
		return
	}

	if tools == nil {
		tools = []models.Tool{}
	}
	utils.SendJSONSuccess(w, tools)
}

// func (h *ToolsHandler) HandleAddTool(w http.ResponseWriter, r *http.Request)

// func (h *ToolsHandler) HandleEditTool(w http.ResponseWriter, r *http.Request)

// func (h *ToolsHandler) HandleDeleteTool(w http.ResponseWriter, r *http.Request)

// func (h *ToolsHandler) HandleSearchTool(w http.ResponseWriter, r *http.Request)
