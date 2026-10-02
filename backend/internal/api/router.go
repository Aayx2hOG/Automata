package api

import (
	"net/http"
	"time"

	"github.com/Aayx2hOG/automata/internal/api/handlers"
	appAuth "github.com/Aayx2hOG/automata/internal/auth"
	appMiddleware "github.com/Aayx2hOG/automata/internal/middleware"
	"github.com/Aayx2hOG/automata/internal/repositories"
	"github.com/Aayx2hOG/automata/internal/scheduler"
	"github.com/Aayx2hOG/automata/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func NewRouter(
	pool *pgxpool.Pool,
	logger *zap.Logger,
	allowedOrigins []string,
	authService *services.AuthService,
	workflowService *services.WorkflowService,
	scheduleRepo repositories.ScheduleRepository,
	sched *scheduler.Scheduler,
	userRepo repositories.UserRepository,
	jwtManager *appAuth.JWTManager,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(appMiddleware.RequestLogger(logger))
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Authorization", "X-Webhook-Timestamp", "X-Webhook-Signature"},
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
		r.Put("/{id}", workflowHandler.Update)
		r.Post("/{id}/run", workflowHandler.Run)
	})

	runHandler := handlers.NewRunHandler(workflowService, logger)
	r.Route("/runs", func(r chi.Router) {
		r.Use(appMiddleware.RequireAuth(jwtManager))
		r.Get("/{id}", runHandler.Get)
		r.Get("/", runHandler.List)
	})

	scheduleHandler := handlers.NewScheduleHandler(workflowService, scheduleRepo, sched, logger)
	r.Route("/workflows/{id}/schedules", func(r chi.Router) {
		r.Use(appMiddleware.RequireAuth(jwtManager))
		r.Post("/", scheduleHandler.Create)
		r.Get("/", scheduleHandler.List)
	})
	r.Route("/schedules", func(r chi.Router) {
		r.Use(appMiddleware.RequireAuth(jwtManager))
		r.Delete("/{scheduleID}", scheduleHandler.Deactivate)
	})

	webhookHandler := handlers.NewWebhookHandler(workflowService, logger)
	webhookLimiter := appMiddleware.NewRateLimiter(5, 20)
	r.Route("/webhook", func(r chi.Router) {

		r.Use(webhookLimiter.Middleware)
		r.Post("/{workflowID}", webhookHandler.Trigger)
	})

	return r
}
