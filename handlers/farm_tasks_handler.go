package handlers

import (
	"database/sql"
	"encoding/json"
	"farmlands-backend/models"
	"farmlands-backend/utils"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type FarmTasksHandler struct {
	DB *sql.DB
}

func NewFarmTasksHandlerHandler(db *sql.DB) *FarmTasksHandler {
	return &FarmTasksHandler{DB: db}
}

func (h *FarmTasksHandler) HandleListFarmTasks(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`
		SELECT id, descripcion FROM farm_tasks
	`)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error al obtener labores agronómicas")
		return
	}
	defer rows.Close()

	var farmTasks []models.FarmTask
	for rows.Next() {
		var p models.FarmTask

		if err := rows.Scan(&p.ID, &p.Descripcion); err != nil {
			utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error leyendo los datos")
			return
		}
		farmTasks = append(farmTasks, p)
	}

	if err := rows.Err(); err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error iterando resultados")
		return
	}

	if farmTasks == nil {
		farmTasks = []models.FarmTask{}
	}
	utils.SendJSONSuccess(w, farmTasks)
}

func (h *FarmTasksHandler) HandleAddFarmTask(w http.ResponseWriter, r *http.Request) {
	var input models.AddFarmTask
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
		INSERT INTO farm_tasks (descripcion)
		VALUES (?)
	`

	res, err := h.DB.Exec(stmt, input.Descripcion)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error agregando nueva labor agronómica")
		return
	}

	id, _ := res.LastInsertId()

	farmTask := models.FarmTask{
		ID:          id,
		Descripcion: input.Descripcion,
	}

	utils.SendJSONSuccess(w, farmTask)
}

func (h *FarmTasksHandler) HandleEditFarmTask(w http.ResponseWriter, r *http.Request) {
	var input models.FarmTask

	// Decodificar JSON de entrada
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.SendJSONError(w, http.StatusBadRequest, "INVALID_JSON", "JSON inválido")
		return
	}

	// Validar que venga el ID
	if input.ID == 0 {
		utils.SendJSONError(w, http.StatusBadRequest, "MISSING_ID", "El ID de la tarea es obligatorio")
		return
	}

	// Buscar tarea actual en BD
	var dbFarmTask models.FarmTask
	err := h.DB.QueryRow(`
		SELECT id, descripcion
		FROM farm_tasks WHERE id = ?`, input.ID,
	).Scan(
		&dbFarmTask.ID, &dbFarmTask.Descripcion,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.SendJSONError(w, http.StatusNotFound, "TOOL_NOT_FOUND", "Labor no encontrada")
			return
		}
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error consultando labor")
		return
	}

	// Comparar y construir dinámicamente la sentencia UPDATE
	updateFields := []string{}
	args := []interface{}{}

	if input.Descripcion != "" && input.Descripcion != dbFarmTask.Descripcion {
		updateFields = append(updateFields, "descripcion = ?")
		args = append(args, input.Descripcion)
	}

	// Si no hay cambios, devolver mensaje
	if len(updateFields) == 0 {
		utils.SendJSONSuccess(w, map[string]string{"message": "No se realizaron cambios"})
		return
	}

	// Ejecutar actualización dinámica
	query := fmt.Sprintf("UPDATE farm_tasks SET %s WHERE id = ?", strings.Join(updateFields, ", "))
	args = append(args, input.ID)

	_, err = h.DB.Exec(query, args...)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_UPDATE_ERROR", "Error actualizando suplemento")
		return
	}

	// Devolver tool actualizado
	if input.Descripcion != "" {
		dbFarmTask.Descripcion = input.Descripcion
	}

	utils.SendJSONSuccess(w, dbFarmTask)
}

func (h *FarmTasksHandler) HandleDeleteFarmTask(w http.ResponseWriter, r *http.Request) {
	var input models.FarmTaskID

	// Decodificar JSON de entrada
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.SendJSONError(w, http.StatusBadRequest, "INVALID_JSON", "JSON inválido")
		return
	}

	res, err := h.DB.Exec("DELETE FROM farm_tasks WHERE id = ?", input.ID)
	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error eliminando suplemento")
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		utils.SendJSONError(w, http.StatusNotFound, "NOT_FOUND", "Labor no encontrada")
		return
	}

	utils.SendJSONSuccess(w, "Labor eliminada correctamente")
}

func (h *FarmTasksHandler) HandleSearchFarmTask(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q") // texto buscado

	// Si no hay query, devolvemos todas las labores
	var rows *sql.Rows
	var err error
	if query == "" {
		rows, err = h.DB.Query("SELECT id, descripcion FROM farm_tasks")
	} else {
		rows, err = h.DB.Query(
			"SELECT id, descripcion FROM farm_tasks WHERE UPPER(descripcion) LIKE UPPER(?)",
			"%"+query+"%",
		)
	}

	if err != nil {
		utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error en la búsqueda")
		return
	}
	defer rows.Close()

	var farmTasks []models.FarmTask
	for rows.Next() {
		var f models.FarmTask
		if err := rows.Scan(&f.ID, &f.Descripcion); err != nil {
			utils.SendJSONError(w, http.StatusInternalServerError, "DB_ERROR", "Error leyendo resultados")
			return
		}
		farmTasks = append(farmTasks, f)
	}

	utils.SendJSONSuccess(w, farmTasks)
}
