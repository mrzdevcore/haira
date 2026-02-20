package haira

import (
	"database/sql"
	"fmt"

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
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("postgres connect: %w", err)
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("postgres ping: %w", err)
	}
	return &DB{conn: conn}, nil
}

// Query executes a SQL query and returns rows as a slice of maps.
// Each map represents a row with column names as keys.
func (db *DB) Query(query string, args ...any) ([]map[string]any, error) {
	rows, err := db.conn.Query(query, args...)
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
	result, err := db.conn.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("postgres execute: %w", err)
	}
	return result.RowsAffected()
}

// Close closes the database connection pool.
func (db *DB) Close() error {
	return db.conn.Close()
}
