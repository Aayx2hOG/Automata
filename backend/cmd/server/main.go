package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Aayx2hOG/automata/internal/api"
	"github.com/Aayx2hOG/automata/internal/config"
	"github.com/Aayx2hOG/automata/internal/database"
	"github.com/joho/godotenv"
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

	ctx := context.Background()

	pool, err := database.NewPostgresPool(ctx, cfg.Database.URL, cfg.Database.MaxConns)
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer pool.Close()
	fmt.Println("Connected to Postgres")

	srv := api.NewServer(cfg.HTTP.Port, cfg.HTTP.ShutdownTimeout, pool)

	go func() {
		if err := srv.StartServer(); err != nil {
			log.Println(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := srv.ShutdownServer(shutdownCtx); err != nil {
		log.Fatalf("Forced shutdown: %v", err)
	}
	log.Println("Server exited cleanly")
}
