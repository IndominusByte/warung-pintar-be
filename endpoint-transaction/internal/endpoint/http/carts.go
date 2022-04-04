package endpoint_http

import (
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
)

type cartsUsecaseIface interface {
}

func AddCarts(r *chi.Mux, uc cartsUsecaseIface, redisCli *redis.Pool) {
	r.Route("/carts", func(r chi.Router) {
	})
}
