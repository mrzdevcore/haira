package haira

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// DB wraps a PostgreSQL connection pool.
type DB struct {
	conn *sql.DB
}

// PostgresConnect opens a connection to a PostgreSQL database.
// connStr is a standard PostgreSQL connection string:
// "postgresql://user:pass@host:port/dbname?sslmode=disable"
func PostgresConnect(connStr string) (*DB, error) {
	if connStr == "" {
		return nil, fmt.Errorf("postgres connect: empty connection string")
	}

	// Inject connect_timeout into the connection string so lib/pq enforces
	// a driver-level timeout on the TCP dial + TLS handshake. This prevents
	// hanging when the host is unreachable (where context timeout alone may
	// not help because DNS resolution or the driver dial can block first).
	connStr = ensureConnectTimeout(connStr, 5)

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("postgres connect: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	if err := conn.PingContext(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("postgres ping: %w", err)
	}
	return &DB{conn: conn}, nil
}

// ensureConnectTimeout adds connect_timeout to the connection string if not
// already present. Supports both URL and keyword=value formats.
func ensureConnectTimeout(connStr string, seconds int) string {
	// URL format: postgresql://... or postgres://...
	if u, err := url.Parse(connStr); err == nil && (u.Scheme == "postgresql" || u.Scheme == "postgres") {
		q := u.Query()
		if q.Get("connect_timeout") == "" {
			q.Set("connect_timeout", fmt.Sprintf("%d", seconds))
			u.RawQuery = q.Encode()
			return u.String()
		}
		return connStr
	}
	// Keyword=value format: "host=... port=... dbname=..."
	if !strings.Contains(connStr, "connect_timeout") {
		return fmt.Sprintf("%s connect_timeout=%d", connStr, seconds)
	}
	return connStr
}

// Query executes a SQL query and returns rows as a slice of maps.
// Each map represents a row with column names as keys.
func (db *DB) Query(query string, args ...any) ([]map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	rows, err := db.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("postgres query: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("getting columns: %w", err)
	}

	var results []map[string]any
	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}

		row := make(map[string]any, len(columns))
		for i, col := range columns {
			val := values[i]
			// Convert []byte to string for readability
			if b, ok := val.([]byte); ok {
				val = string(b)
			}
			row[col] = val
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating rows: %w", err)
	}

	return results, nil
}

// Execute runs a SQL statement (INSERT, UPDATE, DELETE) and returns rows affected.
func (db *DB) Execute(query string, args ...any) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	result, err := db.conn.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("postgres execute: %w", err)
	}
	return result.RowsAffected()
}

// Close closes the database connection pool.
func (db *DB) Close() error {
	return db.conn.Close()
}
