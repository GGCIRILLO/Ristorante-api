package models

import (
	"fmt"
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

// ErrOrdineNonTrovato è un errore personalizzato restituito quando non viene trovato un ordine
// con determinate caratteristiche
type ErrOrdineNonTrovato struct {
	IDTavolo        int
	StatoRichiesto  string
	MessaggioErrore string
}

func (e *ErrOrdineNonTrovato) Error() string {
	return fmt.Sprintf("%s (Tavolo ID: %d, Stato: %s)",
		e.MessaggioErrore, e.IDTavolo, e.StatoRichiesto)
}
