package endpoint_http

import (
	"context"
	"mime/multipart"
	"net/http"

	"github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/constant"
	ordersentity "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/entity/orders"
	"github.com/creent-production/cdk-go/auth"
	"github.com/creent-production/cdk-go/parser"
	"github.com/creent-production/cdk-go/response"
	"github.com/creent-production/cdk-go/validation"
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
)

type ordersUsecaseIface interface {
	Create(ctx context.Context, rw http.ResponseWriter, redisCli *redis.Pool, file *multipart.Form, payload *ordersentity.FormCreateSchema)
	SetReject(ctx context.Context, rw http.ResponseWriter, orderId int)
	SetSuccess(ctx context.Context, rw http.ResponseWriter, orderId int)
	SetOnTheWay(ctx context.Context, rw http.ResponseWriter, orderId int, file *multipart.Form)
}

func AddOrders(r *chi.Mux, uc ordersUsecaseIface, redisCli *redis.Pool) {
	r.Route("/orders", func(r chi.Router) {
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
			r.Post("/", func(rw http.ResponseWriter, r *http.Request) {
				if err := r.ParseMultipartForm(32 << 20); err != nil {
					response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
						"_body": constant.FailedParseBody,
					})
					return
				}

				var p ordersentity.FormCreateSchema

				if err := validation.ParseRequest(&p, r.Form); err != nil {
					response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
						"_body": constant.FailedParseBody,
					})
					return
				}

				uc.Create(r.Context(), rw, redisCli, r.MultipartForm, &p)
			})
			r.Put("/set-reject/{order_id:[1-9][0-9]*}", func(rw http.ResponseWriter, r *http.Request) {
				orderId, _ := parser.ParsePathToInt("/orders/set-reject/(.*)", r.URL.Path)

				uc.SetReject(r.Context(), rw, orderId)
			})
			r.Put("/set-on-the-way/{order_id:[1-9][0-9]*}", func(rw http.ResponseWriter, r *http.Request) {
				orderId, _ := parser.ParsePathToInt("/orders/set-on-the-way/(.*)", r.URL.Path)

				if err := r.ParseMultipartForm(32 << 20); err != nil {
					response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
						"_body": constant.FailedParseBody,
					})
					return
				}

				uc.SetOnTheWay(r.Context(), rw, orderId, r.MultipartForm)
			})
			r.Put("/set-success/{order_id:[1-9][0-9]*}", func(rw http.ResponseWriter, r *http.Request) {
				orderId, _ := parser.ParsePathToInt("/orders/set-success/(.*)", r.URL.Path)

				uc.SetSuccess(r.Context(), rw, orderId)
			})
		})
		// public route
	})
}
