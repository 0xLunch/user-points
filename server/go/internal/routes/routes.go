package routes

import (
	"github.com/0xlunch/user-service/internal/db"
	"github.com/0xlunch/user-service/internal/handlers"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes(r *chi.Mux, db *db.DB) {

	h := handlers.NewHandlers(db)

	// User routes
	r.Route("/user", func(r chi.Router) {
		r.Post("/register", h.RegisterHandler)
		r.Post("/login", h.LoginHandler)
		r.Route("/points", func(r chi.Router) {
			r.Get("/", h.GetPointsHandler)
			r.Post("/", h.UpdatePointsHandler)
		})
	})
}
