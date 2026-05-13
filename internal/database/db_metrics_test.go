package database

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func newTestSQLDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestDBMetrics_GetMetrics_Format(t *testing.T) {
	sqlDB := newTestSQLDB(t)

	m := NewDBMetrics(&staticDB{sqlDB})
	out := m.GetMetrics()

	for _, expected := range []string{
		"# HELP smotra_db_connections_open",
		"# TYPE smotra_db_connections_open gauge",
		"smotra_db_connections_open",
		"smotra_db_connections_in_use",
		"smotra_db_connections_idle",
		"smotra_db_wait_count_total",
		"smotra_db_wait_duration_ms",
	} {
		if !strings.Contains(out, expected) {
			t.Errorf("expected %q in GetMetrics output:\n%s", expected, out)
		}
	}
}

// staticDB is a minimal Database stub that returns a fixed *sql.DB.
type staticDB struct {
	db *sql.DB
}

func (s *staticDB) DB() *sql.DB                                                  { return s.db }
func (s *staticDB) Open(_ context.Context) error                                 { return nil }
func (s *staticDB) Close() error                                                 { return nil }
func (s *staticDB) Ping(_ context.Context) error                                 { return nil }
func (s *staticDB) Health(_ context.Context) (HealthInfo, error)                 { return HealthInfo{}, nil }
func (s *staticDB) BeginTx(_ context.Context) (*sql.Tx, error)                   { return nil, nil }
