package models

// Representación de equipos e implementos
type ProjectData struct {
	ID              int64   `json:"id"`
	Actividad       string  `json:"actividad"`
	LaborAgronomica int64   `json:"laborAgronomica"`
	Encargado       int64   `json:"encargado"`
	Equipos         []int64 `json:"equipos"`
	RecursoHumano   int64   `json:"recursoHumano"`
	Costo           float64 `json:"costo"`
	Observaciones   string  `json:"observaciones"`
}

type ProjectDataID struct {
	ID int64 `json:"id"`
}

type AddProjectData struct {
	Actividad       string  `json:"actividad"`
	IDProject       int64   `json:"idproject"`
	LaborAgronomica int64   `json:"laborAgronomica"`
	Encargado       int64   `json:"encargado"`
	Equipos         []int64 `json:"equipos"`
	RecursoHumano   int64   `json:"recursoHumano"`
	Costo           float64 `json:"costo"`
	Observaciones   string  `json:"observaciones"`
}

type ProjectDataSupervisorID struct {
	ID int64 `json:"idSupervisor"`
}
