package models

// DettaglioPietanza estende DettaglioOrdine con i dettagli della pietanza
type DettaglioPietanza struct {
	ID          int      `json:"id"`
	IDOrdine    int      `json:"id_ordine"`
	Pietanza    Pietanza `json:"pietanza"`
	Quantita    int      `json:"quantita"`
	ParteDiMenu bool     `json:"parte_di_menu"`
	IDMenu      *int     `json:"id_menu,omitempty"`
}

// DettaglioMenuFisso contiene un menu fisso con le sue pietanze
type DettaglioMenuFisso struct {
	Menu     MenuFisso           `json:"menu"`
	Pietanze []DettaglioPietanza `json:"pietanze"`
}

// OrdineCompleto rappresenta un ordine con tutti i dettagli delle pietanze e menu fissi
type OrdineCompleto struct {
	Ordine    Ordine               `json:"ordine"`
	Pietanze  []DettaglioPietanza  `json:"pietanze"`
	MenuFissi []DettaglioMenuFisso `json:"menu_fissi,omitempty"`
}
