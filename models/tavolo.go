package models

// Tavolo rappresenta un tavolo nel ristorante
type Tavolo struct {
	ID           int    `json:"id"`
	MaxPosti     int    `json:"max_posti"`
	Stato        string `json:"stato"` // "libero", "occupato"
	IDRistorante int    `json:"id_ristorante"`
}