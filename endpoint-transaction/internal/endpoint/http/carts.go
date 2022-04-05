package endpoint_http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/constant"
	cartsentity "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/entity/carts"
	"github.com/creent-production/cdk-go/auth"
	"github.com/creent-production/cdk-go/response"
	"github.com/creent-production/cdk-go/validation"
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
)

type cartsUsecaseIface interface {
	CreateUpdate(ctx context.Context, rw http.ResponseWriter, payload *cartsentity.JsonCreateUpdateSchema)
	GetAll(ctx context.Context, rw http.ResponseWriter, payload *cartsentity.QueryParamAllCartSchema)
	Delete(ctx context.Context, rw http.ResponseWriter, payload *cartsentity.JsonMultipleSchema)
	MoveToPayment(ctx context.Context, rw http.ResponseWriter, redisCli *redis.Pool, payload *cartsentity.JsonMultipleSchema)
	ItemInPayment(ctx context.Context, rw http.ResponseWriter, redisCli *redis.Pool)
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
			r.Get("/", func(rw http.ResponseWriter, r *http.Request) {
				var p cartsentity.QueryParamAllCartSchema

				if err := validation.ParseRequest(&p, r.URL.Query()); err != nil {
					response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
						constant.Body: constant.FailedParseBody,
					})
					return
				}

				uc.GetAll(r.Context(), rw, &p)
			})
			r.Delete("/", func(rw http.ResponseWriter, r *http.Request) {
				var p cartsentity.JsonMultipleSchema

				if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
					response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
						constant.Body: constant.FailedParseBody,
					})
					return
				}

				uc.Delete(r.Context(), rw, &p)
			})
			r.Post("/move-to-payment", func(rw http.ResponseWriter, r *http.Request) {
				var p cartsentity.JsonMultipleSchema

				if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
					response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
						constant.Body: constant.FailedParseBody,
					})
					return
				}

				uc.MoveToPayment(r.Context(), rw, redisCli, &p)
			})
			r.Get("/item-in-payment", func(rw http.ResponseWriter, r *http.Request) {
				uc.ItemInPayment(r.Context(), rw, redisCli)
			})
		})
		// public route
	})
}
