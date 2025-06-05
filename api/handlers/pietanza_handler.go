package handlers

// import (
// 	"encoding/json"
// 	"log"
// 	"net/http"
// 	"ristorante-api/database"
// 	"ristorante-api/models"
// 	"strconv"

// 	"github.com/go-chi/chi/v5"
// )

// // PietanzaHandler gestisce le richieste relative alle pietanze
// type PietanzaHandler struct {
// 	DB *database.DB
// }

// // NewPietanzaHandler crea un nuovo handler per le pietanze
// func NewPietanzaHandler(db *database.DB) *PietanzaHandler {
// 	return &PietanzaHandler{
// 		DB: db,
// 	}
// }

// // GetPietanze restituisce tutte le pietanze disponibili
// func (h *PietanzaHandler) GetPietanze(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	pietanze := []models.Pietanza{}

// 	// Tenta di recuperare le pietanze dalla cache
// 	cached, found, err := h.DB.Redis.GetPietanzeCache(ctx)
// 	if err != nil {
// 		log.Printf("Errore nell'accesso alla cache: %v", err)
// 		// Continua con il database in caso di errore della cache
// 	} else if found {
// 		// Usa i dati dalla cache
// 		log.Println("Servendo le pietanze dalla cache Redis")
// 		w.Header().Set("Content-Type", "application/json")
// 		json.NewEncoder(w).Encode(cached)
// 		return
// 	}

// 	// Cache miss o errore, recupera dal database
// 	log.Println("Cache miss, recupero le pietanze dal database")
// 	rows, err := h.DB.Pool.Query(ctx, `
// 		SELECT p.id_pietanza, p.nome, p.prezzo, p.id_categoria, p.disponibile
// 		FROM pietanza p
// 	`)
// 	if err != nil {
// 		http.Error(w, "Errore nel recupero delle pietanze", http.StatusInternalServerError)
// 		log.Printf("Errore nel recupero delle pietanze: %v", err)
// 		return
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var p models.Pietanza
// 		err := rows.Scan(&p.ID, &p.Nome, &p.Prezzo, &p.IDCategoria, &p.Disponibile)
// 		if err != nil {
// 			http.Error(w, "Errore nella lettura delle pietanze", http.StatusInternalServerError)
// 			log.Printf("Errore nella lettura delle pietanze: %v", err)
// 			return
// 		}
// 		pietanze = append(pietanze, p)
// 	}

// 	if err := rows.Err(); err != nil {
// 		http.Error(w, "Errore nell'iterazione sui risultati", http.StatusInternalServerError)
// 		log.Printf("Errore nell'iterazione sui risultati: %v", err)
// 		return
// 	}

// 	// Salva i risultati in cache per le future richieste
// 	if err := h.DB.Redis.SetPietanzeCache(ctx, pietanze); err != nil {
// 		log.Printf("Errore nell'aggiornamento della cache: %v", err)
// 		// Continua comunque
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(pietanze)
// }

// // GetPietanza restituisce una singola pietanza per ID
// func (h *PietanzaHandler) GetPietanza(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	idStr := chi.URLParam(r, "id")
// 	id, err := strconv.Atoi(idStr)
// 	if err != nil {
// 		http.Error(w, "ID non valido", http.StatusBadRequest)
// 		return
// 	}

// 	// Tenta di recuperare la pietanza dalla cache
// 	pietanza, found, err := h.DB.Redis.GetPietanzaCache(ctx, id)
// 	if err != nil {
// 		log.Printf("Errore nell'accesso alla cache: %v", err)
// 		// Continua con il database in caso di errore della cache
// 	} else if found {
// 		// Usa i dati dalla cache
// 		log.Printf("Servendo la pietanza ID %d dalla cache Redis", id)
// 		w.Header().Set("Content-Type", "application/json")
// 		json.NewEncoder(w).Encode(pietanza)
// 		return
// 	}

// 	// Cache miss o errore, recupera dal database
// 	var p models.Pietanza
// 	err = h.DB.Pool.QueryRow(ctx, `
// 		SELECT p.id_pietanza, p.nome, p.prezzo, p.id_categoria, p.disponibile
// 		FROM pietanza p
// 		WHERE p.id_pietanza = $1
// 	`, id).Scan(&p.ID, &p.Nome, &p.Prezzo, &p.IDCategoria, &p.Disponibile)

// 	if err != nil {
// 		http.Error(w, "Pietanza non trovata", http.StatusNotFound)
// 		log.Printf("Errore nel recupero della pietanza: %v", err)
// 		return
// 	}

// 	// Salva in cache per le future richieste
// 	if err := h.DB.Redis.SetPietanzaCache(ctx, id, p); err != nil {
// 		log.Printf("Errore nell'aggiornamento della cache: %v", err)
// 		// Continua comunque
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(p)
// }

// // CreatePietanza crea una nuova pietanza
// func (h *PietanzaHandler) CreatePietanza(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	var p models.Pietanza

// 	// Decodifica il JSON della richiesta
// 	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
// 		http.Error(w, "Formato JSON non valido", http.StatusBadRequest)
// 		log.Printf("Errore nella decodifica JSON: %v", err)
// 		return
// 	}

// 	// Validazione base
// 	if p.Nome == "" || p.Prezzo <= 0 {
// 		http.Error(w, "Nome e prezzo sono campi obbligatori", http.StatusBadRequest)
// 		return
// 	}

