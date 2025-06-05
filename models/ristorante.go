package models

// Ristorante rappresenta un ristorante nel sistema
type Ristorante struct {
	ID           int     `json:"id"`
	Nome         string  `json:"nome"`
	NumeroTavoli int     `json:"numero_tavoli"`
	CostoCoperto float64 `json:"costo_coperto"`
}
