package models


// Pietanza rappresenta un piatto nel menu
type Pietanza struct {
	ID          int     `json:"id"`
	Nome        string  `json:"nome"`
	Prezzo      float64 `json:"prezzo"`
	IDCategoria *int    `json:"id_categoria,omitempty"`
	Disponibile bool    `json:"disponibile"`
}
