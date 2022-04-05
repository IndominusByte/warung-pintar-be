package endpoint_http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/constant"
	cartsentity "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/entity/carts"
	"github.com/creent-production/cdk-go/auth"
	"github.com/creent-production/cdk-go/response"
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
)

type cartsUsecaseIface interface {
	CreateUpdate(ctx context.Context, rw http.ResponseWriter, payload *cartsentity.JsonCreateUpdateSchema)
}

func AddCarts(r *chi.Mux, uc cartsUsecaseIface, redisCli *redis.Pool) {
	r.Route("/carts", func(r chi.Router) {
		// protected route
		r.Group(func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
					if err := auth.ValidateJWT(r.Context(), redisCli, "jwtRequired"); err != nil {
						response.WriteJSONResponse(rw, 401, nil, map[string]interface{}{
							constant.Header: err.Error(),
						})
						return
					}
					// Token is authenticated, pass it through
					next.ServeHTTP(rw, r)
				})
			})
			r.Post("/put-product", func(rw http.ResponseWriter, r *http.Request) {
				var p cartsentity.JsonCreateUpdateSchema

				if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
					response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
						constant.Body: constant.FailedParseBody,
					})
					return
				}

				uc.CreateUpdate(r.Context(), rw, &p)
			})
		})
		// public route
	})
}
