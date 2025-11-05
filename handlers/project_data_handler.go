package handlers

import (
	"database/sql"
	"encoding/json"
	"farmlands-backend/models"
	"farmlands-backend/utils"
	"log"
	"net/http"
)

type ProjectDataHandler struct {
	DB *sql.DB
}

func NewProjectDataHandler(db *sql.DB) *ProjectDataHandler {
	return &ProjectDataHandler{DB: db}
}

func (h *ProjectDataHandler) HandleNewProjectData(w http.ResponseWriter, r *http.Request) {
	var input models.AddProjectData
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Println("Error decodificando JSON:", err)
		utils.SendJSONError(w, http.StatusBadRequest, "INVALID_JSON", "JSON inválido")
		return
	}

	if input.Actividad == "" || input.LaborAgronomica == 0 || input.IDProject == 0 || input.Encargado == 0 || input.RecursoHumano == 0 || input.Costo == 0 {
		utils.SendJSONError(w, http.StatusBadRequest, "INVALID_DATA", "Todos los campos son obligatorios")
		return
	}

	stmt := `
		INSERT INTO projects_data (activity, fk_farm_task, fk_project, fk_user, num_human_resources, cost, details)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	res, err := h.DB.Exec(stmt, input.Actividad, input.LaborAgronomica, input.IDProject, input.Encargado, input.RecursoHumano, input.Costo, input.Observaciones)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error agregando datos del proyecto")
		return
	}

	idEntry, _ := res.LastInsertId()

	projectData := models.ProjectData{
		ID:              idEntry,
		Actividad:       input.Actividad,
		LaborAgronomica: input.LaborAgronomica,
		Encargado:       input.Encargado,
		RecursoHumano:   input.RecursoHumano,
		Costo:           input.Costo,
		Observaciones:   input.Observaciones,
	}

	//Insertar en tabla intermedia equipos asociados a la entrada de datos del proyecto

	if len(input.Equipos) != 0 {
		stmt := `INSERT INTO projects_data_tools (fk_projects_data, fk_tools) VALUES (?, ?)`
		for _, id := range input.Equipos {
			_, err := h.DB.Exec(stmt, idEntry, id)
			if err != nil {
				utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error agregando equipos del proyecto")
				return
			}
		}
	}

	utils.SendJSONSuccess(w, projectData)
}

func (h *ProjectDataHandler) HandleDeleteProjectData(w http.ResponseWriter, r *http.Request) {
	var input models.ProjectDataID

	// Decodificar JSON de entrada
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.SendJSONError(w, http.StatusBadRequest, "INVALID_JSON", "JSON inválido")
		return
	}

	res, err := h.DB.Exec("DELETE FROM projects_data WHERE id = ?", input.ID)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error eliminando datos del proyecto")
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		utils.SendJSONError(w, http.StatusNotFound, "NOT_FOUND", "Datos del proyecto no encontrados")
		return
	}

	utils.SendJSONSuccess(w, "Datos del proyecto eliminado correctamente")
}
