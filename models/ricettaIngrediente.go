package models

// RicettaIngrediente rappresenta la relazione tra ricetta e ingrediente
type RicettaIngrediente struct {
	IDRicetta     int     `json:"id_ricetta"`
	IDIngrediente int     `json:"id_ingrediente"`
	Quantita      float64 `json:"quantita"`
}