package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"ristorante-api/cache"
	"ristorante-api/models"
	"ristorante-api/repository"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// PietanzaHandler gestisce le richieste relative alle pietanze
type PietanzaHandler struct {
	repo             *repository.PietanzaRepository
	cache            *cache.PietanzaCache
	ricettaRepo      *repository.RicettaRepository
	menuRepo         *repository.MenuFissoRepository
	ingredienteCache *cache.IngredienteCache
}

// NewPietanzaHandler crea un nuovo handler per le pietanze
func NewPietanzaHandler(
	repo *repository.PietanzaRepository,
	cache *cache.PietanzaCache,
	ricettaRepo *repository.RicettaRepository,
	menuRepo *repository.MenuFissoRepository,
	ingredienteCache *cache.IngredienteCache,
) *PietanzaHandler {
	return &PietanzaHandler{
		repo:             repo,
		cache:            cache,
		ricettaRepo:      ricettaRepo,
		menuRepo:         menuRepo,
		ingredienteCache: ingredienteCache,
	}
}

// GetPietanze restituisce tutte le pietanze disponibili
func (h *PietanzaHandler) GetPietanze(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Tenta di recuperare le pietanze dalla cache
	cached, found, err := h.cache.GetAll(ctx)
	if err != nil {
		log.Printf("Errore nell'accesso alla cache: %v", err)
		// Continua con il database in caso di errore della cache
	} else if found {
		// Usa i dati dalla cache
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cached)
		return
	}

	// Cache miss o errore, recupera dal database
	pietanze, err := h.repo.GetAll(ctx)
	if err != nil {
		http.Error(w, "Errore nel recupero delle pietanze", http.StatusInternalServerError)
		log.Printf("Errore nel recupero delle pietanze: %v", err)
		return
	}

	// Salva i risultati in cache per le future richieste
	if err := h.cache.SetAll(ctx, pietanze); err != nil {
		log.Printf("Errore nell'aggiornamento della cache: %v", err)
		// Continua comunque
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pietanze)
}

// GetPietanza restituisce una singola pietanza per ID
func (h *PietanzaHandler) GetPietanza(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID non valido", http.StatusBadRequest)
		return
	}

	// Tenta di recuperare la pietanza dalla cache
	pietanza, found, err := h.cache.GetByID(ctx, id)
	if err != nil {
		log.Printf("Errore nell'accesso alla cache: %v", err)
		// Continua con il database in caso di errore della cache
	} else if found {
		// Usa i dati dalla cache
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pietanza)
		return
	}

	// Cache miss o errore, recupera dal database
	pietanza, err = h.repo.GetByID(ctx, id)
	if err != nil {
		http.Error(w, "Pietanza non trovata", http.StatusNotFound)
		log.Printf("Errore nel recupero della pietanza: %v", err)
		return
	}

	// Salva in cache per le future richieste
	if err := h.cache.SetByID(ctx, id, pietanza); err != nil {
		log.Printf("Errore nell'aggiornamento della cache: %v", err)
		// Continua comunque
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pietanza)
}

