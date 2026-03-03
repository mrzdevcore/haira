package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	haira "haira-go-runtime/haira"

	_ "modernc.org/sqlite"
)

func init() {
	haira.RegisterStoreBackend("sqlite", func(path string) haira.Store {
		return NewSQLiteStore(path)
	})
}

// SQLiteStore implements Store using an embedded SQLite database.
type SQLiteStore struct {
	path string
	db   *sql.DB
}

// NewSQLiteStore creates a new SQLite-backed store.
func NewSQLiteStore(path string) *SQLiteStore {
	return &SQLiteStore{path: path}
}

func (s *SQLiteStore) Init() error {
	db, err := sql.Open("sqlite", s.path+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)")
	if err != nil {
		return fmt.Errorf("sqlite open: %w", err)
	}
	s.db = db

	// Create tables
	for _, ddl := range sqliteSchema {
		if _, err := s.db.Exec(ddl); err != nil {
			return fmt.Errorf("sqlite schema: %w", err)
		}
	}
	return nil
}

func (s *SQLiteStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

var sqliteSchema = []string{
	`CREATE TABLE IF NOT EXISTS chat_sessions (
		id TEXT PRIMARY KEY,
		workflow_name TEXT NOT NULL,
		workflow_path TEXT NOT NULL,
		title TEXT DEFAULT '',
		owner TEXT DEFAULT '',
		created_at DATETIME DEFAULT (datetime('now')),
		updated_at DATETIME DEFAULT (datetime('now')),
		message_count INTEGER DEFAULT 0
	)`,
	`CREATE TABLE IF NOT EXISTS chat_messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
		role TEXT NOT NULL,
		content TEXT DEFAULT '',
		ui_events TEXT,
		created_at DATETIME DEFAULT (datetime('now'))
	)`,
	`CREATE INDEX IF NOT EXISTS idx_chat_messages_session ON chat_messages(session_id)`,
	`CREATE TABLE IF NOT EXISTS runs (
		id TEXT PRIMARY KEY,
		workflow_name TEXT NOT NULL,
		workflow_path TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'running',
		params TEXT,
		steps TEXT,
		result TEXT,
		error TEXT DEFAULT '',
		started_at DATETIME DEFAULT (datetime('now')),
		finished_at DATETIME
	)`,
	`CREATE TABLE IF NOT EXISTS observe_generations (
		id TEXT PRIMARY KEY,
		agent_name TEXT NOT NULL DEFAULT '',
		model TEXT NOT NULL DEFAULT '',
		provider TEXT NOT NULL DEFAULT '',
		input_tokens INTEGER DEFAULT 0,
		output_tokens INTEGER DEFAULT 0,
		total_tokens INTEGER DEFAULT 0,
		cost_usd REAL DEFAULT 0,
		latency_ms INTEGER DEFAULT 0,
		temperature REAL DEFAULT 0,
		tool_calls INTEGER DEFAULT 0,
		finish_reason TEXT DEFAULT '',
		timestamp DATETIME DEFAULT (datetime('now')),
		session_id TEXT DEFAULT ''
	)`,
	`CREATE TABLE IF NOT EXISTS observe_tool_execs (
		id TEXT PRIMARY KEY,
		agent_name TEXT NOT NULL DEFAULT '',
		tool_name TEXT NOT NULL DEFAULT '',
		latency_ms INTEGER DEFAULT 0,
		success INTEGER DEFAULT 1,
		timestamp DATETIME DEFAULT (datetime('now')),
		session_id TEXT DEFAULT ''
	)`,
}

// --- Chat Sessions ---

func (s *SQLiteStore) EnsureSession(id, wfName, wfPath, owner string) error {
	_, err := s.db.Exec(
		`INSERT OR IGNORE INTO chat_sessions (id, workflow_name, workflow_path, owner) VALUES (?, ?, ?, ?)`,
		id, wfName, wfPath, owner,
	)
	return err
}

func (s *SQLiteStore) AddMessage(sessionID, role, content string, uiEvents []json.RawMessage) error {
	var eventsJSON *string
	if len(uiEvents) > 0 {
		b, _ := json.Marshal(uiEvents)
		str := string(b)
		eventsJSON = &str
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`INSERT INTO chat_messages (session_id, role, content, ui_events) VALUES (?, ?, ?, ?)`,
		sessionID, role, content, eventsJSON,
	)
	if err != nil {
		return err
	}

	// Update session metadata
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = tx.Exec(`
		UPDATE chat_sessions SET
			message_count = (SELECT COUNT(*) FROM chat_messages WHERE session_id = ?),
			updated_at = ?,
			title = CASE WHEN title = '' AND ? = 'user' AND ? != ''
				THEN SUBSTR(?, 1, 80)
				ELSE title END
		WHERE id = ?`,
		sessionID, now, role, content, content, sessionID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *SQLiteStore) GetSession(id string) (*haira.ChatSessionDetail, error) {
	row := s.db.QueryRow(
		`SELECT id, workflow_name, workflow_path, title, owner, created_at, updated_at, message_count
		 FROM chat_sessions WHERE id = ?`, id,
	)

	var sess haira.ChatSessionDetail
	var createdAt, updatedAt string
	err := row.Scan(
		&sess.ID, &sess.WorkflowName, &sess.WorkflowPath,
		&sess.Title, &sess.Owner, &createdAt, &updatedAt, &sess.MessageCount,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	sess.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	sess.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	// Load messages
	rows, err := s.db.Query(
		`SELECT role, content, ui_events, created_at FROM chat_messages WHERE session_id = ? ORDER BY id ASC`, id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sess.Messages = []haira.ChatMessage{}
	for rows.Next() {
		var msg haira.ChatMessage
		var eventsStr *string
		var ts string
		if err := rows.Scan(&msg.Role, &msg.Content, &eventsStr, &ts); err != nil {
			return nil, err
		}
		msg.Timestamp, _ = time.Parse(time.RFC3339, ts)
		if eventsStr != nil {
			msg.UIEvents = json.RawMessage(*eventsStr)
		}
		sess.Messages = append(sess.Messages, msg)
	}

	return &sess, nil
}

func (s *SQLiteStore) ListSessions(wfPath, owner string) ([]haira.ChatSession, error) {
	query := `SELECT id, workflow_name, workflow_path, title, owner, created_at, updated_at, message_count
			  FROM chat_sessions WHERE 1=1`
	var args []any

	if wfPath != "" {
		query += ` AND workflow_path = ?`
		args = append(args, wfPath)
	}
	if owner != "" {
		query += ` AND owner = ?`
		args = append(args, owner)
	}
	query += ` ORDER BY updated_at DESC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []haira.ChatSession
	for rows.Next() {
		var sess haira.ChatSession
		var createdAt, updatedAt string
		if err := rows.Scan(
			&sess.ID, &sess.WorkflowName, &sess.WorkflowPath,
			&sess.Title, &sess.Owner, &createdAt, &updatedAt, &sess.MessageCount,
		); err != nil {
			return nil, err
		}
		sess.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		sess.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		sessions = append(sessions, sess)
	}

	if sessions == nil {
		sessions = []haira.ChatSession{}
	}
	return sessions, nil
}

func (s *SQLiteStore) DeleteSession(id string) error {
	_, err := s.db.Exec(`DELETE FROM chat_sessions WHERE id = ?`, id)
	return err
}

// --- Runs ---

func (s *SQLiteStore) CreateRun(run *haira.Run) error {
	paramsJSON, _ := json.Marshal(run.Params)
	stepsJSON, _ := json.Marshal(run.Steps)

	_, err := s.db.Exec(
		`INSERT INTO runs (id, workflow_name, workflow_path, status, params, steps, started_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		run.ID, run.WorkflowName, run.WorkflowPath, run.Status,
		string(paramsJSON), string(stepsJSON), run.StartedAt.UTC().Format(time.RFC3339),
	)
	return err
}

func (s *SQLiteStore) UpdateRun(run *haira.Run) error {
	stepsJSON, _ := json.Marshal(run.Steps)
	var resultJSON *string
	if run.Result != nil {
		b, _ := json.Marshal(run.Result)
		str := string(b)
		resultJSON = &str
	}
	var finishedAt *string
	if run.FinishedAt != nil {
		str := run.FinishedAt.UTC().Format(time.RFC3339)
		finishedAt = &str
	}

	_, err := s.db.Exec(
		`UPDATE runs SET status = ?, steps = ?, result = ?, error = ?, finished_at = ? WHERE id = ?`,
		run.Status, string(stepsJSON), resultJSON, run.Error, finishedAt, run.ID,
	)
	return err
}

func (s *SQLiteStore) GetRun(id string) (*haira.Run, error) {
	row := s.db.QueryRow(
		`SELECT id, workflow_name, workflow_path, status, params, steps, result, error, started_at, finished_at
		 FROM runs WHERE id = ?`, id,
	)

	var run haira.Run
	var paramsStr, stepsStr *string
	var resultStr *string
	var startedAt string
	var finishedAt *string

	err := row.Scan(
		&run.ID, &run.WorkflowName, &run.WorkflowPath, &run.Status,
		&paramsStr, &stepsStr, &resultStr, &run.Error, &startedAt, &finishedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	run.StartedAt, _ = time.Parse(time.RFC3339, startedAt)
	if finishedAt != nil {
		t, _ := time.Parse(time.RFC3339, *finishedAt)
		run.FinishedAt = &t
	}
	if paramsStr != nil {
		json.Unmarshal([]byte(*paramsStr), &run.Params)
	}
	if stepsStr != nil {
		json.Unmarshal([]byte(*stepsStr), &run.Steps)
	}
	if resultStr != nil {
		json.Unmarshal([]byte(*resultStr), &run.Result)
	}
	if run.Steps == nil {
		run.Steps = []haira.StepEvent{}
	}

	return &run, nil
}

func (s *SQLiteStore) ListRuns(wfPath string) ([]haira.RunSummary, error) {
	query := `SELECT id, workflow_name, workflow_path, status, started_at, finished_at,
			  (SELECT COUNT(*) FROM json_each(steps)) as step_count
			  FROM runs WHERE 1=1`
	var args []any

	if wfPath != "" {
		query += ` AND workflow_path = ?`
		args = append(args, wfPath)
	}
	query += ` ORDER BY started_at DESC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []haira.RunSummary
	for rows.Next() {
		var r haira.RunSummary
		var startedAt string
		var finishedAt *string
		if err := rows.Scan(
			&r.ID, &r.WorkflowName, &r.WorkflowPath, &r.Status,
			&startedAt, &finishedAt, &r.StepCount,
		); err != nil {
			return nil, err
		}
		r.StartedAt, _ = time.Parse(time.RFC3339, startedAt)
		if finishedAt != nil {
			t, _ := time.Parse(time.RFC3339, *finishedAt)
			r.FinishedAt = &t
		}
		runs = append(runs, r)
	}

	if runs == nil {
		runs = []haira.RunSummary{}
	}
	return runs, nil
}

// --- Observability ---

func (s *SQLiteStore) SaveGeneration(gen haira.LLMGeneration) error {
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO observe_generations
		 (id, agent_name, model, provider, input_tokens, output_tokens, total_tokens,
		  cost_usd, latency_ms, temperature, tool_calls, finish_reason, timestamp, session_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		gen.ID, gen.AgentName, gen.Model, gen.Provider,
		gen.InputTokens, gen.OutputTokens, gen.TotalTokens,
		gen.CostUSD, gen.LatencyMs, gen.Temperature,
		gen.ToolCalls, gen.FinishReason,
		gen.Timestamp.UTC().Format(time.RFC3339), gen.SessionID,
	)
	return err
}

func (s *SQLiteStore) SaveToolExec(exec haira.ToolExec) error {
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO observe_tool_execs
		 (id, agent_name, tool_name, latency_ms, success, timestamp, session_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		exec.ID, exec.AgentName, exec.ToolName,
		exec.LatencyMs, exec.Success,
		exec.Timestamp.UTC().Format(time.RFC3339), exec.SessionID,
	)
	return err
}

func (s *SQLiteStore) LoadGenerations() ([]haira.LLMGeneration, error) {
	rows, err := s.db.Query(
		`SELECT id, agent_name, model, provider, input_tokens, output_tokens, total_tokens,
		        cost_usd, latency_ms, temperature, tool_calls, finish_reason, timestamp, session_id
		 FROM observe_generations ORDER BY timestamp ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gens []haira.LLMGeneration
	for rows.Next() {
		var g haira.LLMGeneration
		var ts string
		if err := rows.Scan(
			&g.ID, &g.AgentName, &g.Model, &g.Provider,
			&g.InputTokens, &g.OutputTokens, &g.TotalTokens,
			&g.CostUSD, &g.LatencyMs, &g.Temperature,
			&g.ToolCalls, &g.FinishReason, &ts, &g.SessionID,
		); err != nil {
			return nil, err
		}
		g.Timestamp, _ = time.Parse(time.RFC3339, ts)
		gens = append(gens, g)
	}
	if gens == nil {
		gens = []haira.LLMGeneration{}
	}
	return gens, nil
}

func (s *SQLiteStore) LoadToolExecs() ([]haira.ToolExec, error) {
	rows, err := s.db.Query(
		`SELECT id, agent_name, tool_name, latency_ms, success, timestamp, session_id
		 FROM observe_tool_execs ORDER BY timestamp ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var execs []haira.ToolExec
	for rows.Next() {
		var e haira.ToolExec
		var ts string
		if err := rows.Scan(
			&e.ID, &e.AgentName, &e.ToolName,
			&e.LatencyMs, &e.Success, &ts, &e.SessionID,
		); err != nil {
			return nil, err
		}
		e.Timestamp, _ = time.Parse(time.RFC3339, ts)
		execs = append(execs, e)
	}
	if execs == nil {
		execs = []haira.ToolExec{}
	}
	return execs, nil
}
