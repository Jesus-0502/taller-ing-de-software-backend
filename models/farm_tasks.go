package models

// Representación de labores agronómicas
type FarmTask struct {
	ID          int64  `json:"id"`
	Descripcion string `json:"descripcion"`
}

type AddFarmTask struct {
	Descripcion string `json:"descripcion"`
}
