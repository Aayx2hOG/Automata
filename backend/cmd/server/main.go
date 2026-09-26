package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Aayx2hOG/automata/internal/api"
	"github.com/Aayx2hOG/automata/internal/auth"
	"github.com/Aayx2hOG/automata/internal/config"
	"github.com/Aayx2hOG/automata/internal/database"
	"github.com/Aayx2hOG/automata/internal/repositories"
	"github.com/Aayx2hOG/automata/internal/scheduler"
	"github.com/Aayx2hOG/automata/internal/services"
	"github.com/Aayx2hOG/automata/internal/telemetry"
	"github.com/Aayx2hOG/automata/internal/workflow"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on real environment variables.")
	}
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}
	log.Printf("Automata starting in %v mode", cfg.Env)

	logger, err := telemetry.NewLogger(cfg.Env)
	if err != nil {
		log.Fatalf("logger error: %v", err)
	}
	defer logger.Sync()

	logger.Info("automata starting", zap.String("env", cfg.Env))

	ctx := context.Background()

	pool, err := database.NewPostgresPool(ctx, cfg.Database.URL, cfg.Database.MaxConns)
	if err != nil {
		logger.Fatal("database connection failed", zap.Error(err))
	}
	defer pool.Close()
	logger.Info("Connected to Postgres")

	userRepo := repositories.NewUserRepository(pool)
	refreshTokenRepo := repositories.NewRefreshTokenRepository(pool)
	jwtManager := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.AccessTokenTTL)
	authService := services.NewAuthService(userRepo, refreshTokenRepo, jwtManager, cfg.Auth.RefreshTokenTTL)

	workflowRepo := repositories.NewWorkflowRepository(pool)
	workflowVersionRepo := repositories.NewWorkflowVersionRepository(pool)
	workflowRunRepo := repositories.NewWorkflowRunRepository(pool)
	nodeRegistry := workflow.NewRegistry(logger)
	engine := workflow.NewEngine(nodeRegistry)
	workflowService := services.NewWorkflowService(workflowRepo, workflowVersionRepo, workflowRunRepo, engine, logger)

	workerCtx, cancelWorkers := context.WithCancel(context.Background())

	scheduleRepo := repositories.NewScheduleRepository(pool)
	sched := scheduler.New(scheduleRepo, workflowService, logger)
	schedulerDone := make(chan struct{})
	go func() { defer close(schedulerDone); sched.Run(workerCtx) }()

	srv := api.NewServer(cfg.HTTP.Port, cfg.HTTP.ShutdownTimeout, pool, logger, cfg.HTTP.AllowedOriginsList(), authService, workflowService, scheduleRepo, sched, userRepo, jwtManager)

	go func() {
		if err := srv.StartServer(); err != nil {
			logger.Info("server stopped", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := srv.ShutdownServer(shutdownCtx); err != nil {
		logger.Error("forced HTTP shutdown", zap.Error(err))
	}

	cancelWorkers()
	select {
	case <-schedulerDone:
	case <-shutdownCtx.Done():
		logger.Warn("durable consumer shutdown timed out")
	}

	logger.Info("server exited cleanly")
}
