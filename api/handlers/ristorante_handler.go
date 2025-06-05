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

type RistoranteHandler struct {
	Repo  *repository.RistoranteRepository
	Cache *cache.RistoranteCache
}

func NewRistoranteHandler(repo *repository.RistoranteRepository, cache *cache.RistoranteCache) *RistoranteHandler {
	return &RistoranteHandler{
		Repo:  repo,
		Cache: cache,
	}
}

func (h *RistoranteHandler) GetRistoranti(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Prova a leggere dalla cache
	ristoranti, err := h.Cache.GetAll(ctx)
	if err != nil {
		log.Printf("Errore nella lettura da cache: %v", err)
	}
	if ristoranti != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ristoranti)
		return
	}

	// Cache miss → query su DB
	ristoranti, err = h.Repo.GetAll(ctx)
	if err != nil {
		http.Error(w, "Errore nel recupero dei ristoranti", http.StatusInternalServerError)
		return
	}

	_ = h.Cache.SetAll(ctx, ristoranti)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ristoranti)
}

func (h *RistoranteHandler) GetRistorante(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID non valido", http.StatusBadRequest)
		return
	}

	risto, err := h.Repo.GetByID(ctx, id)
	if err != nil {
		http.Error(w, "Ristorante non trovato", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(risto)
}

func (h *RistoranteHandler) CreateRistorante(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var risto models.Ristorante

	if err := json.NewDecoder(r.Body).Decode(&risto); err != nil {
		http.Error(w, "Formato JSON non valido", http.StatusBadRequest)
		return
	}

	if risto.Nome == "" || risto.NumeroTavoli <= 0 {
		http.Error(w, "Campi obbligatori mancanti o non validi", http.StatusBadRequest)
		return
	}

	if err := h.Repo.Create(ctx, &risto); err != nil {
		http.Error(w, "Errore nella creazione", http.StatusInternalServerError)
		return
	}

	_ = h.Cache.DeleteAll(ctx)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(risto)
}

func (h *RistoranteHandler) UpdateRistorante(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID non valido", http.StatusBadRequest)
		return
	}

	var risto models.Ristorante
	if err := json.NewDecoder(r.Body).Decode(&risto); err != nil {
		http.Error(w, "Formato JSON non valido", http.StatusBadRequest)
		return
	}

	exists, err := h.Repo.Exists(ctx, id)
	if err != nil || !exists {
		http.Error(w, "Ristorante non trovato", http.StatusNotFound)
		return
	}

	if err := h.Repo.Update(ctx, id, risto); err != nil {
		http.Error(w, "Errore aggiornamento", http.StatusInternalServerError)
		return
	}

	_ = h.Cache.DeleteAll(ctx)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(risto)
}

func (h *RistoranteHandler) DeleteRistorante(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID non valido", http.StatusBadRequest)
		return
	}

	exists, err := h.Repo.Exists(ctx, id)
	if err != nil || !exists {
		http.Error(w, "Ristorante non trovato", http.StatusNotFound)
		return
	}

	if err := h.Repo.Delete(ctx, id); err != nil {
		http.Error(w, "Errore eliminazione", http.StatusInternalServerError)
		return
	}

	_ = h.Cache.DeleteAll(ctx)
	w.WriteHeader(http.StatusNoContent)
}
