package handlers

import (
	"database/sql"
	"farmlands-backend/models"
	"farmlands-backend/utils"
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

// func (h *FarmTasksHandler) HandleAddFarmTask(w http.ResponseWriter, r *http.Request)

// func (h *FarmTasksHandler) HandleEditFarmTask(w http.ResponseWriter, r *http.Request)

// func (h *FarmTasksHandler) HandleDeleteFarmTask(w http.ResponseWriter, r *http.Request)

// func (h *FarmTasksHandler) HandleSearchFarmTask(w http.ResponseWriter, r *http.Request)
