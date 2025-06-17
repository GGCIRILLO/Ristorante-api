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

	// Cache
	ingredienteCache := cache.NewIngredienteCache(db.Redis.Client)
	pietanzaCache := cache.NewPietanzaCache(db.Redis.Client)
	ricettaCache := cache.NewRicettaCache(db.Redis.Client)
	menuFissoCache := cache.NewMenuFissoCache(db.Redis.Client)

	// Pietanze
	pietanzaRepo := repository.NewPietanzaRepository(db.Pool)
	ricettaRepo := repository.NewRicettaRepository(db.Pool, ricettaCache)

	// Menu Fissi
	menuFissoRepo := repository.NewMenuFissoRepository(db.Pool)
	menuFissoHandler := handlers.NewMenuFissoHandler(menuFissoRepo, menuFissoCache)

	// Pietanza Handler
	pietanzaHandler := handlers.NewPietanzaHandler(pietanzaRepo, pietanzaCache, ricettaRepo, menuFissoRepo, ingredienteCache)

	// Ingredienti
	ingredienteRepo := repository.NewIngredienteRepository(db.Pool)
	ingredienteHandler := handlers.NewIngredienteHandler(ingredienteRepo, ingredienteCache)
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
			r.Get("/liberi", tavoloHandler.GetTavoliLiberi)
			r.Get("/occupati", tavoloHandler.GetTavoliOccupati)
		})

		r.Route("/ordini", func(r chi.Router) {
			r.Get("/", ordineHandler.GetOrdini)
			r.Get("/completi", ordineHandler.GetAllOrdiniCompleti)
			r.Get("/{id}", ordineHandler.GetOrdine)
			r.Get("/{id}/completo", ordineHandler.GetOrdineCompleto)
			r.Post("/", ordineHandler.CreateOrdine)
			r.Patch("/{id}", ordineHandler.UpdateStatoOrdine)
			r.Delete("/{id}", ordineHandler.DeleteOrdine)
			r.Get("/tavolo/{id_tavolo}/scontrino", ordineHandler.CalcolaScontrino)
		})

		r.Route("/pietanze", func(r chi.Router) {
			r.Get("/", pietanzaHandler.GetPietanze)
			r.Get("/{id}", pietanzaHandler.GetPietanza)
			r.Get("/{id}/ricetta", pietanzaHandler.GetRicettaByPietanzaID)
			r.Post("/", pietanzaHandler.CreatePietanza)
			r.Put("/{id}", pietanzaHandler.UpdatePietanza)
			r.Delete("/{id}", pietanzaHandler.DeletePietanza)
			r.Post("/ordine/{id_ordine}", pietanzaHandler.AddPietanzaToOrdine)
			r.Post("/menu-fisso/ordine/{id_ordine}", pietanzaHandler.AddMenuFissoToOrdine)
			r.Post("/ordine/{id_ordine}/bevanda", pietanzaHandler.AddBevandaToOrdine)
		})

		r.Route("/menu-fissi", func(r chi.Router) {
			r.Get("/", menuFissoHandler.GetMenuFissi)
			r.Get("/completi", menuFissoHandler.GetAllMenuFissiCompleti)
			r.Get("/{id}", menuFissoHandler.GetMenuFisso)
			r.Post("/", menuFissoHandler.CreateMenuFisso)
			r.Put("/{id}", menuFissoHandler.UpdateMenuFisso)
			r.Delete("/{id}", menuFissoHandler.DeleteMenuFisso)
			r.Get("/{id}/composizione", menuFissoHandler.GetComposizione)
			r.Get("/{id}/completo", menuFissoHandler.GetMenuFissoCompleto)
			r.Post("/{id}/pietanza", menuFissoHandler.AddPietanzaToMenu)
			r.Delete("/{id}/pietanza/{id_pietanza}", menuFissoHandler.RemovePietanzaFromMenu)
		})

		r.Route("/ingredienti", func(r chi.Router) {
			r.Get("/", ingredienteHandler.GetIngredienti)
			r.Get("/{id}", ingredienteHandler.GetIngredienteByID)
			r.Post("/", ingredienteHandler.CreateIngrediente)
			r.Put("/{id}", ingredienteHandler.UpdateIngrediente)
			r.Delete("/{id}", ingredienteHandler.DeleteIngrediente)
			r.Get("/da-riordinare", ingredienteHandler.GetIngredientiDaRiordinare)
			r.Post("/{id}/rifornisci", ingredienteHandler.RifornisciIngrediente)
		})

	})

	return r
}
