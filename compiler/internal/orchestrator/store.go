package orchestrator

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Deployment represents a deployed Haira project.
type Deployment struct {
	ID         string
	Name       string
	SourcePath string
	BinaryPath string
	Port       int
	Status     string // "running" | "stopped" | "crashed" | "deploying"
	PID        int
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Restarts   int
}

// Store manages deployment metadata in SQLite.
type Store struct {
	db *sql.DB
}

var schema = []string{
	`CREATE TABLE IF NOT EXISTS deployments (
		id TEXT PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		source_path TEXT NOT NULL,
		binary_path TEXT NOT NULL,
		port INTEGER NOT NULL,
		status TEXT NOT NULL DEFAULT 'deploying',
		pid INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		restarts INTEGER NOT NULL DEFAULT 0
	)`,
}

// NewStore creates a SQLite-backed deployment store.
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	for _, ddl := range schema {
		if _, err := db.Exec(ddl); err != nil {
			return nil, fmt.Errorf("schema: %w", err)
		}
	}
	return &Store{db: db}, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Create inserts a new deployment record.
func (s *Store) Create(d *Deployment) error {
	_, err := s.db.Exec(
		`INSERT INTO deployments (id, name, source_path, binary_path, port, status, pid, created_at, updated_at, restarts)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		d.ID, d.Name, d.SourcePath, d.BinaryPath, d.Port, d.Status, d.PID,
		d.CreatedAt, d.UpdatedAt, d.Restarts,
	)
	return err
}

// GetByName retrieves a deployment by name.
func (s *Store) GetByName(name string) (*Deployment, error) {
	d := &Deployment{}
	err := s.db.QueryRow(
		`SELECT id, name, source_path, binary_path, port, status, pid, created_at, updated_at, restarts
		 FROM deployments WHERE name = ?`, name,
	).Scan(&d.ID, &d.Name, &d.SourcePath, &d.BinaryPath, &d.Port, &d.Status, &d.PID,
		&d.CreatedAt, &d.UpdatedAt, &d.Restarts)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return d, nil
}

// List returns all deployments.
func (s *Store) List() ([]*Deployment, error) {
	rows, err := s.db.Query(
		`SELECT id, name, source_path, binary_path, port, status, pid, created_at, updated_at, restarts
		 FROM deployments ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Deployment
	for rows.Next() {
		d := &Deployment{}
		if err := rows.Scan(&d.ID, &d.Name, &d.SourcePath, &d.BinaryPath, &d.Port, &d.Status, &d.PID,
			&d.CreatedAt, &d.UpdatedAt, &d.Restarts); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// Update saves changes to a deployment record.
func (s *Store) Update(d *Deployment) error {
	d.UpdatedAt = time.Now()
	_, err := s.db.Exec(
		`UPDATE deployments SET source_path=?, binary_path=?, port=?, status=?, pid=?, updated_at=?, restarts=?
		 WHERE name=?`,
		d.SourcePath, d.BinaryPath, d.Port, d.Status, d.PID, d.UpdatedAt, d.Restarts, d.Name,
	)
	return err
}

// Delete removes a deployment record by name.
func (s *Store) Delete(name string) error {
	_, err := s.db.Exec(`DELETE FROM deployments WHERE name = ?`, name)
	return err
}
