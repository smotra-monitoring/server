## Roadmap

### Completed
- [x] SQLite and PostgreSQL database support with interface abstraction
- [x] Health check endpoints (health, ready, live)
- [x] Prometheus metrics endpoint
- [x] Structured logging with slog
- [x] Configuration management (YAML/JSON)
- [x] Middleware (logging, request ID, recovery, CORS)
- [x] OpenAPI-based code generation
- [x] Unit and integration testing infrastructure
- [x] Update copilot-instructions.md about sqlc generations and file-structure
- [x] Add GetTitle() to MetricsProvider interface and use it in metrics.buildPrometheusMetrics in output labels
- [x] Correct OpenAPI spec. Endpoint enabled is required
- [x] Double check OpenAPI spec. AgentConfig.tags is required AgentConfig.Endpoints.tags is optional
- [x] Correct DB schema. Add optional port to the Endpoint
- [x] Database schema with support for tenants, sections, agents, endpoints, and tags
- [x] sqlc integration for type-safe database queries
- [x] Agent configuration endpoint implementation (GET /agent/{agentId}/configuration)
- [x] Database versioning triggers for automatic configuration version bumping
- [x] justfile for build automation (replacing Makefile)

### Bugfixes that are part of a current PR
- [ ] Is it possible to replace in handlers.authenticated_handler.GetAgentConfiguration function body to call of the middleware.RequiredAuth ?

- [ ] middleware/auth.go AgentAPIKeyAuth should NOT return HTTP Unauthorized (line 63) or HTTP InternalServerError (line 71). This change will fail some tests, but logic must be correct.

- [ ] Return back to auth.RequireAuth and review api.Error return blocks
- [ ] Try to refactor auth.RequireAuth to use api.gen.go strict types. (I've added error handlers manually, without using strict handlers)

- [ ] Update copilot-instruct to use api.Error struct if HTTP error should be returned. Do not use inline JSON without api.Erorr
- [ ] Check that existing code follows api.Error struct approach

- [ ] Update copilot-instruct to use UUIDv7 for all entities.id in DB

- [ ] Implement rate-limiting for endpoints that are using security schema AgentApiKey

- [ ] Seems that there is some mess with "version" in handlers/agent_configuration/configuration.go 
      - version-parameter sent to constructor is an app version
      - version in the GetAgentConfiguration handler is version that coming from DB and tracking agent version. 
- [ ] Related to the previous mess. 
      - OpenAPI spec saying that version parameter should be send in query
      - Check the agent. I think that it sends version in HTTP-header

### Short Term
- [ ] Database migrations management with go-migrate or similar tool
- [ ] JWT authentication implementation
- [ ] User management endpoints
- [ ] Agent registration and management
- [ ] Docker and docker-compose setup

### Medium Term
- [ ] OAuth2 integration
- [ ] Metrics collection from agents
- [ ] Alert configuration and notification system
- [ ] Web dashboard (frontend)
- [ ] TimescaleDB integration for time-series data

### Long Term
- [ ] Kubernetes deployment with Helm charts
- [ ] Plugin system for extensibility
- [ ] Advanced monitoring features
- [ ] Distributed tracing
- [ ] Multi-tenant support

