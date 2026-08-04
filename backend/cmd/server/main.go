package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Aayx2hOG/automata/internal/api"
	"github.com/Aayx2hOG/automata/internal/config"
	"github.com/Aayx2hOG/automata/internal/database"
	"github.com/Aayx2hOG/automata/internal/telemetry"
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
		logger.Fatal("database error: %v", zap.Error(err))
	}
	defer pool.Close()
	logger.Info("Connected to Postgres")

	srv := api.NewServer(cfg.HTTP.Port, cfg.HTTP.ShutdownTimeout, pool, logger, cfg.HTTP.AllowedOriginsList())

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
		logger.Fatal("Forced shutdown: %v", zap.Error(err))
	}
	logger.Info("server exited cleanly")
}
