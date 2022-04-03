package endpoint_http

import (
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
)

type productsUsecaseIface interface {
}

func AddProducts(r *chi.Mux, uc productsUsecaseIface, redisCli *redis.Pool) {
	r.Route("/products", func(r chi.Router) {
		// protected route
		// public route
	})
}
