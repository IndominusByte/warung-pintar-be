package handler_http

import (
	"net/http"
	"strings"

	"github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/config"
	endpoint_http "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/endpoint/http"
	authrepo "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/repo/auth"
	cartsrepo "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/repo/carts"
	cartsusecase "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/usecase/carts"
	"github.com/creent-production/cdk-go/auth"
	"github.com/creent-production/cdk-go/filestatic"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth"
	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Server struct {
	Router *chi.Mux
	// Db config can be added here
	db       *sqlx.DB
	redisCli *redis.Pool
	cfg      *config.Config
}

func CreateNewServer(db *sqlx.DB, redisCli *redis.Pool, cfg *config.Config) *Server {
	s := &Server{db: db, redisCli: redisCli, cfg: cfg}
	s.Router = chi.NewRouter()
	return s
}

func (s *Server) MountHandlers() error {
	// jwt
	publicKey, privateKey := auth.DecodeRSA(s.cfg.JWT.PublicKey, s.cfg.JWT.PrivateKey)
	TokenAuthRS256 := jwtauth.New(s.cfg.JWT.Algorithm, privateKey, publicKey)
	s.Router.Use(jwtauth.Verifier(TokenAuthRS256))

	// middleware stack
	s.Router.Use(middleware.RealIP)
	s.Router.Use(middleware.Logger)
	s.Router.Use(middleware.Recoverer)
	s.Router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	s.Router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("doc.json"), //The url pointing to API definition
	))
	// serve file static
	fileServer := http.FileServer(filestatic.FileSystem{Static: http.Dir("static")})
	s.Router.Handle("/static/*", http.StripPrefix(strings.TrimRight("/static/", "/"), fileServer))

	// you can insert your behaviors here
	authRepo, err := authrepo.New(s.db)
	if err != nil {
		return err
	}

	cartsRepo, err := cartsrepo.New(s.db)
	if err != nil {
		return err
	}
	cartsUsecase := cartsusecase.NewCartsUsecase(cartsRepo, authRepo)
	endpoint_http.AddCarts(s.Router, cartsUsecase, s.redisCli)

	return nil
}
