package handlers

import (
	"database/sql"
	"encoding/json"
	"farmlands-backend/models"
	"farmlands-backend/utils"
	"log"
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

func (h *ToolsHandler) HandleAddTool(w http.ResponseWriter, r *http.Request) {
	var input models.AddTool
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println("Error decodificando JSON:", err)
		utils.SendJSONError(w, http.StatusBadRequest, "INVALID_JSON", "JSON inválido")
		return
	}

	if input.Descripcion == "" {
		utils.SendJSONError(w, http.StatusBadRequest, "INVALID_INPUT", "La descripción es obligatoria")
		return
	}

	stmt := `
		INSERT INTO tools (descripcion)
		VALUES (?)
	`

	res, err := h.DB.Exec(stmt, input.Descripcion)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error agregando nuevo implemento")
		return
	}

	id, _ := res.LastInsertId()

	tool := models.Tool{
		ID:          id,
		Descripcion: input.Descripcion,
	}

	utils.SendJSONSuccess(w, tool)
}

// func (h *ToolsHandler) HandleEditTool(w http.ResponseWriter, r *http.Request)

// func (h *ToolsHandler) HandleDeleteTool(w http.ResponseWriter, r *http.Request)

func (h *ToolsHandler) HandleSearchTool(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q") // texto buscado

	// Si no hay query, devolvemos todos los suplementos
	var rows *sql.Rows
	var err error
	if query == "" {
		rows, err = h.DB.Query("SELECT id, descripcion FROM tools")
	} else {
		rows, err = h.DB.Query(
			"SELECT id, descripcion FROM tools WHERE UPPER(descripcion) LIKE UPPER(?)",
			"%"+query+"%",
		)
	}

	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error en la búsqueda")
		return
	}
	defer rows.Close()

	var tools []models.Tool
	for rows.Next() {
		var t models.Tool
		if err := rows.Scan(&t.ID, &t.Descripcion); err != nil {
			utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error leyendo resultados")
			return
		}
		tools = append(tools, t)
	}

	utils.SendJSONSuccess(w, tools)
}
