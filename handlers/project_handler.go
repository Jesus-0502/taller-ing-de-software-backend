package handlers

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"farmlands-backend/models"
	"farmlands-backend/utils"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
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

	if projects == nil {
		projects = []models.Project{}
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

	if projects == nil {
		projects = []models.Project{}
	}

	utils.SendJSONSuccess(w, projects)
}

func (h *ProjectHandler) HandleDeleteProject(w http.ResponseWriter, r *http.Request) {

	bodyBytes, _ := io.ReadAll(r.Body)

	var input struct {
		ID int64 `json:"id"`
	}

	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.SendJSONError(w, http.StatusBadRequest, "INVALID_JSON", "JSON inválido")
		return
	}

	res, err := h.DB.Exec("DELETE FROM projects WHERE id = ?", input.ID)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error eliminando proyecto")
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		utils.SendJSONError(w, http.StatusNotFound, "NOT_FOUND", "Proyecto no encontrado")
		return
	}

	utils.SendJSONSuccess(w, map[string]string{"message": "Proyecto eliminado correctamente"})
}

func (h *ProjectHandler) HandleUpdateProject(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ID          int64   `json:"id"`
		Descripcion *string `json:"descripcion,omitempty"`
		FechaInicio *string `json:"fecha_inicio,omitempty"`
		FechaCierre *string `json:"fecha_cierre,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.SendJSONError(w, http.StatusBadRequest, "INVALID_JSON", "JSON inválido o mal formado")
		return
	}
	defer r.Body.Close()

	if input.ID == 0 {
		utils.SendJSONError(w, http.StatusBadRequest, "MISSING_ID", "El campo 'id' es obligatorio")
		return
	}

	// Construir la query dinámicamente
	query := "UPDATE projects SET "
	args := []interface{}{}
	updates := []string{}

	if input.Descripcion != nil {
		updates = append(updates, "descripcion = ?")
		args = append(args, *input.Descripcion)
	}
	if input.FechaInicio != nil {
		updates = append(updates, "fecha_inicio = ?")
		args = append(args, *input.FechaInicio)
	}
	if input.FechaCierre != nil {
		updates = append(updates, "fecha_cierre = ?")
		args = append(args, *input.FechaCierre)
	}

	if len(updates) == 0 {
		utils.SendJSONError(w, http.StatusBadRequest, "NO_FIELDS", "No se proporcionaron campos para actualizar")
		return
	}

	query += strings.Join(updates, ", ") + " WHERE id = ?"
	args = append(args, input.ID)

	_, err := h.DB.Exec(query, args...)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error al actualizar el proyecto")
		return
	}

	utils.SendJSONSuccess(w, map[string]string{
		"message": "Proyecto actualizado correctamente",
	})
}

func (h *ProjectHandler) HandleUpdateProjectStatus(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ID     int64  `json:"id"`
		Estado string `json:"estado"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.SendJSONError(w, http.StatusBadRequest, "INVALID_JSON", "JSON inválido o mal formado")
		return
	}
	defer r.Body.Close()

	if input.ID == 0 || input.Estado == "" {
		utils.SendJSONError(w, http.StatusBadRequest, "MISSING_FIELDS", "El 'id' y 'estado' son obligatorios")
		return
	}

	// Validar que el estado sea uno permitido
	validStates := map[string]bool{
		"abierto":  true,
		"cerrado":  true,
		"en pausa": true,
	}
	if !validStates[input.Estado] {
		utils.SendJSONError(w, http.StatusBadRequest, "INVALID_STATE", "Estado no válido")
		return
	}

	// Actualizar en la base de datos
	stmt := `UPDATE projects SET estado = ? WHERE id = ?`
	_, err := h.DB.Exec(stmt, input.Estado, input.ID)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error al actualizar el estado del proyecto")
		return
	}

	utils.SendJSONSuccess(w, map[string]string{
		"message": fmt.Sprintf("Estado del proyecto #%d actualizado a '%s'", input.ID, input.Estado),
	})
}

