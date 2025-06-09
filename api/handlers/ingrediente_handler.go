package handlers

import (
	"encoding/json"
	"net/http"
	"ristorante-api/cache"
	"ristorante-api/models"
	"ristorante-api/repository"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// IngredienteHandler gestisce le richieste relative agli ingredienti
type IngredienteHandler struct {
	repo  *repository.IngredienteRepository
	cache *cache.IngredienteCache
}

// NewIngredienteHandler crea un nuovo handler per gli ingredienti
func NewIngredienteHandler(repo *repository.IngredienteRepository, cache *cache.IngredienteCache) *IngredienteHandler {
	return &IngredienteHandler{
		repo:  repo,
		cache: cache,
	}
}

// GetIngredienti restituisce tutti gli ingredienti disponibili
func (h *IngredienteHandler) GetIngredienti(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Tenta di recuperare gli ingredienti dalla cache
	cached, found, err := h.cache.GetAll(ctx)
	if err != nil {
		http.Error(w, "Errore nell'accesso alla cache", http.StatusInternalServerError)
		return
	} else if found {
		// Usa i dati dalla cache
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cached)
		return
	}

	// Cache miss o errore, recupera dal database
	ingredienti, err := h.repo.GetAll(ctx)
	if err != nil {
		http.Error(w, "Errore nel recupero degli ingredienti", http.StatusInternalServerError)
		return
	}

	// Salva i risultati in cache per le future richieste
	if err := h.cache.SetAll(ctx, ingredienti); err != nil {
		http.Error(w, "Errore nell'aggiornamento della cache", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ingredienti)
}

// GetIngredienteByID restituisce un ingrediente specifico in base all'ID
func (h *IngredienteHandler) GetIngredienteByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID non valido", http.StatusBadRequest)
		return
	}

	// Tenta di recuperare l'ingrediente dalla cache
	cached, found, err := h.cache.GetByID(ctx, id)
	if err != nil {
		http.Error(w, "Errore nell'accesso alla cache", http.StatusInternalServerError)
		return
	} else if found {
		// Usa i dati dalla cache
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cached)
		return
	}

	// Cache miss o errore, recupera dal database
	ingrediente, err := h.repo.GetByID(ctx, id)
	if err != nil {
		http.Error(w, "Errore nel recupero dell'ingrediente", http.StatusInternalServerError)
		return
	}
	if ingrediente == nil {
		http.NotFound(w, r)
		return
	}

	// Salva il risultato in cache per le future richieste
	if err := h.cache.SetByID(ctx, ingrediente); err != nil {
		http.Error(w, "Errore nell'aggiornamento della cache", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ingrediente)
}

