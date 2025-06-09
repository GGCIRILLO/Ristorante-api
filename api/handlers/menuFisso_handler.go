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

// MenuFissoHandler gestisce le richieste relative ai menu fissi
type MenuFissoHandler struct {
	repo  *repository.MenuFissoRepository
	cache *cache.MenuFissoCache
}

// NewMenuFissoHandler crea un nuovo handler per i menu fissi
func NewMenuFissoHandler(repo *repository.MenuFissoRepository, cache *cache.MenuFissoCache) *MenuFissoHandler {
	return &MenuFissoHandler{
		repo:  repo,
		cache: cache,
	}
}

// GetMenuFissi restituisce tutti i menu fissi disponibili
func (h *MenuFissoHandler) GetMenuFissi(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Tenta di recuperare i menu fissi dalla cache
	cached, found, err := h.cache.GetAll(ctx)
	if err != nil {
		log.Printf("Errore nell'accesso alla cache: %v", err)
		// Continua con il database in caso di errore della cache
	} else if found {
		// Usa i dati dalla cache
		log.Println("Servendo i menu fissi dalla cache Redis")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cached)
		return
	}

	// Cache miss o errore, recupera dal database
	log.Println("Cache miss, recupero i menu fissi dal database")
	menuFissi, err := h.repo.GetAll(ctx)
	if err != nil {
		http.Error(w, "Errore nel recupero dei menu fissi", http.StatusInternalServerError)
		log.Printf("Errore nel recupero dei menu fissi: %v", err)
		return
	}

	// Salva i risultati in cache per le future richieste
	if err := h.cache.SetAll(ctx, menuFissi); err != nil {
		log.Printf("Errore nell'aggiornamento della cache: %v", err)
		// Continua comunque
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(menuFissi)
}

// GetMenuFisso restituisce un singolo menu fisso per ID
func (h *MenuFissoHandler) GetMenuFisso(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID non valido", http.StatusBadRequest)
		return
	}

	// Tenta di recuperare il menu fisso dalla cache
	menuFisso, found, err := h.cache.GetByID(ctx, id)
	if err != nil {
		log.Printf("Errore nell'accesso alla cache: %v", err)
		// Continua con il database in caso di errore della cache
	} else if found {
		// Usa i dati dalla cache
		log.Printf("Servendo il menu fisso ID %d dalla cache Redis", id)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(menuFisso)
		return
	}

	// Cache miss o errore, recupera dal database
	menuFisso, err = h.repo.GetByID(ctx, id)
	if err != nil {
		http.Error(w, "Menu fisso non trovato", http.StatusNotFound)
		log.Printf("Errore nel recupero del menu fisso: %v", err)
		return
	}

	// Salva in cache per le future richieste
	if err := h.cache.SetByID(ctx, id, menuFisso); err != nil {
		log.Printf("Errore nell'aggiornamento della cache: %v", err)
		// Continua comunque
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(menuFisso)
}

