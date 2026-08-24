package api

import (
	"net/http"
	"time"

	"github.com/Aayx2hOG/automata/internal/api/handlers"
	appAuth "github.com/Aayx2hOG/automata/internal/auth"
	appMiddleware "github.com/Aayx2hOG/automata/internal/middleware"
	"github.com/Aayx2hOG/automata/internal/repositories"
	"github.com/Aayx2hOG/automata/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func NewRouter(pool *pgxpool.Pool, logger *zap.Logger, allowedOrigins []string, authService *services.AuthService, workflowService *services.WorkflowService, userRepo repositories.UserRepository, jwtManager *appAuth.JWTManager) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.ClientIPFromRemoteAddr)
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

	authHandler := handlers.NewAuthHandler(authService, logger)
	authLimiter := appMiddleware.NewRateLimiter(0.5, 5)
	r.Route("/auth", func(r chi.Router) {
		r.Use(authLimiter.Middleware)
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
		r.Post("/logout", authHandler.Logout)
	})

	userHandler := handlers.NewUserHandler(userRepo, logger)
	r.Route("/users", func(r chi.Router) {
		r.Use(appMiddleware.RequireAuth(jwtManager))
		r.Get("/me", userHandler.Me)
	})

	workflowHandler := handlers.NewWorkflowHandler(workflowService, logger)
	r.Route("/workflows", func(r chi.Router) {
		r.Use(appMiddleware.RequireAuth(jwtManager))
		r.Post("/", workflowHandler.Create)
		r.Get("/", workflowHandler.List)
		r.Get("/{id}", workflowHandler.Get)
		r.Post("/{id}/run", workflowHandler.Run)
	})
	return r
}
