package models

import (
	"time"
)

// Ordine rappresenta un ordine effettuato da un tavolo
type Ordine struct {
	ID           int       `json:"id"`
	IDTavolo     int       `json:"id_tavolo"`
	NumPersone   int       `json:"num_persone"`
	DataOrdine   time.Time `json:"data_ordine"`
	Stato        string    `json:"stato"`
	IDRistorante int       `json:"id_ristorante"`
	CostoTotale  float64   `json:"costo_totale"`
}
