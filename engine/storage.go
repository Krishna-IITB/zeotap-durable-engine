package engine

import (
	"database/sql"
	"encoding/json"

	_ "github.com/mattn/go-sqlite3" // sqlite driver
)

type Storage struct {
	db *sql.DB
}

// StepRecord maps directly to our SQL table
type StepRecord struct {
	WorkflowID string
	StepKey    string
	Status     string
	Output     string // JSON blob
}

func NewStorage(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Enforce schema.
	// Composite primary key ensures we don't get duplicate steps for the same workflow.
	schema := `
	CREATE TABLE IF NOT EXISTS steps (
		workflow_id TEXT,
		step_key TEXT,
		status TEXT,
		output TEXT,
		PRIMARY KEY (workflow_id, step_key)
	);`

	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveStep(workflowID, stepKey, status string, output interface{}) error {
	// Serialize output to store as text/json
	outputJSON, err := json.Marshal(output)
	if err != nil {
		return err
	}

	// Upsert: if it exists (maybe from a failed previous run), update it.
	query := `
		INSERT OR REPLACE INTO steps (workflow_id, step_key, status, output) 
		VALUES (?, ?, ?, ?)`

	_, err = s.db.Exec(query, workflowID, stepKey, status, string(outputJSON))
	return err
}

func (s *Storage) GetStep(workflowID, stepKey string) (*StepRecord, error) {
	var record StepRecord

	query := `
		SELECT workflow_id, step_key, status, output 
		FROM steps 
		WHERE workflow_id = ? AND step_key = ?`

	err := s.db.QueryRow(query, workflowID, stepKey).Scan(
		&record.WorkflowID,
		&record.StepKey,
		&record.Status,
		&record.Output,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not found is not an error, just means we run the step
	}
	if err != nil {
		return nil, err
	}

	return &record, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}
