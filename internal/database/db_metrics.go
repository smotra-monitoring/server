package database

import (
	"fmt"
)

// DBMetrics exposes database connection pool statistics as a MetricsProvider.
type DBMetrics struct {
	db Database
}

// NewDBMetrics creates a new DBMetrics wrapping the given Database.
func NewDBMetrics(db Database) *DBMetrics {
	return &DBMetrics{db: db}
}

// GetMetrics returns Prometheus-formatted connection pool metrics.
func (m *DBMetrics) GetMetrics() string {
	stats := m.db.DB().Stats()

	var out string

	out += "# HELP smotra_db_connections_open Current number of open connections to the database\n"
	out += "# TYPE smotra_db_connections_open gauge\n"
	out += fmt.Sprintf("smotra_db_connections_open %d\n", stats.OpenConnections)
	out += "\n"

	out += "# HELP smotra_db_connections_in_use Number of connections currently in use\n"
	out += "# TYPE smotra_db_connections_in_use gauge\n"
	out += fmt.Sprintf("smotra_db_connections_in_use %d\n", stats.InUse)
	out += "\n"

	out += "# HELP smotra_db_connections_idle Number of idle connections\n"
	out += "# TYPE smotra_db_connections_idle gauge\n"
	out += fmt.Sprintf("smotra_db_connections_idle %d\n", stats.Idle)
	out += "\n"

	out += "# HELP smotra_db_wait_count_total Total number of times a goroutine waited for a connection\n"
	out += "# TYPE smotra_db_wait_count_total counter\n"
	out += fmt.Sprintf("smotra_db_wait_count_total %d\n", stats.WaitCount)
	out += "\n"

	out += "# HELP smotra_db_wait_duration_ms Total time blocked waiting for a new connection (ms)\n"
	out += "# TYPE smotra_db_wait_duration_ms counter\n"
	out += fmt.Sprintf("smotra_db_wait_duration_ms %d\n", stats.WaitDuration.Milliseconds())
	out += "\n"

	return out
}
