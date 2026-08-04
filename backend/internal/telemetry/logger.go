package telemetry

import "go.uber.org/zap"

func NewLogger(env string) (*zap.Logger, error) {
	if env == "development" {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}
