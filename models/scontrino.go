package models

import "time"

// Scontrino rappresenta i dettagli del conto finale di un ordine
type Scontrino struct {
	IDOrdine          int       `json:"id_ordine"`
	IDTavolo          int       `json:"id_tavolo"`
	DataOrdine        time.Time `json:"data_ordine"`
	CostoTotale       float64   `json:"costo_totale"`
	NumCoperti        int       `json:"num_coperti"`
	CostoCoperto      float64   `json:"costo_coperto"`
	ImportoCoperto    float64   `json:"importo_coperto"`
	TotaleComplessivo float64   `json:"totale_complessivo"`
	DataPagamento     time.Time `json:"data_pagamento"`
}
