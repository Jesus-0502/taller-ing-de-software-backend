package models

import "time"

// Project representa un proyecto agrícola
type Project struct {
	ID          int64     `json:"id"`
	Descripcion string    `json:"descripcion"`
	FechaInicio time.Time `json:"fecha_inicio"`
	FechaCierre time.Time `json:"fecha_cierre"`
	Estado      string    `json:"estado"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateProjectInput representa los datos necesarios para crear un proyecto
type CreateProjectInput struct {
	Descripcion string    `json:"descripcion"`
	FechaInicio time.Time `json:"fecha_inicio"`
	FechaCierre time.Time `json:"fecha_cierre"`
}
