package api

import (
	"net/http"
	"time"

	"github.com/Aayx2hOG/automata/internal/api/handlers"
	appMiddleware "github.com/Aayx2hOG/automata/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func NewRouter(pool *pgxpool.Pool, logger *zap.Logger, allowedOrigins []string) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(appMiddleware.RequestLogger(logger))
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))
	r.Get("/healthz", handlers.HealthCheck(pool, logger))
	return r
}
