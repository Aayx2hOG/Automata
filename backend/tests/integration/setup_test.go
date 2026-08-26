package integration

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/Aayx2hOG/automata/internal/api"
	"github.com/Aayx2hOG/automata/internal/auth"
	"github.com/Aayx2hOG/automata/internal/database"
	"github.com/Aayx2hOG/automata/internal/repositories"
	"github.com/Aayx2hOG/automata/internal/services"
	"github.com/Aayx2hOG/automata/internal/telemetry"
	"github.com/Aayx2hOG/automata/internal/workflow"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type testEnv struct {
	router http.Handler
	pool   *pgxpool.Pool
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	_ = godotenv.Load("../../.env")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set: skipping integration tests")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		t.Skip("JWT_SECRET not set or too short; skipping integration tests")
	}

	logger, err := telemetry.NewLogger("development")
	if err != nil {
		t.Fatalf("logger init: %v", err)
	}

	ctx := context.Background()
	pool, err := database.NewPostgresPool(ctx, dbURL, 5)
	if err != nil {
		t.Fatalf("database connection %v", err)
	}

	userRepo := repositories.NewUserRepository(pool)
	refreshTokenRepo := repositories.NewRefreshTokenRepository(pool)
	jwtManager := auth.NewJWTManager(jwtSecret, 15*time.Minute)
	authService := services.NewAuthService(userRepo, refreshTokenRepo, jwtManager, 168*time.Hour)

	workflowRepo := repositories.NewWorkflowRepository(pool)
	workflowVersionRepo := repositories.NewWorkflowVersionRepository(pool)
	workflowRunRepo := repositories.NewWorkflowRunRepository(pool)
	registry := workflow.NewRegistry(logger)
	engine := workflow.NewEngine(registry)
	workflowService := services.NewWorkflowService(workflowRepo, workflowVersionRepo, workflowRunRepo, engine)

	router := api.NewRouter(pool, logger, []string{"http://localhost:3000"}, authService, workflowService, userRepo, jwtManager)

	t.Cleanup(func() {
		pool.Close()
	})

	return &testEnv{router: router, pool: pool}
}
