package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

type RouterFunc func(r chi.Router)

func NewRouter(modules ...RouterFunc) chi.Router {
	router := chi.NewRouter()
	router.Use(recoveryMiddleware)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://localhost:*", "http://localhost:*", "https://theanuragmishra.github.io"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	for _, module := range modules {
		module(router)
	}

	// static assets
	router.Get("/static/*", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))).ServeHTTP(w, r)
	})

	// hi
	router.Get("/say-hi", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})
	return router
}
