package handlers

import (
	"database/sql"
	"encoding/json"
	"farmlands-backend/models"
	"farmlands-backend/utils"
	"log"
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
		log.Println("Error decodificando JSON:", err)
		utils.SendJSONError(w, http.StatusBadRequest, "INVALID_JSON", "JSON inválido")
		return
	}

	if input.Descripcion == "" {
		utils.SendJSONError(w, http.StatusBadRequest, "INVALID_INPUT", "La descripción es obligatoria")
		return
	}

	layout := "2006-01-02"
	if _, err := time.Parse(layout, input.FechaInicio); err != nil {
		utils.SendJSONError(w, http.StatusBadRequest, "INVALID_DATE", "Formato de fecha inválido (YYYY-MM-DD)")
		return
	}

	stmt := `
		INSERT INTO projects (descripcion, fecha_inicio, fecha_cierre, estado, created_at)
		VALUES (?, ?, ?, 'abierto', ?)
	`
	createdAt := time.Now().Format(layout)

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
		CreatedAt:   createdAt,
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

		if err := rows.Scan(&p.ID, &p.Descripcion, &p.FechaInicio, &p.FechaCierre, &p.Estado, &p.CreatedAt); err != nil {
			utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error leyendo los datos")
			return
		}
		projects = append(projects, p)
	}

	if err := rows.Err(); err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error iterando resultados")
		return
	}

	utils.SendJSONSuccess(w, projects)
}

func (h *ProjectHandler) HandleProjectQuery(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q") // texto buscado

	// Si no hay query, devolvemos todos los proyectos
	var rows *sql.Rows
	var err error
	if query == "" {
		rows, err = h.DB.Query("SELECT id, descripcion, fecha_inicio, fecha_cierre, estado, created_at FROM projects")
	} else {
		rows, err = h.DB.Query(
			"SELECT id, descripcion, fecha_inicio, fecha_cierre, estado, created_at FROM projects WHERE UPPER(descripcion) LIKE UPPER(?)",
			"%"+query+"%",
		)
	}

	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error en la búsqueda")
		return
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(&p.ID, &p.Descripcion, &p.FechaInicio, &p.FechaCierre, &p.Estado, &p.CreatedAt); err != nil {
			utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error leyendo resultados")
			return
		}
		projects = append(projects, p)
	}

	utils.SendJSONSuccess(w, projects)
}
