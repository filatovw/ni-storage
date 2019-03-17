package api

import (
	"context"
	"time"

	"net/http"

	"github.com/filatovw/ni-storage/config"
	"github.com/filatovw/ni-storage/engine"
	"github.com/filatovw/ni-storage/logger"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func New(ctx context.Context, log logger.Logger, storage engine.Storage, cfg config.Config) *http.Server {
	mux := chi.NewRouter()
	mux.Use(render.SetContentType(render.ContentTypeJSON))
	mux.Use(middleware.RequestID)
	// TODO: replace with external logger
	mux.Use(middleware.Logger)

	mux.Mount("/debug", middleware.Profiler())

	mux.Handle("/health", HealthHandler(log))
	mux.Handle("/metrics", promhttp.Handler())

	server := Server{storage: storage, log: log}

	mux.Route("/keys", func(mux chi.Router) {
		mux.Get("/", server.GetAllHandler)
		mux.Delete("/", server.DeleteAllHandler)
		mux.Route("/{id}", func(mux chi.Router) {
			mux.Get("/", server.GetHandler)
			mux.Put("/", server.SetHandler) // setting {id} in url seems more logical to me
			mux.Head("/", server.CheckHandler)
			mux.Delete("/", server.DeleteHandler)
		})
	})
	s := &http.Server{
		Addr:         cfg.HTTPServer.Address(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  30 * time.Second,
		Handler:      mux,
	}
	return s
}
