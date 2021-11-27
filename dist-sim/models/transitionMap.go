package models

type TransitionMap struct {
	Transitions []int `json:"T"` // Transiciones contenidas por el proceso
	Ancestors   []int `json:"A"` // Lista de procesos que preceden al actual
	MinTime     int   `json:"M"` // Tiempo mínimo estimado para atravesar la subred
}