// CreateIngrediente crea un nuovo ingrediente
func (h *IngredienteHandler) CreateIngrediente(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var ingrediente models.Ingrediente

	// Decodifica il corpo della richiesta in un oggetto Ingrediente
	if err := json.NewDecoder(r.Body).Decode(&ingrediente); err != nil {
		http.Error(w, "Errore nella decodifica del corpo della richiesta", http.StatusBadRequest)
		return
	}

	// Crea l'ingrediente nel database
	if err := h.repo.Create(ctx, &ingrediente); err != nil {
		http.Error(w, "Errore nella creazione dell'ingrediente", http.StatusInternalServerError)
		return
	}

	// Invalida la cache degli ingredienti
	if err := h.cache.InvalidateAll(ctx); err != nil {
		http.Error(w, "Errore nell'invalidazione della cache", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ingrediente)
}

// UpdateIngrediente aggiorna un ingrediente esistente
func (h *IngredienteHandler) UpdateIngrediente(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID non valido", http.StatusBadRequest)
		return
	}

	var ingrediente models.Ingrediente
	if err := json.NewDecoder(r.Body).Decode(&ingrediente); err != nil {
		http.Error(w, "Errore nella decodifica del corpo della richiesta", http.StatusBadRequest)
		return
	}

	ingrediente.ID = id

	// Aggiorna l'ingrediente nel database
	if err := h.repo.Update(ctx, &ingrediente); err != nil {
		http.Error(w, "Errore nell'aggiornamento dell'ingrediente", http.StatusInternalServerError)
		return
	}

	// Invalida la cache degli ingredienti
	if err := h.cache.InvalidateByID(ctx, id); err != nil {
		http.Error(w, "Errore nell'invalidazione della cache", http.StatusInternalServerError)
		return
	}
	if err := h.cache.InvalidateAll(ctx); err != nil {
		http.Error(w, "Errore nell'invalidazione della cache", http.StatusInternalServerError)
		return
	}
	if err := h.cache.InvalidateDaRiordinare(ctx); err != nil {
		http.Error(w, "Errore nell'invalidazione della cache degli ingredienti da riordinare", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ingrediente)
}

// DeleteIngrediente elimina un ingrediente per ID
func (h *IngredienteHandler) DeleteIngrediente(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID non valido", http.StatusBadRequest)
		return
	}

	// Elimina l'ingrediente dal database
	if err := h.repo.Delete(ctx, id); err != nil {
		http.Error(w, "Errore nell'eliminazione dell'ingrediente", http.StatusInternalServerError)
		return
	}

	// Invalida la cache dell'ingrediente specifico e di tutti gli ingredienti, anche quelli da riordinare
	if err := h.cache.InvalidateByID(ctx, id); err != nil {
		http.Error(w, "Errore nell'invalidazione della cache", http.StatusInternalServerError)
		return
	}
	if err := h.cache.InvalidateAll(ctx); err != nil {
		http.Error(w, "Errore nell'invalidazione della cache", http.StatusInternalServerError)
		return
	}
	if err := h.cache.InvalidateDaRiordinare(ctx); err != nil {
		http.Error(w, "Errore nell'invalidazione della cache degli ingredienti da riordinare", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetIngredientiDaRiordinare restituisce gli ingredienti sotto la soglia di riordino
func (h *IngredienteHandler) GetIngredientiDaRiordinare(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Tenta di recuperare gli ingredienti da riordinare dalla cache
	cached, found, err := h.cache.GetDaRiordinare(ctx)
	if err != nil {
		http.Error(w, "Errore nell'accesso alla cache", http.StatusInternalServerError)
		return
	} else if found {
		// Usa i dati dalla cache
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cached)
		return
	}

	// Cache miss o errore
	// Recupera gli ingredienti da riordinare dal database
	ingredienti, err := h.repo.IngredientiDaRiordinare(ctx)
	if err != nil {
		http.Error(w, "Errore nel recupero degli ingredienti da riordinare", http.StatusInternalServerError)
		return
	}

	// Salva i risultati in cache per le future richieste
	if err := h.cache.SetDaRiordinare(ctx, ingredienti); err != nil {
		http.Error(w, "Errore nell'aggiornamento della cache", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ingredienti)
}

// Prenota una quantità di un ingrediente
func (h *IngredienteHandler) PrenotaIngrediente(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID non valido", http.StatusBadRequest)
		return
	}

	var richiesta struct {
		Quantita float64 `json:"quantita"`
	}
	if err := json.NewDecoder(r.Body).Decode(&richiesta); err != nil {
		http.Error(w, "Errore nella decodifica del corpo della richiesta", http.StatusBadRequest)
		return
	}

	// Prenota l'ingrediente nel database
	if err := h.repo.Prenota(ctx, id, richiesta.Quantita); err != nil {
		http.Error(w, "Errore nella prenotazione dell'ingrediente", http.StatusInternalServerError)
		return
	}

	// Invalida la cache degli ingredienti e degli ingredienti da riordinare
	if err := h.cache.InvalidateAll(ctx); err != nil {
		http.Error(w, "Errore nell'invalidazione della cache", http.StatusInternalServerError)
		return
	}

	if err := h.cache.InvalidateByID(ctx, id); err != nil {
		http.Error(w, "Errore nell'invalidazione della cache", http.StatusInternalServerError)
		return
	}

	if err := h.cache.InvalidateDaRiordinare(ctx); err != nil {
		http.Error(w, "Errore nell'invalidazione della cache degli ingredienti da riordinare", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Rifornisci un ingrediente con una certa quantità
func (h *IngredienteHandler) RifornisciIngrediente(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID non valido", http.StatusBadRequest)
		return
	}

	var richiesta struct {
		Quantita float64 `json:"quantita"`
	}
	if err := json.NewDecoder(r.Body).Decode(&richiesta); err != nil {
		http.Error(w, "Errore nella decodifica del corpo della richiesta", http.StatusBadRequest)
		return
	}

	// Rifornisci l'ingrediente nel database
	if err := h.repo.Rifornisci(ctx, id, richiesta.Quantita); err != nil {
		http.Error(w, "Errore nel rifornimento dell'ingrediente", http.StatusInternalServerError)
		return
	}

	// Invalida la cache degli ingredienti e degli ingredienti da riordinare
	if err := h.cache.InvalidateAll(ctx); err != nil {
		http.Error(w, "Errore nell'invalidazione della cache", http.StatusInternalServerError)
		return
	}

	if err := h.cache.InvalidateByID(ctx, id); err != nil {
		http.Error(w, "Errore nell'invalidazione della cache", http.StatusInternalServerError)
		return
	}

	if err := h.cache.InvalidateDaRiordinare(ctx); err != nil {
		http.Error(w, "Errore nell'invalidazione della cache degli ingredienti da riordinare", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
