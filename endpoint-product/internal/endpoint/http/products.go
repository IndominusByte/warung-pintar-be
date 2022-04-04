package endpoint_http

import (
	"context"
	"mime/multipart"
	"net/http"

	"github.com/IndominusByte/warung-pintar-be/endpoint-product/internal/constant"
	productsentity "github.com/IndominusByte/warung-pintar-be/endpoint-product/internal/entity/products"
	"github.com/creent-production/cdk-go/auth"
	"github.com/creent-production/cdk-go/parser"
	"github.com/creent-production/cdk-go/response"
	"github.com/creent-production/cdk-go/validation"
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
)

type productsUsecaseIface interface {
	Create(ctx context.Context, rw http.ResponseWriter, file *multipart.Form, payload *productsentity.FormCreateUpdateSchema)
	GetAll(ctx context.Context, rw http.ResponseWriter, payload *productsentity.QueryParamAllProductSchema)
	Update(ctx context.Context, rw http.ResponseWriter, file *multipart.Form, payload *productsentity.FormCreateUpdateSchema, productId int)
	Delete(ctx context.Context, rw http.ResponseWriter, productId int)
	GetBySlug(ctx context.Context, rw http.ResponseWriter, slug string)
}

func AddProducts(r *chi.Mux, uc productsUsecaseIface, redisCli *redis.Pool) {
	r.Route("/products", func(r chi.Router) {
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

				var p productsentity.FormCreateUpdateSchema

				if err := validation.ParseRequest(&p, r.Form); err != nil {
					response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
						"_body": constant.FailedParseBody,
					})
					return
				}

				uc.Create(r.Context(), rw, r.MultipartForm, &p)
			})
			r.Put("/{product_id:[1-9][0-9]*}", func(rw http.ResponseWriter, r *http.Request) {
				productId, _ := parser.ParsePathToInt("/products/(.*)", r.URL.Path)

				if err := r.ParseMultipartForm(32 << 20); err != nil {
					response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
						"_body": constant.FailedParseBody,
					})
					return
				}

				var p productsentity.FormCreateUpdateSchema

				if err := validation.ParseRequest(&p, r.Form); err != nil {
					response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
						"_body": constant.FailedParseBody,
					})
					return
				}

				uc.Update(r.Context(), rw, r.MultipartForm, &p, productId)
			})
			r.Delete("/{product_id:[1-9][0-9]*}", func(rw http.ResponseWriter, r *http.Request) {
				productId, _ := parser.ParsePathToInt("/products/(.*)", r.URL.Path)

				uc.Delete(r.Context(), rw, productId)
			})
		})

		// public route
		r.Get("/{slug}", func(rw http.ResponseWriter, r *http.Request) {
			slug, _ := parser.ParsePathToStr("/products/(.*)", r.URL.Path)

			uc.GetBySlug(r.Context(), rw, slug)
		})
		r.Get("/", func(rw http.ResponseWriter, r *http.Request) {
			var p productsentity.QueryParamAllProductSchema

			if err := validation.ParseRequest(&p, r.URL.Query()); err != nil {
				response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
					constant.Body: constant.FailedParseBody,
				})
				return
			}

			uc.GetAll(r.Context(), rw, &p)
		})
	})
}
