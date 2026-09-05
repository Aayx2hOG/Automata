package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	appAuth "github.com/Aayx2hOG/automata/internal/auth"
	"github.com/Aayx2hOG/automata/internal/repositories"
	"github.com/Aayx2hOG/automata/internal/scheduler"
	"github.com/Aayx2hOG/automata/internal/services"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Server struct {
	httpServer *http.Server
	logger     *zap.Logger
}

func NewServer(port int, shutdownTimeout time.Duration, pool *pgxpool.Pool, logger *zap.Logger, allowedOrigins []string, authService *services.AuthService, workflowService *services.WorkflowService, scheduleRepo repositories.ScheduleRepository, sched *scheduler.Scheduler, userRepo repositories.UserRepository, jwtManager *appAuth.JWTManager) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      NewRouter(pool, logger, allowedOrigins, authService, workflowService, scheduleRepo, sched, userRepo, jwtManager),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		logger: logger,
	}
}

func (s *Server) StartServer() error {
	s.logger.Info("server starting", zap.String("addr", s.httpServer.Addr))
	return s.httpServer.ListenAndServe()
}

func (s *Server) ShutdownServer(ctx context.Context) error {
	s.logger.Info("Shutting down the server...")
	return s.httpServer.Shutdown(ctx)
}