// CreatePietanza crea una nuova pietanza
func (h *PietanzaHandler) CreatePietanza(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var p models.Pietanza

	// Decodifica il JSON della richiesta
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Formato JSON non valido", http.StatusBadRequest)
		log.Printf("Errore nella decodifica JSON: %v", err)
		return
	}

	// Validazione base
	if p.Nome == "" || p.Prezzo <= 0 {
		http.Error(w, "Nome e prezzo sono campi obbligatori", http.StatusBadRequest)
		return
	}

	// Inserimento nel database
	err := h.repo.Create(ctx, &p)
	if err != nil {
		http.Error(w, "Errore nella creazione della pietanza", http.StatusInternalServerError)
		log.Printf("Errore nella creazione della pietanza: %v", err)
		return
	}

	// Invalida la cache delle pietanze
	if err := h.cache.InvalidateAll(ctx); err != nil {
		log.Printf("Errore nell'invalidazione della cache: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

// UpdatePietanza aggiorna una pietanza esistente
func (h *PietanzaHandler) UpdatePietanza(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID non valido", http.StatusBadRequest)
		return
	}

	var p models.Pietanza

	// Decodifica il JSON della richiesta
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Formato JSON non valido", http.StatusBadRequest)
		log.Printf("Errore nella decodifica JSON: %v", err)
		return
	}

	// Validazione base
	if p.Nome == "" || p.Prezzo <= 0 {
		http.Error(w, "Nome e prezzo sono campi obbligatori", http.StatusBadRequest)
		return
	}

	// Verifica che la pietanza esista
	exists, err := h.repo.Exists(ctx, id)
	if err != nil || !exists {
		http.Error(w, "Pietanza non trovata", http.StatusNotFound)
		return
	}

	// Imposta l'ID e aggiorna nel database
	p.ID = id
	err = h.repo.Update(ctx, &p)
	if err != nil {
		http.Error(w, "Errore nell'aggiornamento della pietanza", http.StatusInternalServerError)
		log.Printf("Errore nell'aggiornamento della pietanza: %v", err)
		return
	}

	// Invalida la cache per questa pietanza e per l'elenco completo
	if err := h.cache.InvalidateByID(ctx, id); err != nil {
		log.Printf("Errore nell'invalidazione della cache pietanza: %v", err)
	}
	if err := h.cache.InvalidateAll(ctx); err != nil {
		log.Printf("Errore nell'invalidazione della cache pietanze: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

// DeletePietanza elimina una pietanza
func (h *PietanzaHandler) DeletePietanza(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID non valido", http.StatusBadRequest)
		return
	}

	// Verifica che la pietanza esista
	exists, err := h.repo.Exists(ctx, id)
	if err != nil || !exists {
		http.Error(w, "Pietanza non trovata", http.StatusNotFound)
		return
	}

	// Eliminazione dal database
	err = h.repo.Delete(ctx, id)
	if err != nil {
		http.Error(w, "Errore nell'eliminazione della pietanza", http.StatusInternalServerError)
		log.Printf("Errore nell'eliminazione della pietanza: %v", err)
		return
	}

	// Invalida la cache per questa pietanza e per l'elenco completo
	if err := h.cache.InvalidateByID(ctx, id); err != nil {
		log.Printf("Errore nell'invalidazione della cache pietanza: %v", err)
	}
	if err := h.cache.InvalidateAll(ctx); err != nil {
		log.Printf("Errore nell'invalidazione della cache pietanze: %v", err)
	}

	// Risposta 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// AddPietanzaToOrdine aggiunge una pietanza a un ordine esistente
func (h *PietanzaHandler) AddPietanzaToOrdine(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idOrdineStr := chi.URLParam(r, "id_ordine")
	idOrdine, err := strconv.Atoi(idOrdineStr)
	if err != nil {
		http.Error(w, "ID ordine non valido", http.StatusBadRequest)
		return
	}

	// Struttura per la richiesta
	var requestBody struct {
		IDPietanza int `json:"id_pietanza"`
		Quantita   int `json:"quantita"`
	}

	// Decodifica il JSON della richiesta
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Formato JSON non valido", http.StatusBadRequest)
		log.Printf("Errore nella decodifica JSON: %v", err)
		return
	}

	// Validazione base
	if requestBody.IDPietanza <= 0 || requestBody.Quantita <= 0 {
		http.Error(w, "ID pietanza e quantità devono essere maggiori di zero", http.StatusBadRequest)
		return
	}

	// Verifica che la pietanza esista
	exists, err := h.repo.Exists(ctx, requestBody.IDPietanza)
	if err != nil || !exists {
		http.Error(w, "Pietanza non trovata", http.StatusNotFound)
		return
	}

	// Aggiungi la pietanza all'ordine
	err = h.repo.AddPietanzaToOrdine(ctx, idOrdine, requestBody.IDPietanza, requestBody.Quantita, h.ricettaRepo, h.ingredienteCache)
	if err != nil {
		switch err {
		case repository.ErrPietanzaNonDisponibile:
			http.Error(w, "La pietanza non è disponibile", http.StatusBadRequest)
		case repository.ErrIngredientiInsufficienti:
			http.Error(w, "Ingredienti insufficienti per preparare la pietanza", http.StatusBadRequest)
		default:
			http.Error(w, "Errore nell'aggiunta della pietanza all'ordine", http.StatusInternalServerError)
			log.Printf("Errore nell'aggiunta della pietanza all'ordine: %v", err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Pietanza aggiunta all'ordine con successo"})
}

// AddMenuFissoToOrdine aggiunge un menu fisso a un ordine
func (h *PietanzaHandler) AddMenuFissoToOrdine(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idOrdineStr := chi.URLParam(r, "id_ordine")
	idOrdine, err := strconv.Atoi(idOrdineStr)
	if err != nil {
		http.Error(w, "ID ordine non valido", http.StatusBadRequest)
		return
	}

	// Struttura per la richiesta
	var requestBody struct {
		IDMenu int `json:"id_menu"`
	}

	// Decodifica il JSON della richiesta
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Formato JSON non valido", http.StatusBadRequest)
		log.Printf("Errore nella decodifica JSON: %v", err)
		return
	}

	// Validazione base
	if requestBody.IDMenu <= 0 {
		http.Error(w, "ID menu deve essere maggiore di zero", http.StatusBadRequest)
		return
	}

	// Verificare che il menu esista
	menuFisso, err := h.menuRepo.GetByID(ctx, requestBody.IDMenu)
	if err != nil {
		http.Error(w, "Menu fisso non trovato", http.StatusNotFound)
		return
	}

	// Aggiungi il menu all'ordine
	err = h.repo.AddMenuFissoToOrdine(ctx, idOrdine, requestBody.IDMenu, h.ricettaRepo, h.menuRepo, h.ingredienteCache)
	if err != nil {
		switch err {
		case repository.ErrMenuNonDisponibile:
			http.Error(w, "Il menu fisso non è disponibile: una o più pietanze non sono disponibili o mancano ingredienti", http.StatusBadRequest)
		default:
			http.Error(w, "Errore nell'aggiunta del menu fisso all'ordine", http.StatusInternalServerError)
			log.Printf("Errore nell'aggiunta del menu fisso all'ordine: %v", err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":   "Menu fisso aggiunto all'ordine con successo",
		"nome_menu": menuFisso.Nome,
		"prezzo":    fmt.Sprintf("%.2f", menuFisso.Prezzo),
	})
}
