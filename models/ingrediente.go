package models


// Ingrediente rappresenta un ingrediente in magazzino
type Ingrediente struct {
	ID                 int     `json:"id"`
	Nome               string  `json:"nome"`
	QuantitaDisponibile float64 `json:"quantita_disponibile"`
	UnitaMisura        string  `json:"unita_misura"`
	SogliaRiordino     float64 `json:"soglia_riordino"`
}