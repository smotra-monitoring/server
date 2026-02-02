package handlers

import (
	"context"

	healthAPI "github.com/smotra-monitoring/server/internal/api/health"
	api "github.com/smotra-monitoring/server/internal/api/v1"
	"github.com/smotra-monitoring/server/internal/config"
	"github.com/smotra-monitoring/server/internal/database"
	"github.com/smotra-monitoring/server/internal/handlers/agent_claim"
	"github.com/smotra-monitoring/server/internal/handlers/agent_claim_status"
	"github.com/smotra-monitoring/server/internal/handlers/agent_configuration"
	"github.com/smotra-monitoring/server/internal/handlers/agent_register"
	"github.com/smotra-monitoring/server/internal/handlers/metrics"
	"github.com/smotra-monitoring/server/internal/logger"
)

// CombinedHandler combines all handler implementations
type CombinedHandler struct {
	metrics             *metrics.Handler
	agent_configuration *agent_configuration.Handler
	agent_register      *agent_register.Handler
	agent_claim_status  *agent_claim_status.Handler
	agent_claim         *agent_claim.Handler
}

// NewCombinedHandler creates a new combined handler
func NewCombinedHandler(logger *logger.Logger, db database.Database, cfg *config.Config, appVersion string, metricsHandler *metrics.Handler) *CombinedHandler {
	configHandler := agent_configuration.NewHandler(logger, db, appVersion)
	registerHandler := agent_register.NewHandler(logger, db, cfg)
	claimStatusHandler := agent_claim_status.NewHandler(logger, db)
	claimHandler := agent_claim.NewHandler(logger, db)

	// Register handlers as metrics providers
	metricsHandler.RegisterMetricsProvider(configHandler)
	metricsHandler.RegisterMetricsProvider(registerHandler)
	metricsHandler.RegisterMetricsProvider(claimStatusHandler)
	metricsHandler.RegisterMetricsProvider(claimHandler)

	// Note: Claim-related handlers use string metrics, not metrics provider interface
	// Their metrics are exposed through a different mechanism

	return &CombinedHandler{
		metrics:             metricsHandler,
		agent_configuration: configHandler,
		agent_register:      registerHandler,
		agent_claim_status:  claimStatusHandler,
		agent_claim:         claimHandler,
	}
}

// PrometheusMetrics delegates to metrics handler
func (h *CombinedHandler) PrometheusMetrics(ctx context.Context, request healthAPI.PrometheusMetricsRequestObject) (healthAPI.PrometheusMetricsResponseObject, error) {
	return h.metrics.PrometheusMetrics(ctx, request)
}

// GetAgentConfiguration delegates to configuration handler
func (h *CombinedHandler) GetAgentConfiguration(ctx context.Context, request api.GetAgentConfigurationRequestObject) (api.GetAgentConfigurationResponseObject, error) {
	return h.agent_configuration.GetAgentConfiguration(ctx, request)
}

// RegisterAgentSelf delegates to agent register handler
func (h *CombinedHandler) RegisterAgentSelf(ctx context.Context, request api.RegisterAgentSelfRequestObject) (api.RegisterAgentSelfResponseObject, error) {
	return h.agent_register.Handle(ctx, request)
}

// GetAgentClaimStatus delegates to agent claim status handler
func (h *CombinedHandler) GetAgentClaimStatus(ctx context.Context, request api.GetAgentClaimStatusRequestObject) (api.GetAgentClaimStatusResponseObject, error) {
	return h.agent_claim_status.Handle(ctx, request)
}

// ClaimAgent delegates to agent claim handler
func (h *CombinedHandler) ClaimAgent(ctx context.Context, request api.ClaimAgentRequestObject) (api.ClaimAgentResponseObject, error) {
	return h.agent_claim.Handle(ctx, request)
}
