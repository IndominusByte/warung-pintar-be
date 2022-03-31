package endpoint_http

import (
	"github.com/IndominusByte/warung-pintar-be/endpoint-auth/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
)

type authUsecaseIface interface {
}

func AddAuth(r *chi.Mux, uc authUsecaseIface, redisCli *redis.Pool, cfg *config.Config) {
	r.Route("/auth", func(r chi.Router) {
		// protected route
		// public route
	})
}
