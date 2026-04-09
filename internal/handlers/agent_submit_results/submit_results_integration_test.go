package agent_submit_results

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	api "github.com/smotra-monitoring/server/internal/api/v1"
	"github.com/smotra-monitoring/server/internal/database"
	"github.com/smotra-monitoring/server/internal/logger"
	"github.com/smotra-monitoring/server/internal/testutil"
)

// testServerImpl satisfies the full api.StrictServerInterface by delegating
// SubmitAgentResults to the handler under test and stubbing everything else.
type testServerImpl struct {
	*Handler
}

func (s *testServerImpl) SubmitAgentResults(ctx context.Context, req api.SubmitAgentResultsRequestObject) (api.SubmitAgentResultsResponseObject, error) {
	return s.Handle(ctx, req)
}
func (s *testServerImpl) GetAgentConfiguration(ctx context.Context, req api.GetAgentConfigurationRequestObject) (api.GetAgentConfigurationResponseObject, error) {
	return nil, nil
}
func (s *testServerImpl) RegisterAgentSelf(ctx context.Context, req api.RegisterAgentSelfRequestObject) (api.RegisterAgentSelfResponseObject, error) {
	return nil, nil
}
func (s *testServerImpl) GetAgentClaimStatus(ctx context.Context, req api.GetAgentClaimStatusRequestObject) (api.GetAgentClaimStatusResponseObject, error) {
	return nil, nil
}
func (s *testServerImpl) PostClaimAgent(ctx context.Context, req api.PostClaimAgentRequestObject) (api.PostClaimAgentResponseObject, error) {
	return nil, nil
}

func setupTestRouter(h *Handler) *chi.Mux {
	impl := &testServerImpl{Handler: h}
	r := chi.NewRouter()
	api.HandlerFromMux(api.NewStrictHandler(impl, nil), r)
	return r
}

// setupRealDB creates a SQLite DB with all migrations applied plus a test tenant/section/agent.
func setupRealDB(t *testing.T) (database.Database, uuid.UUID) {
	t.Helper()
	db := testutil.SetupTestSQLiteDB(t)
	ctx := context.Background()
	testutil.ApplyMigrations(t, ctx, db.DB(), "../../../data/db/dev/migrations")

	tenantID := uuid.Must(uuid.NewV7()).String()
	if _, err := db.DB().ExecContext(ctx, `INSERT INTO tenants (id, name) VALUES (?, ?)`, tenantID, "Test Tenant"); err != nil {
		t.Fatalf("insert tenant: %v", err)
	}
	sectionID := uuid.Must(uuid.NewV7()).String()
	if _, err := db.DB().ExecContext(ctx, `INSERT INTO sections (id, tenant_id, name) VALUES (?, ?, ?)`, sectionID, tenantID, "Default"); err != nil {
		t.Fatalf("insert section: %v", err)
	}
	agentID := uuid.Must(uuid.NewV7())
	if _, err := db.DB().ExecContext(ctx,
		`INSERT INTO agents (id, section_id, name, api_key_hash, base_config) VALUES (?, ?, ?, ?, ?)`,
		agentID.String(), sectionID, "test-agent", "fakehash", "{}"); err != nil {
		t.Fatalf("insert agent: %v", err)
	}

	return db, agentID
}

func pingCheckType(t *testing.T, resolved string, successes, failures int32) api.CheckType {
	t.Helper()
	avg := 12.5
	raw, _ := json.Marshal(api.PingCheck{
		Type: "ping",
		Result: api.PingResult{
			ResolvedIp:        resolved,
			Successes:         successes,
			Failures:          failures,
			AvgResponseTimeMs: &avg,
			SuccessLatencies:  []float64{10.0, 15.0},
		},
	})
	var ct api.CheckType
	if err := json.Unmarshal(raw, &ct); err != nil {
		t.Fatalf("unmarshal CheckType: %v", err)
	}
	return ct
}

func tracerouteCheckType(t *testing.T, hops []api.TracerouteHop) api.CheckType {
	t.Helper()
	total := 45.2
	raw, _ := json.Marshal(api.TracerouteCheck{
		Type: "traceroute",
		Result: api.TracerouteResult{
			TargetReached: true,
			TotalTimeMs:   &total,
			Hops:          hops,
		},
	})
	var ct api.CheckType
	if err := json.Unmarshal(raw, &ct); err != nil {
		t.Fatalf("unmarshal CheckType: %v", err)
	}
	return ct
}

