package models

// Menu rappresenta un menu di ristorante
type Menu struct {
	ID_Ristorante int `json:"id_ristorante"`
	ID_Pietanza   int `json:"id_pietanza"`
}
