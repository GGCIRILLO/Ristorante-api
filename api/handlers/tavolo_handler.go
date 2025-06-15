package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"ristorante-api/cache"
	"ristorante-api/models"
	"ristorante-api/repository"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type TavoloHandler struct {
	Repo  *repository.TavoloRepository
	Cache *cache.TavoloCache
}

func NewTavoloHandler(repo *repository.TavoloRepository, cache *cache.TavoloCache) *TavoloHandler {
	return &TavoloHandler{
		Repo:  repo,
		Cache: cache,
	}
}

func (h *TavoloHandler) GetTavoli(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tavoli, err := h.Cache.GetTavoli(ctx)
	if err != nil {
		log.Printf("Errore cache GetTavoli: %v", err)
	}
	if tavoli == nil {
		tavoli, err = h.Repo.GetAll(ctx)
		if err != nil {
			http.Error(w, "Errore recupero tavoli", http.StatusInternalServerError)
			return
		}
		_ = h.Cache.SetTavoli(ctx, tavoli)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tavoli)
}

func (h *TavoloHandler) GetTavolo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID non valido", http.StatusBadRequest)
		return
	}
	tavolo, err := h.Repo.GetByID(ctx, id)
	if err != nil {
		http.Error(w, "Tavolo non trovato", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tavolo)
}

func (h *TavoloHandler) CreateTavolo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var t models.Tavolo
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "JSON non valido", http.StatusBadRequest)
		return
	}
	if t.MaxPosti <= 0 || t.IDRistorante <= 0 {
		http.Error(w, "MaxPosti e IDRistorante obbligatori", http.StatusBadRequest)
		return
	}
	if t.Stato == "" {
		t.Stato = "libero"
	}
	if err := h.Repo.Create(ctx, &t); err != nil {
		http.Error(w, "Errore creazione tavolo", http.StatusInternalServerError)
		return
	}
	_ = h.Cache.InvalidateTavoli(ctx)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

func (h *TavoloHandler) UpdateTavolo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID non valido", http.StatusBadRequest)
		return
	}
	var t models.Tavolo
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "JSON non valido", http.StatusBadRequest)
		return
	}
	if t.MaxPosti <= 0 || t.IDRistorante <= 0 {
		http.Error(w, "MaxPosti e IDRistorante obbligatori", http.StatusBadRequest)
		return
	}
	if err := h.Repo.Update(ctx, id, t); err != nil {
		http.Error(w, "Errore aggiornamento", http.StatusInternalServerError)
		return
	}
	_ = h.Cache.InvalidateTavoli(ctx)
	t.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

func (h *TavoloHandler) DeleteTavolo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID non valido", http.StatusBadRequest)
		return
	}
	if err := h.Repo.Delete(ctx, id); err != nil {
		http.Error(w, "Errore eliminazione", http.StatusInternalServerError)
		return
	}
	_ = h.Cache.InvalidateTavoli(ctx)
	w.WriteHeader(http.StatusNoContent)
}

func (h *TavoloHandler) CambiaStatoTavolo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID non valido", http.StatusBadRequest)
		return
	}
	var body struct {
		Stato string `json:"stato"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "JSON non valido", http.StatusBadRequest)
		return
	}
	if body.Stato != "libero" && body.Stato != "occupato" {
		http.Error(w, "Stato non valido", http.StatusBadRequest)
		return
	}
	t, err := h.Repo.CambiaStato(ctx, id, body.Stato)
	if err != nil {
		http.Error(w, "Errore aggiornamento stato", http.StatusInternalServerError)
		return
	}
	_ = h.Cache.InvalidateTavoli(ctx)
	_ = h.Cache.InvalidateTavoliLiberi(ctx)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

// GetTavoliLiberi restituisce i tavoli liberi per un ristorante specifico
func (h *TavoloHandler) GetTavoliLiberi(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tavoli, err := h.Cache.GetTavoliLiberi(ctx)
	if err != nil {
		log.Printf("Errore cache GetTavoliLiberi: %v", err)
	}
	if tavoli == nil {
		tavoli, err = h.Repo.GetTavoliLiberi(ctx, 1) // Supponiamo ID ristorante 1
		if err != nil {
			http.Error(w, "Errore recupero tavoli liberi", http.StatusInternalServerError)
			return
		}
		_ = h.Cache.SetTavoliLiberi(ctx, tavoli)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tavoli)
}
