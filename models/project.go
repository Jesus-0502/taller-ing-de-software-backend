package models

// Project representa un proyecto agrícola
type Project struct {
	ID          int64  `json:"id"`
	Descripcion string `json:"descripcion"`
	FechaInicio string `json:"fecha_inicio"`
	FechaCierre string `json:"fecha_cierre"`
	Estado      string `json:"estado"`
	CreatedAt   string `json:"created_at"`
}

// CreateProjectInput representa los datos necesarios para crear un proyecto
type CreateProjectInput struct {
	Descripcion string `json:"descripcion"`
	FechaInicio string `json:"fecha_inicio"`
	FechaCierre string `json:"fecha_cierre"`
}