func postBatch(t *testing.T, router *chi.Mux, agentID uuid.UUID, results []api.MonitoringResult) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(api.BatchMonitoringResults{Results: results})
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/agent/%s/results", agentID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestIntegration_PingBatch_Accepted(t *testing.T) {
	db, agentID := setupRealDB(t)
	h := NewHandler(logger.Default(), db)
	router := setupTestRouter(h)

	result := api.MonitoringResult{
		Id:        uuid.Must(uuid.NewV7()),
		AgentId:   agentID,
		CheckType: pingCheckType(t, "8.8.8.8", 5, 0),
		Target:    api.Endpoint{Address: "8.8.8.8", Tags: []string{}},
		Timestamp: time.Now().UTC(),
	}

	w := postBatch(t, router, agentID, []api.MonitoringResult{result})
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", w.Code, w.Body.String())
	}

	var ack api.ResultsBatchAcknowledgment
	if err := json.Unmarshal(w.Body.Bytes(), &ack); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ack.Accepted != 1 {
		t.Errorf("expected accepted=1, got %d", ack.Accepted)
	}
	if ack.DuplicatesSkipped != nil && *ack.DuplicatesSkipped != 0 {
		t.Errorf("expected duplicates=0, got %d", *ack.DuplicatesSkipped)
	}
	if ack.SubmissionId == (uuid.UUID{}) {
		t.Error("expected non-zero submission_id")
	}

	var cnt int
	db.DB().QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM check_results WHERE agent_id = ? AND check_type = 'ping'`, agentID.String()).Scan(&cnt)
	if cnt != 1 {
		t.Errorf("expected 1 row in check_results, got %d", cnt)
	}
	var pingCnt int
	db.DB().QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM ping_check_results WHERE check_id = ?`, result.Id.String()).Scan(&pingCnt)
	if pingCnt != 1 {
		t.Errorf("expected 1 row in ping_check_results, got %d", pingCnt)
	}
}

