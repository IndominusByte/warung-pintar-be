package endpoint_http

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/IndominusByte/warung-pintar-be/endpoint-product/internal/constant"
	categoriesentity "github.com/IndominusByte/warung-pintar-be/endpoint-product/internal/entity/categories"
	"github.com/creent-production/cdk-go/auth"
	"github.com/creent-production/cdk-go/parser"
	"github.com/creent-production/cdk-go/response"
	"github.com/creent-production/cdk-go/validation"
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
)

type categoriesUsecaseIface interface {
	Create(ctx context.Context, rw http.ResponseWriter, payload *categoriesentity.JsonCreateUpdateSchema)
	GetAll(ctx context.Context, rw http.ResponseWriter, payload *categoriesentity.QueryParamAllCategorySchema)
	GetById(ctx context.Context, rw http.ResponseWriter, categoryId int)
	Update(ctx context.Context, rw http.ResponseWriter, payload *categoriesentity.JsonCreateUpdateSchema, categoryId int)
	Delete(ctx context.Context, rw http.ResponseWriter, categoryId int)
}

func AddCategories(r *chi.Mux, uc categoriesUsecaseIface, redisCli *redis.Pool) {
	r.Route("/categories", func(r chi.Router) {
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
				var p categoriesentity.JsonCreateUpdateSchema

				if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
					response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
						constant.Body: constant.FailedParseBody,
					})
					return
				}

				uc.Create(r.Context(), rw, &p)
			})
			r.Put("/{category_id:[1-9][0-9]*}", func(rw http.ResponseWriter, r *http.Request) {
				categoryId, _ := parser.ParsePathToInt("/categories/(.*)", r.URL.Path)

				var p categoriesentity.JsonCreateUpdateSchema

				if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
					response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
						constant.Body: constant.FailedParseBody,
					})
					return
				}

				uc.Update(r.Context(), rw, &p, categoryId)
			})
			r.Delete("/{category_id:[1-9][0-9]*}", func(rw http.ResponseWriter, r *http.Request) {
				categoryId, _ := parser.ParsePathToInt("/categories/(.*)", r.URL.Path)

				uc.Delete(r.Context(), rw, categoryId)
			})
		})

		// public route
		r.Get("/", func(rw http.ResponseWriter, r *http.Request) {
			var p categoriesentity.QueryParamAllCategorySchema

			if err := validation.ParseRequest(&p, r.URL.Query()); err != nil {
				response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
					constant.Body: constant.FailedParseBody,
				})
				return
			}

			uc.GetAll(r.Context(), rw, &p)
		})

		r.Get("/{category_id:[1-9][0-9]*}", func(rw http.ResponseWriter, r *http.Request) {
			categoryId, _ := parser.ParsePathToInt("/categories/(.*)", r.URL.Path)

			uc.GetById(r.Context(), rw, categoryId)
		})
	})
}
