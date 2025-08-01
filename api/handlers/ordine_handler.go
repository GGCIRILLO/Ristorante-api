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

type OrdineHandler struct {
	Repo  *repository.OrdineRepository
	Cache *cache.OrdineCache
}

func NewOrdineHandler(repo *repository.OrdineRepository, cache *cache.OrdineCache) *OrdineHandler {
	return &OrdineHandler{Repo: repo, Cache: cache}
}

func (h *OrdineHandler) GetOrdini(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ordini, err := h.Cache.GetAll(ctx)
	if err != nil || ordini == nil {
		ordini, err = h.Repo.GetAll(ctx)
		if err != nil {
			http.Error(w, "Errore nel recupero degli ordini", http.StatusInternalServerError)
			log.Printf("Errore nel recupero: %v", err)
			return
		}
		h.Cache.SetAll(ctx, ordini)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ordini)
}

func (h *OrdineHandler) GetOrdine(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID non valido", http.StatusBadRequest)
		return
	}

	// Altrimenti recupera l'ordine base
	ordine, err := h.Repo.GetByID(ctx, id)
	if err != nil {
		http.Error(w, "Ordine non trovato", http.StatusNotFound)
		log.Printf("Errore nel recupero ordine: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ordine)
}

// GetOrdineCompleto recupera un ordine con tutti i suoi dettagli (pietanze e menu)
func (h *OrdineHandler) GetOrdineCompleto(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID non valido", http.StatusBadRequest)
		return
	}

	ordineCompleto, err := h.Repo.GetOrdineCompleto(ctx, id)
	if err != nil {
		http.Error(w, "Ordine non trovato", http.StatusNotFound)
		log.Printf("Errore nel recupero ordine completo: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ordineCompleto)
}

func (h *OrdineHandler) CreateOrdine(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var ordine models.Ordine
	if err := json.NewDecoder(r.Body).Decode(&ordine); err != nil {
		http.Error(w, "JSON non valido", http.StatusBadRequest)
		return
	}
	if ordine.IDTavolo <= 0 || ordine.NumPersone <= 0 || ordine.IDRistorante <= 0 {
		http.Error(w, "Tutti i campi obbligatori devono essere validi", http.StatusBadRequest)
		return
	}
	if err := h.Repo.Create(ctx, &ordine); err != nil {
		http.Error(w, "Errore nella creazione dell'ordine", http.StatusInternalServerError)
		log.Printf("Errore creazione ordine: %v", err)
		return
	}
	h.Cache.Invalidate(ctx)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ordine)
}

func (h *OrdineHandler) UpdateStatoOrdine(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Formato JSON non valido", http.StatusBadRequest)
		return
	}
	ordine, err := h.Repo.UpdateStato(ctx, id, body.Stato)
	if err != nil {
		http.Error(w, "Errore nell'aggiornamento stato ordine", http.StatusInternalServerError)
		log.Printf("Errore aggiornamento stato: %v", err)
		return
	}
	h.Cache.Invalidate(ctx)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ordine)
}

func (h *OrdineHandler) DeleteOrdine(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID non valido", http.StatusBadRequest)
		return
	}
	if err := h.Repo.Delete(ctx, id); err != nil {
		http.Error(w, "Errore nella cancellazione dell'ordine", http.StatusInternalServerError)
		log.Printf("Errore cancellazione ordine: %v", err)
		return
	}
	h.Cache.Invalidate(ctx)
	w.WriteHeader(http.StatusNoContent)
}

// CalcolaScontrino calcola il conto per un tavolo
func (h *OrdineHandler) CalcolaScontrino(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idTavoloStr := chi.URLParam(r, "id_tavolo")
	idTavolo, err := strconv.Atoi(idTavoloStr)
	if err != nil {
		http.Error(w, "ID tavolo non valido", http.StatusBadRequest)
		return
	}

	// Calcola lo scontrino
	scontrino, err := h.Repo.CalcolaScontrino(ctx, idTavolo)
	if err != nil {
		// Verifica se è un errore personalizzato ErrOrdineNonTrovato
		if ordineErr, ok := err.(*models.ErrOrdineNonTrovato); ok {
			http.Error(w, ordineErr.Error(), http.StatusNotFound)
			log.Printf("Ordine non trovato: %v", ordineErr)
			return
		}
		// Altrimenti è un errore generico
		http.Error(w, "Errore nel calcolo dello scontrino", http.StatusInternalServerError)
		log.Printf("Errore nel calcolo dello scontrino: %v", err)
		return
	}

	// Invalida la cache degli ordini
	if h.Cache != nil {
		h.Cache.Invalidate(ctx)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scontrino)
}

// GetAllOrdiniCompleti recupera tutti gli ordini con i dettagli completi (pietanze e menu)
func (h *OrdineHandler) GetAllOrdiniCompleti(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Tenta di recuperare dalla cache
	ordiniCompleti, err := h.Cache.GetAllOrdiniCompleti(ctx)
	if err != nil || ordiniCompleti == nil {
		// Cache miss o errore, recupera dal database
		ordiniCompleti, err = h.Repo.GetAllOrdiniCompleti(ctx)
		if err != nil {
			http.Error(w, "Errore nel recupero degli ordini completi", http.StatusInternalServerError)
			log.Printf("Errore nel recupero degli ordini completi: %v", err)
			return
		}

		// Salva in cache per le future richieste
		if err := h.Cache.SetAllOrdiniCompleti(ctx, ordiniCompleti); err != nil {
			log.Printf("Errore nell'aggiornamento della cache per ordini completi: %v", err)
			// Continua comunque
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ordiniCompleti)
}
