package models


// Ricetta rappresenta la ricetta di una pietanza
type Ricetta struct {
	ID                int    `json:"id"`
	Nome              string `json:"nome"`
	Descrizione       string `json:"descrizione"`
	IDPietanza        int    `json:"id_pietanza"`
	TempoPreparazione int    `json:"tempo_preparazione"`
	Istruzioni        string `json:"istruzioni"`
}
