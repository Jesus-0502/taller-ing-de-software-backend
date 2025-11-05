package handlers

import (
	"database/sql"
	"encoding/json"
	"farmlands-backend/models"
	"farmlands-backend/utils"
	"log"
	"net/http"
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

// func (h *FarmTasksHandler) HandleEditFarmTask(w http.ResponseWriter, r *http.Request)

// func (h *FarmTasksHandler) HandleDeleteFarmTask(w http.ResponseWriter, r *http.Request)

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