// CreateMenuFisso crea un nuovo menu fisso
func (h *MenuFissoHandler) CreateMenuFisso(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var m models.MenuFisso

	// Decodifica il JSON della richiesta
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "Formato JSON non valido", http.StatusBadRequest)
		log.Printf("Errore nella decodifica JSON: %v", err)
		return
	}

	// Validazione base
	if m.Nome == "" || m.Prezzo <= 0 {
		http.Error(w, "Nome e prezzo sono campi obbligatori", http.StatusBadRequest)
		return
	}

	// Inserimento nel database
	err := h.repo.Create(ctx, &m)
	if err != nil {
		http.Error(w, "Errore nella creazione del menu fisso", http.StatusInternalServerError)
		log.Printf("Errore nella creazione del menu fisso: %v", err)
		return
	}

	// Invalida la cache dei menu fissi
	if err := h.cache.InvalidateAll(ctx); err != nil {
		log.Printf("Errore nell'invalidazione della cache: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m)
}

// UpdateMenuFisso aggiorna un menu fisso esistente
func (h *MenuFissoHandler) UpdateMenuFisso(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID non valido", http.StatusBadRequest)
		return
	}

	var m models.MenuFisso

	// Decodifica il JSON della richiesta
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "Formato JSON non valido", http.StatusBadRequest)
		log.Printf("Errore nella decodifica JSON: %v", err)
		return
	}

	// Validazione base
	if m.Nome == "" || m.Prezzo <= 0 {
		http.Error(w, "Nome e prezzo sono campi obbligatori", http.StatusBadRequest)
		return
	}

	// Verifica che il menu fisso esista
	exists, err := h.repo.Exists(ctx, id)
	if err != nil || !exists {
		http.Error(w, "Menu fisso non trovato", http.StatusNotFound)
		return
	}

	// Imposta l'ID e aggiorna nel database
	m.ID = id
	err = h.repo.Update(ctx, &m)
	if err != nil {
		http.Error(w, "Errore nell'aggiornamento del menu fisso", http.StatusInternalServerError)
		log.Printf("Errore nell'aggiornamento del menu fisso: %v", err)
		return
	}

	// Invalida la cache per questo menu fisso e per l'elenco completo
	if err := h.cache.InvalidateByID(ctx, id); err != nil {
		log.Printf("Errore nell'invalidazione della cache menu fisso: %v", err)
	}
	if err := h.cache.InvalidateComposizione(ctx, id); err != nil {
		log.Printf("Errore nell'invalidazione della cache composizione menu: %v", err)
	}
	if err := h.cache.InvalidateAll(ctx); err != nil {
		log.Printf("Errore nell'invalidazione della cache menu fissi: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

// DeleteMenuFisso elimina un menu fisso
func (h *MenuFissoHandler) DeleteMenuFisso(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID non valido", http.StatusBadRequest)
		return
	}

	// Verifica che il menu fisso esista
	exists, err := h.repo.Exists(ctx, id)
	if err != nil || !exists {
		http.Error(w, "Menu fisso non trovato", http.StatusNotFound)
		return
	}

	// Eliminazione dal database
	err = h.repo.Delete(ctx, id)
	if err != nil {
		http.Error(w, "Errore nell'eliminazione del menu fisso", http.StatusInternalServerError)
		log.Printf("Errore nell'eliminazione del menu fisso: %v", err)
		return
	}

	// Invalida la cache per questo menu fisso e per l'elenco completo
	if err := h.cache.InvalidateByID(ctx, id); err != nil {
		log.Printf("Errore nell'invalidazione della cache menu fisso: %v", err)
	}
	if err := h.cache.InvalidateComposizione(ctx, id); err != nil {
		log.Printf("Errore nell'invalidazione della cache composizione menu: %v", err)
	}
	if err := h.cache.InvalidateAll(ctx); err != nil {
		log.Printf("Errore nell'invalidazione della cache menu fissi: %v", err)
	}

	// Risposta 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// GetComposizione restituisce le pietanze che compongono un menu fisso
func (h *MenuFissoHandler) GetComposizione(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID non valido", http.StatusBadRequest)
		return
	}

	// Verifica che il menu fisso esista
	exists, err := h.repo.Exists(ctx, id)
	if err != nil || !exists {
		http.Error(w, "Menu fisso non trovato", http.StatusNotFound)
		return
	}

	// Tenta di recuperare la composizione dalla cache
	pietanze, found, err := h.cache.GetComposizione(ctx, id)
	if err != nil {
		log.Printf("Errore nell'accesso alla cache: %v", err)
		// Continua con il database in caso di errore della cache
	} else if found {
		// Usa i dati dalla cache
		log.Printf("Servendo la composizione del menu fisso ID %d dalla cache Redis", id)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pietanze)
		return
	}

	// Cache miss o errore, recupera dal database
	pietanze, err = h.repo.GetComposizione(ctx, id)
	if err != nil {
		http.Error(w, "Errore nel recupero della composizione del menu fisso", http.StatusInternalServerError)
		log.Printf("Errore nel recupero della composizione del menu fisso: %v", err)
		return
	}

	// Salva in cache per le future richieste
	if err := h.cache.SetComposizione(ctx, id, pietanze); err != nil {
		log.Printf("Errore nell'aggiornamento della cache: %v", err)
		// Continua comunque
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pietanze)
}

// AddPietanzaToMenu aggiunge una pietanza a un menu fisso
func (h *MenuFissoHandler) AddPietanzaToMenu(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID menu non valido", http.StatusBadRequest)
		return
	}

	// Struttura per la richiesta
	var requestBody struct {
		IDPietanza int `json:"id_pietanza"`
	}

	// Decodifica il JSON della richiesta
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Formato JSON non valido", http.StatusBadRequest)
		log.Printf("Errore nella decodifica JSON: %v", err)
		return
	}

	// Validazione base
	if requestBody.IDPietanza <= 0 {
		http.Error(w, "ID pietanza deve essere maggiore di zero", http.StatusBadRequest)
		return
	}

	// Verifica che il menu fisso esista
	exists, err := h.repo.Exists(ctx, id)
	if err != nil || !exists {
		http.Error(w, "Menu fisso non trovato", http.StatusNotFound)
		return
	}

	// Aggiungi la pietanza al menu
	err = h.repo.AddPietanzaToMenu(ctx, id, requestBody.IDPietanza)
	if err != nil {
		http.Error(w, "Errore nell'aggiunta della pietanza al menu", http.StatusInternalServerError)
		log.Printf("Errore nell'aggiunta della pietanza al menu: %v", err)
		return
	}

	// Invalida la cache della composizione del menu
	if err := h.cache.InvalidateComposizione(ctx, id); err != nil {
		log.Printf("Errore nell'invalidazione della cache: %v", err)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Pietanza aggiunta al menu con successo"})
}

// RemovePietanzaFromMenu rimuove una pietanza da un menu fisso
func (h *MenuFissoHandler) RemovePietanzaFromMenu(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")
	idPietanzaStr := chi.URLParam(r, "id_pietanza")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID menu non valido", http.StatusBadRequest)
		return
	}

	idPietanza, err := strconv.Atoi(idPietanzaStr)
	if err != nil {
		http.Error(w, "ID pietanza non valido", http.StatusBadRequest)
		return
	}

	// Verifica che il menu fisso esista
	exists, err := h.repo.Exists(ctx, id)
	if err != nil || !exists {
		http.Error(w, "Menu fisso non trovato", http.StatusNotFound)
		return
	}

	// Rimuovi la pietanza dal menu
	err = h.repo.RemovePietanzaFromMenu(ctx, id, idPietanza)
	if err != nil {
		http.Error(w, "Errore nella rimozione della pietanza dal menu", http.StatusInternalServerError)
		log.Printf("Errore nella rimozione della pietanza dal menu: %v", err)
		return
	}

	// Invalida la cache della composizione del menu
	if err := h.cache.InvalidateComposizione(ctx, id); err != nil {
		log.Printf("Errore nell'invalidazione della cache: %v", err)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Pietanza rimossa dal menu con successo"})
}
