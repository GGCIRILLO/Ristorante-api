package models

// ComposizioneMenuFisso rappresenta le pietanze contenute in un menu fisso
type ComposizioneMenuFisso struct {
	IDMenu     int `json:"id_menu"`
	IDPietanza int `json:"id_pietanza"`
}
