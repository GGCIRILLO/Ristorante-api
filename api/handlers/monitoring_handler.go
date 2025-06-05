package handlers

import (
	"encoding/json"
	"net/http"
	"ristorante-api/database"
)

// MonitoringHandler gestisce le richieste di monitoraggio
type MonitoringHandler struct {
	DB *database.DB
}

// NewMonitoringHandler crea un nuovo handler per il monitoraggio
func NewMonitoringHandler(db *database.DB) *MonitoringHandler {
	return &MonitoringHandler{
		DB: db,
	}
}

// RedisStatus rappresenta lo stato di Redis
type RedisStatus struct {
	Status                   string  `json:"status"`
	KeysCount                int64   `json:"keys_count"`
	UsedMemoryHuman          string  `json:"used_memory_human"`
	TotalConnectionsReceived int64   `json:"total_connections_received"`
	HitRate                  float64 `json:"hit_rate,omitempty"`
	MissRate                 float64 `json:"miss_rate,omitempty"`
}

// GetRedisStatus restituisce lo stato attuale di Redis
func (h *MonitoringHandler) GetRedisStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, err := h.DB.Redis.Client.Info(ctx).Result()
	if err != nil {
		http.Error(w, "Errore nel recupero delle informazioni Redis", http.StatusInternalServerError)
		return
	}

	// Ottieni il numero di chiavi (o stima) nel database
	keyCount, err := h.DB.Redis.Client.DBSize(ctx).Result()
	if err != nil {
		http.Error(w, "Errore nel recupero delle dimensioni del database Redis", http.StatusInternalServerError)
		return
	}

	// Ottieni statistiche sulla memoria
	_, err = h.DB.Redis.Client.Info(ctx, "memory").Result()
	if err != nil {
		http.Error(w, "Errore nel recupero delle informazioni sulla memoria Redis", http.StatusInternalServerError)
		return
	}

	// Ottieni statistiche sulle connessioni
	_, err = h.DB.Redis.Client.Info(ctx, "stats").Result()
	if err != nil {
		http.Error(w, "Errore nel recupero delle statistiche Redis", http.StatusInternalServerError)
		return
	}

	// Per semplicità, estraiamo solo alcune informazioni di base
	// In un'applicazione reale, si potrebbe fare un parsing più dettagliato
	status := RedisStatus{
		Status:                   "online",
		KeysCount:                keyCount,
		UsedMemoryHuman:          "N/A",
		TotalConnectionsReceived: 0,
	}

	// Calcola le statistiche di hit/miss se disponibili
	hitCount := h.DB.Redis.GetHitCount()
	missCount := h.DB.Redis.GetMissCount()
	totalReqs := hitCount + missCount

	if totalReqs > 0 {
		status.HitRate = float64(hitCount) / float64(totalReqs) * 100
		status.MissRate = float64(missCount) / float64(totalReqs) * 100
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