func TestIntegration_Deduplication(t *testing.T) {
	db, agentID := setupRealDB(t)
	h := NewHandler(logger.Default(), db)
	router := setupTestRouter(h)

	result := api.MonitoringResult{
		Id:        uuid.Must(uuid.NewV7()),
		AgentId:   agentID,
		CheckType: pingCheckType(t, "1.2.3.4", 3, 1),
		Target:    api.Endpoint{Address: "1.2.3.4", Tags: []string{}},
		Timestamp: time.Now().UTC(),
	}

	w1 := postBatch(t, router, agentID, []api.MonitoringResult{result})
	if w1.Code != http.StatusAccepted {
		t.Fatalf("first: expected 202, got %d", w1.Code)
	}

	w2 := postBatch(t, router, agentID, []api.MonitoringResult{result})
	if w2.Code != http.StatusAccepted {
		t.Fatalf("second: expected 202, got %d: %s", w2.Code, w2.Body.String())
	}

	var ack api.ResultsBatchAcknowledgment
	if err := json.Unmarshal(w2.Body.Bytes(), &ack); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ack.Accepted != 0 {
		t.Errorf("expected 0 accepted on second POST, got %d", ack.Accepted)
	}
	if ack.DuplicatesSkipped == nil || *ack.DuplicatesSkipped != 1 {
		t.Errorf("expected 1 duplicate, got %v", ack.DuplicatesSkipped)
	}

	var cnt int
	db.DB().QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM check_results WHERE id = ?`, result.Id.String()).Scan(&cnt)
	if cnt != 1 {
		t.Errorf("expected exactly 1 row, got %d", cnt)
	}
}

func TestIntegration_UpdatesLastSeenAt(t *testing.T) {
	db, agentID := setupRealDB(t)
	h := NewHandler(logger.Default(), db)
	router := setupTestRouter(h)

	before := time.Now().UTC().Add(-time.Second)
	result := api.MonitoringResult{
		Id:        uuid.Must(uuid.NewV7()),
		AgentId:   agentID,
		CheckType: pingCheckType(t, "9.9.9.9", 1, 0),
		Target:    api.Endpoint{Address: "9.9.9.9", Tags: []string{}},
		Timestamp: time.Now().UTC(),
	}

	w := postBatch(t, router, agentID, []api.MonitoringResult{result})
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", w.Code)
	}

	var lastSeen sql.NullTime
	db.DB().QueryRowContext(context.Background(),
		`SELECT last_seen_at FROM agents WHERE id = ?`, agentID.String()).Scan(&lastSeen)
	if !lastSeen.Valid {
		t.Fatal("last_seen_at should be set")
	}
	if lastSeen.Time.Before(before) {
		t.Errorf("last_seen_at %v before submission time %v", lastSeen.Time, before)
	}
}

func TestIntegration_EndpointIDResolved(t *testing.T) {
	db, agentID := setupRealDB(t)
	ctx := context.Background()

	endpointID := uuid.Must(uuid.NewV7()).String()
	if _, err := db.DB().ExecContext(ctx,
		`INSERT INTO endpoints (id, agent_id, address, enabled) VALUES (?, ?, ?, 1)`,
		endpointID, agentID.String(), "10.0.0.1"); err != nil {
		t.Fatalf("insert endpoint: %v", err)
	}

	h := NewHandler(logger.Default(), db)
	router := setupTestRouter(h)
	result := api.MonitoringResult{
		Id:        uuid.Must(uuid.NewV7()),
		AgentId:   agentID,
		CheckType: pingCheckType(t, "10.0.0.1", 2, 0),
		Target:    api.Endpoint{Address: "10.0.0.1", Tags: []string{}},
		Timestamp: time.Now().UTC(),
	}

	w := postBatch(t, router, agentID, []api.MonitoringResult{result})
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", w.Code, w.Body.String())
	}

	var resolvedID sql.NullString
	db.DB().QueryRowContext(ctx,
		`SELECT endpoint_id FROM check_results WHERE id = ?`, result.Id.String()).Scan(&resolvedID)
	if !resolvedID.Valid || resolvedID.String != endpointID {
		t.Errorf("expected endpoint_id %q, got %v", endpointID, resolvedID)
	}
}

func TestIntegration_TracerouteHopsStored(t *testing.T) {
	db, agentID := setupRealDB(t)
	h := NewHandler(logger.Default(), db)
	router := setupTestRouter(h)

	hop1addr := "192.168.1.1"
	hop2addr := "10.0.0.1"
	rt1, rt2 := 1.2, 5.6
	hops := []api.TracerouteHop{
		{Hop: 1, Address: &hop1addr, ResponseTimeMs: &rt1},
		{Hop: 2, Address: &hop2addr, ResponseTimeMs: &rt2},
	}
	result := api.MonitoringResult{
		Id:        uuid.Must(uuid.NewV7()),
		AgentId:   agentID,
		CheckType: tracerouteCheckType(t, hops),
		Target:    api.Endpoint{Address: "10.0.0.1", Tags: []string{}},
		Timestamp: time.Now().UTC(),
	}

	w := postBatch(t, router, agentID, []api.MonitoringResult{result})
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", w.Code, w.Body.String())
	}

	rows, err := db.DB().QueryContext(context.Background(),
		`SELECT hop, address FROM traceroute_hops WHERE check_id = ? ORDER BY hop`, result.Id.String())
	if err != nil {
		t.Fatalf("query traceroute_hops: %v", err)
	}
	defer rows.Close()

	type hopRow struct {
		hop  int
		addr sql.NullString
	}
	var got []hopRow
	for rows.Next() {
		var hr hopRow
		rows.Scan(&hr.hop, &hr.addr)
		got = append(got, hr)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 hops, got %d", len(got))
	}
	if got[0].hop != 1 || !got[0].addr.Valid || got[0].addr.String != hop1addr {
		t.Errorf("hop1 wrong: %+v", got[0])
	}
	if got[1].hop != 2 || !got[1].addr.Valid || got[1].addr.String != hop2addr {
		t.Errorf("hop2 wrong: %+v", got[1])
	}
}

func TestIntegration_MixedTypeBatch(t *testing.T) {
	db, agentID := setupRealDB(t)
	h := NewHandler(logger.Default(), db)
	router := setupTestRouter(h)

	results := []api.MonitoringResult{
		{
			Id: uuid.Must(uuid.NewV7()), AgentId: agentID,
			CheckType: pingCheckType(t, "8.8.8.8", 3, 0),
			Target:    api.Endpoint{Address: "8.8.8.8", Tags: []string{}},
			Timestamp: time.Now().UTC(),
		},
		{
			Id: uuid.Must(uuid.NewV7()), AgentId: agentID,
			CheckType: tracerouteCheckType(t, []api.TracerouteHop{{Hop: 1}}),
			Target:    api.Endpoint{Address: "8.8.8.8", Tags: []string{}},
			Timestamp: time.Now().UTC(),
		},
	}

	w := postBatch(t, router, agentID, results)
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", w.Code, w.Body.String())
	}

	var ack api.ResultsBatchAcknowledgment
	if err := json.Unmarshal(w.Body.Bytes(), &ack); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ack.Accepted != 2 {
		t.Errorf("expected 2 accepted, got %d", ack.Accepted)
	}

	var cnt int
	db.DB().QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM check_results WHERE agent_id = ?`, agentID.String()).Scan(&cnt)
	if cnt != 2 {
		t.Errorf("expected 2 rows in check_results, got %d", cnt)
	}
}
