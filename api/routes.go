package api

import (
	"ristorante-api/api/handlers"
	"ristorante-api/cache"
	"ristorante-api/database"
	"ristorante-api/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func SetupRoutes(db *database.DB) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Monitoring
	monitoringHandler := handlers.NewMonitoringHandler(db)

	// Ristoranti
	ristoranteRepo := repository.NewRistoranteRepository(db.Pool)
	ristoranteCache := cache.NewRistoranteCache(db.Redis.Client)
	ristoranteHandler := handlers.NewRistoranteHandler(ristoranteRepo, ristoranteCache)

	// Tavoli
	tavoloRepo := repository.NewTavoloRepository(db.Pool)
	tavoloCache := cache.NewTavoloCache(db.Redis.Client)
	tavoloHandler := handlers.NewTavoloHandler(tavoloRepo, tavoloCache)

	// Ordini
	ordineRepo := repository.NewOrdineRepository(db.Pool)
	ordineCache := cache.NewOrdineCache(db.Redis.Client)
	ordineHandler := handlers.NewOrdineHandler(ordineRepo, ordineCache)

	// Monitoring Routes
	r.Route("/monitoring", func(r chi.Router) {
		r.Get("/redis", monitoringHandler.GetRedisStatus)
	})

	// API Routes
	r.Route("/api", func(r chi.Router) {

		r.Route("/ristoranti", func(r chi.Router) {
			r.Get("/", ristoranteHandler.GetRistoranti)
			r.Post("/", ristoranteHandler.CreateRistorante)
			r.Get("/{id}", ristoranteHandler.GetRistorante)
			r.Put("/{id}", ristoranteHandler.UpdateRistorante)
			r.Delete("/{id}", ristoranteHandler.DeleteRistorante)
		})

		r.Route("/tavoli", func(r chi.Router) {
			r.Get("/", tavoloHandler.GetTavoli)
			r.Get("/{id}", tavoloHandler.GetTavolo)
			r.Post("/", tavoloHandler.CreateTavolo)
			r.Put("/{id}", tavoloHandler.UpdateTavolo)
			r.Delete("/{id}", tavoloHandler.DeleteTavolo)
			r.Patch("/{id}/stato", tavoloHandler.CambiaStatoTavolo)
		})

		r.Route("/ordini", func(r chi.Router) {
			r.Get("/", ordineHandler.GetOrdini)
			r.Get("/{id}", ordineHandler.GetOrdine)
			r.Post("/", ordineHandler.CreateOrdine)
			r.Patch("/{id}", ordineHandler.UpdateStatoOrdine)
			r.Delete("/{id}", ordineHandler.DeleteOrdine)
		})

	})

	return r
}
