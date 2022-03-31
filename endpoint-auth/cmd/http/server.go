package main

import (
	"net/http"
	"time"

	"github.com/IndominusByte/warung-pintar-be/endpoint-auth/internal/config"
	"github.com/go-chi/chi/v5"
)

func startServer(mux *chi.Mux, cfg *config.Config) error {
	readTimeout, err := time.ParseDuration(cfg.Server.HTTP.ReadTimeout)
	if err != nil {
		return err
	}
	writeTimeout, err := time.ParseDuration(cfg.Server.HTTP.WriteTimeout)
	if err != nil {
		return err
	}

	srv := http.Server{
		Addr:         cfg.Server.HTTP.Address,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		Handler:      mux,
	}

	return srv.ListenAndServe()
}
