package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"

	haira "haira-go-runtime/haira"

	_ "github.com/lib/pq"
)

func init() {
	haira.RegisterStoreBackend("postgres", func(url string) haira.Store {
		return NewPostgresStore(url)
	})
}

// PostgresStore implements Store using a PostgreSQL database.
type PostgresStore struct {
	connStr string
	db      *sql.DB
}

// NewPostgresStore creates a new Postgres-backed store.
func NewPostgresStore(connStr string) *PostgresStore {
	return &PostgresStore{connStr: connStr}
}

func (s *PostgresStore) Init() error {
	db, err := sql.Open("postgres", s.connStr)
	if err != nil {
		return fmt.Errorf("postgres open: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("postgres ping: %w", err)
	}
	s.db = db

	for _, ddl := range postgresSchema {
		if _, err := s.db.Exec(ddl); err != nil {
			return fmt.Errorf("postgres schema: %w", err)
		}
	}
	return nil
}

func (s *PostgresStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

var postgresSchema = []string{
	`CREATE TABLE IF NOT EXISTS chat_sessions (
		id TEXT PRIMARY KEY,
		workflow_name TEXT NOT NULL,
		workflow_path TEXT NOT NULL,
		title TEXT DEFAULT '',
		owner TEXT DEFAULT '',
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW(),
		message_count INTEGER DEFAULT 0
	)`,
	`CREATE TABLE IF NOT EXISTS chat_messages (
		id SERIAL PRIMARY KEY,
		session_id TEXT NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
		role TEXT NOT NULL,
		content TEXT DEFAULT '',
		ui_events JSONB,
		created_at TIMESTAMPTZ DEFAULT NOW()
	)`,
	`CREATE INDEX IF NOT EXISTS idx_chat_messages_session ON chat_messages(session_id)`,
	`CREATE TABLE IF NOT EXISTS runs (
		id TEXT PRIMARY KEY,
		workflow_name TEXT NOT NULL,
		workflow_path TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'running',
		params JSONB,
		steps JSONB,
		result JSONB,
		error TEXT DEFAULT '',
		started_at TIMESTAMPTZ DEFAULT NOW(),
		finished_at TIMESTAMPTZ
	)`,
	`CREATE TABLE IF NOT EXISTS observe_generations (
		id TEXT PRIMARY KEY,
		agent_name TEXT NOT NULL DEFAULT '',
		model TEXT NOT NULL DEFAULT '',
		provider TEXT NOT NULL DEFAULT '',
		input_tokens INTEGER DEFAULT 0,
		output_tokens INTEGER DEFAULT 0,
		total_tokens INTEGER DEFAULT 0,
		cost_usd DOUBLE PRECISION DEFAULT 0,
		latency_ms BIGINT DEFAULT 0,
		temperature DOUBLE PRECISION DEFAULT 0,
		tool_calls INTEGER DEFAULT 0,
		finish_reason TEXT DEFAULT '',
		timestamp TIMESTAMPTZ DEFAULT NOW(),
		session_id TEXT DEFAULT ''
	)`,
	`CREATE TABLE IF NOT EXISTS observe_tool_execs (
		id TEXT PRIMARY KEY,
		agent_name TEXT NOT NULL DEFAULT '',
		tool_name TEXT NOT NULL DEFAULT '',
		latency_ms BIGINT DEFAULT 0,
		success BOOLEAN DEFAULT TRUE,
		timestamp TIMESTAMPTZ DEFAULT NOW(),
		session_id TEXT DEFAULT ''
	)`,
}

// --- Chat Sessions ---

func (s *PostgresStore) EnsureSession(id, wfName, wfPath, owner string) error {
	_, err := s.db.Exec(
		`INSERT INTO chat_sessions (id, workflow_name, workflow_path, owner)
		 VALUES ($1, $2, $3, $4) ON CONFLICT (id) DO NOTHING`,
		id, wfName, wfPath, owner,
	)
	return err
}

func (s *PostgresStore) AddMessage(sessionID, role, content string, uiEvents []json.RawMessage) error {
	var eventsJSON *string
	if len(uiEvents) > 0 {
		b, _ := json.Marshal(uiEvents)
		s := string(b)
		eventsJSON = &s
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`INSERT INTO chat_messages (session_id, role, content, ui_events) VALUES ($1, $2, $3, $4)`,
		sessionID, role, content, eventsJSON,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		UPDATE chat_sessions SET
			message_count = (SELECT COUNT(*) FROM chat_messages WHERE session_id = $1),
			updated_at = NOW(),
			title = CASE WHEN title = '' AND $2 = 'user' AND $3 != ''
				THEN LEFT($3, 80)
				ELSE title END
		WHERE id = $1`,
		sessionID, role, content,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *PostgresStore) GetSession(id string) (*haira.ChatSessionDetail, error) {
	row := s.db.QueryRow(
		`SELECT id, workflow_name, workflow_path, title, COALESCE(owner, ''), created_at, updated_at, message_count
		 FROM chat_sessions WHERE id = $1`, id,
	)

	var sess haira.ChatSessionDetail
	err := row.Scan(
		&sess.ID, &sess.WorkflowName, &sess.WorkflowPath,
		&sess.Title, &sess.Owner, &sess.CreatedAt, &sess.UpdatedAt, &sess.MessageCount,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(
		`SELECT role, content, ui_events, created_at FROM chat_messages WHERE session_id = $1 ORDER BY id ASC`, id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sess.Messages = []haira.ChatMessage{}
	for rows.Next() {
		var msg haira.ChatMessage
		var eventsRaw *[]byte
		if err := rows.Scan(&msg.Role, &msg.Content, &eventsRaw, &msg.Timestamp); err != nil {
			return nil, err
		}
		if eventsRaw != nil {
			msg.UIEvents = json.RawMessage(*eventsRaw)
		}
		sess.Messages = append(sess.Messages, msg)
	}

	return &sess, nil
}

func (s *PostgresStore) ListSessions(wfPath, owner string) ([]haira.ChatSession, error) {
	query := `SELECT id, workflow_name, workflow_path, title, COALESCE(owner, ''), created_at, updated_at, message_count
			  FROM chat_sessions WHERE TRUE`
	var args []any
	n := 0

	if wfPath != "" {
		n++
		query += fmt.Sprintf(` AND workflow_path = $%d`, n)
		args = append(args, wfPath)
	}
	if owner != "" {
		n++
		query += fmt.Sprintf(` AND owner = $%d`, n)
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
		if err := rows.Scan(
			&sess.ID, &sess.WorkflowName, &sess.WorkflowPath,
			&sess.Title, &sess.Owner, &sess.CreatedAt, &sess.UpdatedAt, &sess.MessageCount,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, sess)
	}

	if sessions == nil {
		sessions = []haira.ChatSession{}
	}
	return sessions, nil
}

func (s *PostgresStore) DeleteSession(id string) error {
	_, err := s.db.Exec(`DELETE FROM chat_sessions WHERE id = $1`, id)
	return err
}

// --- Runs ---

func (s *PostgresStore) CreateRun(run *haira.Run) error {
	paramsJSON, _ := json.Marshal(run.Params)
	stepsJSON, _ := json.Marshal(run.Steps)

	_, err := s.db.Exec(
		`INSERT INTO runs (id, workflow_name, workflow_path, status, params, steps, started_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		run.ID, run.WorkflowName, run.WorkflowPath, run.Status,
		string(paramsJSON), string(stepsJSON), run.StartedAt,
	)
	return err
}

func (s *PostgresStore) UpdateRun(run *haira.Run) error {
	stepsJSON, _ := json.Marshal(run.Steps)
	var resultJSON *string
	if run.Result != nil {
		b, _ := json.Marshal(run.Result)
		s := string(b)
		resultJSON = &s
	}

	_, err := s.db.Exec(
		`UPDATE runs SET status = $1, steps = $2, result = $3, error = $4, finished_at = $5 WHERE id = $6`,
		run.Status, string(stepsJSON), resultJSON, run.Error, run.FinishedAt, run.ID,
	)
	return err
}

func (s *PostgresStore) GetRun(id string) (*haira.Run, error) {
	row := s.db.QueryRow(
		`SELECT id, workflow_name, workflow_path, status, params, steps, result, error, started_at, finished_at
		 FROM runs WHERE id = $1`, id,
	)

	var run haira.Run
	var paramsRaw, stepsRaw, resultRaw *[]byte

	err := row.Scan(
		&run.ID, &run.WorkflowName, &run.WorkflowPath, &run.Status,
		&paramsRaw, &stepsRaw, &resultRaw, &run.Error, &run.StartedAt, &run.FinishedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if paramsRaw != nil {
		if err := json.Unmarshal(*paramsRaw, &run.Params); err != nil {
			return nil, fmt.Errorf("unmarshal run params: %w", err)
		}
	}
	if stepsRaw != nil {
		if err := json.Unmarshal(*stepsRaw, &run.Steps); err != nil {
			return nil, fmt.Errorf("unmarshal run steps: %w", err)
		}
	}
	if resultRaw != nil {
		if err := json.Unmarshal(*resultRaw, &run.Result); err != nil {
			return nil, fmt.Errorf("unmarshal run result: %w", err)
		}
	}
	if run.Steps == nil {
		run.Steps = []haira.StepEvent{}
	}

	return &run, nil
}

func (s *PostgresStore) ListRuns(wfPath string) ([]haira.RunSummary, error) {
	query := `SELECT id, workflow_name, workflow_path, status, started_at, finished_at,
			  COALESCE(jsonb_array_length(steps), 0) as step_count
			  FROM runs WHERE TRUE`
	var args []any

	if wfPath != "" {
		query += ` AND workflow_path = $1`
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
		if err := rows.Scan(
			&r.ID, &r.WorkflowName, &r.WorkflowPath, &r.Status,
			&r.StartedAt, &r.FinishedAt, &r.StepCount,
		); err != nil {
			return nil, err
		}
		runs = append(runs, r)
	}

	if runs == nil {
		runs = []haira.RunSummary{}
	}
	return runs, nil
}

// --- Observability ---

func (s *PostgresStore) SaveGeneration(gen haira.LLMGeneration) error {
	_, err := s.db.Exec(
		`INSERT INTO observe_generations
		 (id, agent_name, model, provider, input_tokens, output_tokens, total_tokens,
		  cost_usd, latency_ms, temperature, tool_calls, finish_reason, timestamp, session_id)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		 ON CONFLICT (id) DO NOTHING`,
		gen.ID, gen.AgentName, gen.Model, gen.Provider,
		gen.InputTokens, gen.OutputTokens, gen.TotalTokens,
		gen.CostUSD, gen.LatencyMs, gen.Temperature,
		gen.ToolCalls, gen.FinishReason, gen.Timestamp, gen.SessionID,
	)
	return err
}

func (s *PostgresStore) SaveToolExec(exec haira.ToolExec) error {
	_, err := s.db.Exec(
		`INSERT INTO observe_tool_execs
		 (id, agent_name, tool_name, latency_ms, success, timestamp, session_id)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)
		 ON CONFLICT (id) DO NOTHING`,
		exec.ID, exec.AgentName, exec.ToolName,
		exec.LatencyMs, exec.Success, exec.Timestamp, exec.SessionID,
	)
	return err
}

func (s *PostgresStore) LoadGenerations() ([]haira.LLMGeneration, error) {
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
		if err := rows.Scan(
			&g.ID, &g.AgentName, &g.Model, &g.Provider,
			&g.InputTokens, &g.OutputTokens, &g.TotalTokens,
			&g.CostUSD, &g.LatencyMs, &g.Temperature,
			&g.ToolCalls, &g.FinishReason, &g.Timestamp, &g.SessionID,
		); err != nil {
			return nil, err
		}
		gens = append(gens, g)
	}
	if gens == nil {
		gens = []haira.LLMGeneration{}
	}
	return gens, nil
}

func (s *PostgresStore) LoadToolExecs() ([]haira.ToolExec, error) {
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
		if err := rows.Scan(
			&e.ID, &e.AgentName, &e.ToolName,
			&e.LatencyMs, &e.Success, &e.Timestamp, &e.SessionID,
		); err != nil {
			return nil, err
		}
		execs = append(execs, e)
	}
	if execs == nil {
		execs = []haira.ToolExec{}
	}
	return execs, nil
}
