package models

// MenuFisso rappresenta un menu a prezzo fisso
type MenuFisso struct {
	ID          int     `json:"id"`
	Nome        string  `json:"nome"`
	Prezzo      float64 `json:"prezzo"`
	Descrizione string  `json:"descrizione"`
}