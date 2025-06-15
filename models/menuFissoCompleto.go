package models

// MenuFissoCompleto rappresenta un menu fisso con tutte le pietanze che lo compongono
type MenuFissoCompleto struct {
	Menu     MenuFisso  `json:"menu"`
	Pietanze []Pietanza `json:"pietanze"`
}