// 	// Inserimento nel database
// 	err := h.DB.Pool.QueryRow(ctx, `
// 		INSERT INTO pietanza (nome, prezzo, id_categoria, disponibile)
// 		VALUES ($1, $2, $3, $4)
// 		RETURNING id_pietanza
// 	`, p.Nome, p.Prezzo, p.IDCategoria, p.Disponibile).Scan(&p.ID)

// 	if err != nil {
// 		http.Error(w, "Errore nella creazione della pietanza", http.StatusInternalServerError)
// 		log.Printf("Errore nella creazione della pietanza: %v", err)
// 		return
// 	}

// 	// Invalida la cache delle pietanze
// 	if err := h.DB.Redis.InvalidatePietanzeCache(ctx); err != nil {
// 		log.Printf("Errore nell'invalidazione della cache: %v", err)
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(p)
// }

// // UpdatePietanza aggiorna una pietanza esistente
// func (h *PietanzaHandler) UpdatePietanza(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	idStr := chi.URLParam(r, "id")
// 	id, err := strconv.Atoi(idStr)
// 	if err != nil {
// 		http.Error(w, "ID non valido", http.StatusBadRequest)
// 		return
// 	}

// 	var p models.Pietanza

// 	// Decodifica il JSON della richiesta
// 	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
// 		http.Error(w, "Formato JSON non valido", http.StatusBadRequest)
// 		log.Printf("Errore nella decodifica JSON: %v", err)
// 		return
// 	}

// 	// Validazione base
// 	if p.Nome == "" || p.Prezzo <= 0 {
// 		http.Error(w, "Nome e prezzo sono campi obbligatori", http.StatusBadRequest)
// 		return
// 	}

// 	// Verifica che la pietanza esista
// 	var exists bool
// 	err = h.DB.Pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pietanza WHERE id_pietanza = $1)", id).Scan(&exists)
// 	if err != nil || !exists {
// 		http.Error(w, "Pietanza non trovata", http.StatusNotFound)
// 		return
// 	}

// 	// Aggiornamento nel database
// 	_, err = h.DB.Pool.Exec(ctx, `
// 		UPDATE pietanza
// 		SET nome = $1, prezzo = $2, id_categoria = $3, disponibile = $4
// 		WHERE id_pietanza = $5
// 	`, p.Nome, p.Prezzo, p.IDCategoria, p.Disponibile, id)

// 	if err != nil {
// 		http.Error(w, "Errore nell'aggiornamento della pietanza", http.StatusInternalServerError)
// 		log.Printf("Errore nell'aggiornamento della pietanza: %v", err)
// 		return
// 	}

// 	// Invalida la cache per questa pietanza e per l'elenco completo
// 	if err := h.DB.Redis.InvalidatePietanzaCache(ctx, id); err != nil {
// 		log.Printf("Errore nell'invalidazione della cache pietanza: %v", err)
// 	}
// 	if err := h.DB.Redis.InvalidatePietanzeCache(ctx); err != nil {
// 		log.Printf("Errore nell'invalidazione della cache pietanze: %v", err)
// 	}

// 	// Imposta l'ID nella risposta
// 	p.ID = id

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(p)
// }

// // DeletePietanza elimina una pietanza
// func (h *PietanzaHandler) DeletePietanza(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()
// 	idStr := chi.URLParam(r, "id")
// 	id, err := strconv.Atoi(idStr)
// 	if err != nil {
// 		http.Error(w, "ID non valido", http.StatusBadRequest)
// 		return
// 	}

// 	// Verifica che la pietanza esista
// 	var exists bool
// 	err = h.DB.Pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pietanza WHERE id_pietanza = $1)", id).Scan(&exists)
// 	if err != nil || !exists {
// 		http.Error(w, "Pietanza non trovata", http.StatusNotFound)
// 		return
// 	}

// 	// Eliminazione dal database
// 	_, err = h.DB.Pool.Exec(ctx, "DELETE FROM pietanza WHERE id_pietanza = $1", id)
// 	if err != nil {
// 		http.Error(w, "Errore nell'eliminazione della pietanza", http.StatusInternalServerError)
// 		log.Printf("Errore nell'eliminazione della pietanza: %v", err)
// 		return
// 	}

// 	// Invalida la cache per questa pietanza e per l'elenco completo
// 	if err := h.DB.Redis.InvalidatePietanzaCache(ctx, id); err != nil {
// 		log.Printf("Errore nell'invalidazione della cache pietanza: %v", err)
// 	}
// 	if err := h.DB.Redis.InvalidatePietanzeCache(ctx); err != nil {
// 		log.Printf("Errore nell'invalidazione della cache pietanze: %v", err)
// 	}

// 	// Risposta 204 No Content
// 	w.WriteHeader(http.StatusNoContent)
// }

// // HealthCheck verifica lo stato dell'API
// func (h *PietanzaHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
// 	type HealthResponse struct {
// 		Status  string `json:"status"`
// 		Message string `json:"message"`
// 	}

// 	// Verifica connessione al database
// 	err := h.DB.Pool.Ping(r.Context())
// 	if err != nil {
// 		response := HealthResponse{
// 			Status:  "error",
// 			Message: "Database connection failed",
// 		}
// 		w.Header().Set("Content-Type", "application/json")
// 		w.WriteHeader(http.StatusServiceUnavailable)
// 		json.NewEncoder(w).Encode(response)
// 		return
// 	}

// 	// Risposta OK
// 	response := HealthResponse{
// 		Status:  "ok",
// 		Message: "Service is healthy",
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(response)
// }
