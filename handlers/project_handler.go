package handlers

import (
	"database/sql"
	"encoding/json"
	"farmlands-backend/models"
	"farmlands-backend/utils"
	"net/http"
	"time"
)

type ProjectHandler struct {
	DB *sql.DB
}

func NewProjectHandler(db *sql.DB) *ProjectHandler {
	return &ProjectHandler{DB: db}
}

func (h *ProjectHandler) HandleCreateProject(w http.ResponseWriter, r *http.Request) {
	var input models.CreateProjectInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.SendJSONError(w, http.StatusBadRequest, "INVALID_JSON", "JSON inválido")
		return
	}

	if input.Descripcion == "" {
		utils.SendJSONError(w, http.StatusBadRequest, "INVALID_INPUT", "La descripción es obligatoria")
		return
	}

	stmt := `
		INSERT INTO projects (descripcion, fecha_inicio, fecha_cierre, estado, created_at)
		VALUES (?, ?, ?, 'abierto', ?)
	`
	createdAt := time.Now().Format("2006-01-02 15:04:05")

	res, err := h.DB.Exec(stmt, input.Descripcion, input.FechaInicio, input.FechaCierre, createdAt)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error al crear el proyecto")
		return
	}

	id, _ := res.LastInsertId()

	project := models.Project{
		ID:          id,
		Descripcion: input.Descripcion,
		FechaInicio: input.FechaInicio,
		FechaCierre: input.FechaCierre,
		Estado:      "abierto",
		CreatedAt:   time.Now(),
	}

	utils.SendJSONSuccess(w, project)
}

func (h *ProjectHandler) HandleListProjects(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`
		SELECT id, descripcion, fecha_inicio, fecha_cierre, estado, created_at
		FROM projects
	`)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error al obtener proyectos")
		return
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		var fechaInicio, fechaCierre, createdAt sql.NullString
		layout := "2006-01-02 15:04:05"

		if err := rows.Scan(&p.ID, &p.Descripcion, &fechaInicio, &fechaCierre, &p.Estado, &createdAt); err != nil {
			utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error leyendo los datos")
			return
		}

		// Convertir sql.NullString a string
		p.FechaInicio, _ = time.Parse(layout, fechaInicio.String)
		p.FechaCierre, _ = time.Parse(layout, fechaCierre.String)
		p.CreatedAt, _ = time.Parse(layout, createdAt.String)

		projects = append(projects, p)
	}

	if err := rows.Err(); err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error iterando resultados")
		return
	}

	utils.SendJSONSuccess(w, projects)
}
