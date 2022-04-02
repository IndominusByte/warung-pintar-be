package endpoint_http

import (
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"

	"github.com/IndominusByte/warung-pintar-be/endpoint-auth/internal/config"
	"github.com/IndominusByte/warung-pintar-be/endpoint-auth/internal/constant"
	authentity "github.com/IndominusByte/warung-pintar-be/endpoint-auth/internal/entity/auth"
	"github.com/creent-production/cdk-go/auth"
	"github.com/creent-production/cdk-go/mail"
	"github.com/creent-production/cdk-go/parser"
	"github.com/creent-production/cdk-go/response"
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
)

type authUsecaseIface interface {
	Register(ctx context.Context, rw http.ResponseWriter, payload *authentity.JsonRegisterSchema, m *mail.Mail)
	UserConfirm(ctx context.Context, rw http.ResponseWriter, token string, cfg *config.Config)
	ResendEmail(ctx context.Context, rw http.ResponseWriter, payload *authentity.JsonEmailSchema, m *mail.Mail)
	Login(ctx context.Context, rw http.ResponseWriter, payload *authentity.JsonLoginSchema, cfg *config.Config)
	FreshToken(ctx context.Context, rw http.ResponseWriter, payload *authentity.JsonPasswordOnlySchema, cfg *config.Config)
	RefreshToken(ctx context.Context, rw http.ResponseWriter, cfg *config.Config)
	AccessRevoke(ctx context.Context, rw http.ResponseWriter, redisCli *redis.Pool, cfg *config.Config)
	RefreshRevoke(ctx context.Context, rw http.ResponseWriter, redisCli *redis.Pool, cfg *config.Config)
	PasswordResetSend(ctx context.Context, rw http.ResponseWriter, payload *authentity.JsonEmailSchema, m *mail.Mail)
	PasswordReset(ctx context.Context, rw http.ResponseWriter, token string, payload *authentity.JsonPasswordResetSchema)
	UpdatePassword(ctx context.Context, rw http.ResponseWriter, payload *authentity.JsonUpdatePasswordSchema)
	UpdateAvatar(ctx context.Context, rw http.ResponseWriter, file *multipart.Form)
	UpdateAccount(ctx context.Context, rw http.ResponseWriter, payload *authentity.JsonUpdateAccountSchema)
	GetUser(ctx context.Context, rw http.ResponseWriter)
}

func AddAuth(r *chi.Mux, uc authUsecaseIface, redisCli *redis.Pool, cfg *config.Config, m *mail.Mail) {
	r.Route("/auth", func(r chi.Router) {
		// protected route
		r.Group(func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
					if err := auth.ValidateJWT(r.Context(), redisCli, "jwtFreshRequired"); err != nil {
						response.WriteJSONResponse(rw, 401, nil, map[string]interface{}{
							constant.Header: err.Error(),
						})
						return
					}
					// Token is authenticated, pass it through
					next.ServeHTTP(rw, r)
				})
			})
			r.Put("/update-password", func(rw http.ResponseWriter, r *http.Request) {
				var p authentity.JsonUpdatePasswordSchema

				if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
					response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
						constant.Body: constant.FailedParseBody,
					})
					return
				}

				uc.UpdatePassword(r.Context(), rw, &p)
			})
		})

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
			r.Get("/", func(rw http.ResponseWriter, r *http.Request) {
				uc.GetUser(r.Context(), rw)
			})
			r.Put("/update-avatar", func(rw http.ResponseWriter, r *http.Request) {
				if err := r.ParseMultipartForm(32 << 20); err != nil {
					response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
						"_body": constant.FailedParseBody,
					})
					return
				}

				uc.UpdateAvatar(r.Context(), rw, r.MultipartForm)
			})
			r.Put("/update-account", func(rw http.ResponseWriter, r *http.Request) {
				var p authentity.JsonUpdateAccountSchema

				if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
					response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
						constant.Body: constant.FailedParseBody,
					})
					return
				}

				uc.UpdateAccount(r.Context(), rw, &p)
			})
			r.Post("/fresh-token", func(rw http.ResponseWriter, r *http.Request) {
				var p authentity.JsonPasswordOnlySchema

				if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
					response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
						constant.Body: constant.FailedParseBody,
					})
					return
				}

				uc.FreshToken(r.Context(), rw, &p, cfg)
			})
			r.Delete("/access-revoke", func(rw http.ResponseWriter, r *http.Request) {
				uc.AccessRevoke(r.Context(), rw, redisCli, cfg)
			})
		})

		r.Group(func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
					if err := auth.ValidateJWT(r.Context(), redisCli, "jwtRefreshRequired"); err != nil {
						response.WriteJSONResponse(rw, 401, nil, map[string]interface{}{
							constant.Header: err.Error(),
						})
						return
					}
					// Token is authenticated, pass it through
					next.ServeHTTP(rw, r)
				})
			})
			r.Post("/refresh-token", func(rw http.ResponseWriter, r *http.Request) {
				uc.RefreshToken(r.Context(), rw, cfg)
			})
			r.Delete("/refresh-revoke", func(rw http.ResponseWriter, r *http.Request) {
				uc.RefreshRevoke(r.Context(), rw, redisCli, cfg)
			})
		})
		// public route
		r.Post("/register", func(rw http.ResponseWriter, r *http.Request) {
			var p authentity.JsonRegisterSchema

			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
					constant.Body: constant.FailedParseBody,
				})
				return
			}

			uc.Register(r.Context(), rw, &p, m)
		})
		r.Get("/confirm/{token}", func(rw http.ResponseWriter, r *http.Request) {
			token, _ := parser.ParsePathToStr("/auth/confirm/(.*)", r.URL.Path)

			uc.UserConfirm(r.Context(), rw, token, cfg)
		})
		r.Post("/resend-email", func(rw http.ResponseWriter, r *http.Request) {
			var p authentity.JsonEmailSchema

			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
					constant.Body: constant.FailedParseBody,
				})
				return
			}

			uc.ResendEmail(r.Context(), rw, &p, m)
		})
		r.Post("/login", func(rw http.ResponseWriter, r *http.Request) {
			var p authentity.JsonLoginSchema

			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
					constant.Body: constant.FailedParseBody,
				})
				return
			}

			uc.Login(r.Context(), rw, &p, cfg)
		})
		r.Post("/password-reset/send", func(rw http.ResponseWriter, r *http.Request) {
			var p authentity.JsonEmailSchema

			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
					constant.Body: constant.FailedParseBody,
				})
				return
			}

			uc.PasswordResetSend(r.Context(), rw, &p, m)
		})
		r.Put("/password-reset/{token}", func(rw http.ResponseWriter, r *http.Request) {
			token, _ := parser.ParsePathToStr("/auth/password-reset/(.*)", r.URL.Path)

			var p authentity.JsonPasswordResetSchema

			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
					constant.Body: constant.FailedParseBody,
				})
				return
			}

			uc.PasswordReset(r.Context(), rw, token, &p)
		})
	})
}
