package models

type TransitionMap struct {
	Transitions []int `json:"T"` // Transiciones contenidas por el proceso
	Ancestors   []int `json:"A"` // Lista de procesos que preceden al actual
	Successors  []int `json:"S"` // Procesos que reciben de este
	MinTime     int   `json:"M"` // Tiempo mínimo estimado para atravesar la subred
}
