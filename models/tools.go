package models

// Representación de equipos e implementos
type Tool struct {
	ID          int64  `json:"id"`
	Descripcion string `json:"descripcion"`
}

type ToolID struct {
	ID int64 `json:"id"`
}

type AddTool struct {
	Descripcion string `json:"descripcion"`
}
