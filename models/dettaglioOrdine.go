package models

// DettaglioOrdine rappresenta una riga di un ordine, collegata a una pietanza
type DettaglioOrdine struct {
	ID             int  `json:"id"`
	IDOrdine       int  `json:"id_ordine"`
	IDPietanza     int  `json:"id_pietanza"`
	Quantita       int  `json:"quantita"`
	ParteDiMenu    bool `json:"parte_di_menu"`
	IDMenu         *int `json:"id_menu,omitempty"`
} 