func (h *ProjectHandler) HandleDownloadProjects(w http.ResponseWriter, r *http.Request) {
	// Consultar todos los proyectos
	rows, err := h.DB.Query(`SELECT id, descripcion, fecha_inicio, fecha_cierre, estado, created_at FROM projects`)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error al obtener los proyectos")
		return
	}
	defer rows.Close()

	// Crear un buffer para escribir el CSV
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment;filename=proyectos.csv")

	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// Escribir encabezados del CSV
	headers := []string{"ID", "Descripcion", "FechaInicio", "FechaCierre", "Estado", "CreatedAt"}
	if err := csvWriter.Write(headers); err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "CSV_ERROR", "Error escribiendo encabezado CSV")
		return
	}

	// Escribir filas con los datos de la base
	for rows.Next() {
		var id int
		var descripcion, fechaInicio, fechaCierre, estado, createdAt string

		if err := rows.Scan(&id, &descripcion, &fechaInicio, &fechaCierre, &estado, &createdAt); err != nil {
			continue
		}

		record := []string{
			fmt.Sprintf("%d", id),
			descripcion,
			fechaInicio,
			fechaCierre,
			estado,
			createdAt,
		}
		if err := csvWriter.Write(record); err != nil {
			continue
		}
	}
}

func (h *ProjectHandler) HandleGetUserProjects(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID int `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.SendJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Formato JSON inválido")
		return
	}

	if input.UserID == 0 {
		utils.SendJSONError(w, http.StatusBadRequest, "MISSING_FIELDS", "Falta el campo user_id")
		return
	}

	query := `
		SELECT 
			p.id, 
			p.descripcion
		FROM projects p
		INNER JOIN user_projects up ON p.id = up.project_id
		WHERE up.user_id = ?
	`

	rows, err := h.DB.Query(query, input.UserID)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error al obtener los proyectos del usuario")
		return
	}
	defer rows.Close()

	var proyectos []models.Project

	for rows.Next() {
		var p models.Project
		if err := rows.Scan(&p.ID, &p.Descripcion); err != nil {
			continue
		}

		proyectos = append(proyectos, p)
	}

	if proyectos == nil {
		proyectos = []models.Project{}
	}

	utils.SendJSONSuccess(w, proyectos)
}

func (h *ProjectHandler) HandleGetAvailableProjectsForUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID int `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.SendJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Formato JSON inválido")
		return
	}

	if input.UserID == 0 {
		utils.SendJSONError(w, http.StatusBadRequest, "MISSING_FIELDS", "Falta el campo user_id")
		return
	}

	query := `
		SELECT 
			p.id, 
			p.descripcion
		FROM projects p
		WHERE p.id NOT IN (
			SELECT project_id FROM user_projects WHERE user_id = ?
		)
	`

	rows, err := h.DB.Query(query, input.UserID)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error al obtener proyectos disponibles")
		return
	}
	defer rows.Close()

	var proyectos []models.ProjectIdentification

	for rows.Next() {
		var p models.ProjectIdentification
		if err := rows.Scan(&p.ID, &p.Descripcion); err != nil {
			continue
		}

		proyectos = append(proyectos, p)
	}

	if proyectos == nil {
		proyectos = []models.ProjectIdentification{}
	}

	utils.SendJSONSuccess(w, proyectos)
}
func (h *ProjectHandler) HandleAssignUserToProject(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID    int    `json:"user_id"`
		ProjectID int    `json:"project_id"`
		Role      string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.SendJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Formato JSON inválido")
		return
	}

	// 🔹 Validar campos requeridos
	if input.UserID == 0 || input.ProjectID == 0 {
		utils.SendJSONError(w, http.StatusBadRequest, "MISSING_FIELDS", "Faltan campos obligatorios")
		return
	}

	if input.Role == "" {
		input.Role = "colaborador"
	}

	// 🔹 Validar existencia del usuario
	var exists int
	err := h.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE id = ?`, input.UserID).Scan(&exists)
	if err != nil || exists == 0 {
		utils.SendJSONError(w, http.StatusNotFound, "USER_NOT_FOUND", "El usuario no existe")
		return
	}

	// 🔹 Validar existencia del proyecto
	err = h.DB.QueryRow(`SELECT COUNT(*) FROM projects WHERE id = ?`, input.ProjectID).Scan(&exists)
	if err != nil || exists == 0 {
		utils.SendJSONError(w, http.StatusNotFound, "PROJECT_NOT_FOUND", "El proyecto no existe")
		return
	}

	// 🔹 Verificar si ya está asociado
	err = h.DB.QueryRow(`
		SELECT COUNT(*) FROM user_projects 
		WHERE user_id = ? AND project_id = ?
	`, input.UserID, input.ProjectID).Scan(&exists)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error consultando asociaciones existentes")
		return
	}
	if exists > 0 {
		utils.SendJSONError(w, http.StatusConflict, "ALREADY_ASSIGNED", "El usuario ya está asignado a este proyecto")
		return
	}

	// 🔹 Insertar nueva relación
	stmt := `
		INSERT INTO user_projects (user_id, project_id, role_in_project, assigned_at)
		VALUES (?, ?, ?, datetime('now'))
	`
	_, err = h.DB.Exec(stmt, input.UserID, input.ProjectID, input.Role)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_INSERT_ERROR", "Error al asignar el usuario al proyecto")
		return
	}

	utils.SendJSONSuccess(w, map[string]string{
		"message": "Usuario asignado correctamente al proyecto",
	})
}

// HandleChangeAssociation permite cambiar el proyecto al que está asociado un usuario
func (h *ProjectHandler) HandleChangeAssociation(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID       int64 `json:"user_id"`
		OldProjectID int64 `json:"old_project_id"`
		NewProjectID int64 `json:"new_project_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.SendJSONError(w, http.StatusBadRequest, "INVALID_JSON", "JSON inválido")
		return
	}

	if input.UserID == 0 || input.OldProjectID == 0 || input.NewProjectID == 0 {
		utils.SendJSONError(w, http.StatusBadRequest, "MISSING_FIELDS", "Faltan campos obligatorios")
		return
	}

	// Verificar que el nuevo proyecto existe
	var exists int
	err := h.DB.QueryRow(`SELECT COUNT(*) FROM projects WHERE id = ?`, input.NewProjectID).Scan(&exists)
	if err != nil || exists == 0 {
		utils.SendJSONError(w, http.StatusNotFound, "PROJECT_NOT_FOUND", "El nuevo proyecto no existe")
		return
	}

	// Actualizar asociación del usuario
	stmt := `
		UPDATE user_projects
		SET project_id = ?, assigned_at = DATETIME('now')
		WHERE user_id = ? AND project_id = ?
	`
	res, err := h.DB.Exec(stmt, input.NewProjectID, input.UserID, input.OldProjectID)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error al cambiar asociación")
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		utils.SendJSONError(w, http.StatusNotFound, "ASSOCIATION_NOT_FOUND", "No se encontró la asociación especificada")
		return
	}

	utils.SendJSONSuccess(w, map[string]string{"message": "Asociación cambiada correctamente"})
}

// HandleRemoveAssociation elimina la relación entre un usuario y un proyecto
func (h *ProjectHandler) HandleRemoveAssociation(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID    int64 `json:"user_id"`
		ProjectID int64 `json:"project_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.SendJSONError(w, http.StatusBadRequest, "INVALID_JSON", "JSON inválido")
		return
	}

	if input.UserID == 0 || input.ProjectID == 0 {
		utils.SendJSONError(w, http.StatusBadRequest, "MISSING_FIELDS", "Faltan campos obligatorios")
		return
	}

	stmt := `DELETE FROM user_projects WHERE user_id = ? AND project_id = ?`
	res, err := h.DB.Exec(stmt, input.UserID, input.ProjectID)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error al eliminar asociación")
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		utils.SendJSONError(w, http.StatusNotFound, "ASSOCIATION_NOT_FOUND", "No se encontró la asociación")
		return
	}

	utils.SendJSONSuccess(w, map[string]string{"message": "Asociación eliminada correctamente"})
}
