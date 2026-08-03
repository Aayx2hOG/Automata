package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(port int, shutdownTimeout time.Duration, pool *pgxpool.Pool) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%v", port),
			Handler:      NewRouter(pool),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}
}

func (s *Server) StartServer() error {
	log.Printf("Listening on %v", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) ShutdownServer(ctx context.Context) error {
	log.Println("Shutting down the server...")
	return s.httpServer.Shutdown(ctx)
}
