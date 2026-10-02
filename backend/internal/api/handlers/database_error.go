package handlers

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

// Only infrastructure failures are retryable; SQL bugs keep their normal 500.
func respondDatabaseUnavailable(w http.ResponseWriter, logger *zap.Logger, err error) bool {
	var networkError net.Error
	var connectError *pgconn.ConnectError
	var postgresError *pgconn.PgError
	unavailable := errors.As(err, &networkError) || errors.As(err, &connectError) ||
		errors.Is(err, context.DeadlineExceeded) || errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, net.ErrClosed)
	if errors.As(err, &postgresError) {
		unavailable = strings.HasPrefix(postgresError.Code, "08") ||
			postgresError.Code == "57P01" || postgresError.Code == "57P02" ||
			postgresError.Code == "57P03" || postgresError.Code == "53300"
	}
	if !unavailable {
		return false
	}
	logger.Error("database unavailable for workflow trigger", zap.Error(err))
	w.Header().Set("Retry-After", "5")
	respondError(w, logger, http.StatusServiceUnavailable, "database temporarily unavailable; retry request")
	return true
}
