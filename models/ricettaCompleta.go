package models

// IngredienteConQuantita rappresenta un ingrediente con la quantità richiesta nella ricetta
type IngredienteConQuantita struct {
	Ingrediente Ingrediente `json:"ingrediente"`
	Quantita    float64     `json:"quantita"`
}

// RicettaCompleta rappresenta una ricetta completa con tutti i suoi ingredienti
type RicettaCompleta struct {
	Ricetta     Ricetta                  `json:"ricetta"`
	Ingredienti []IngredienteConQuantita `json:"ingredienti"`
}
