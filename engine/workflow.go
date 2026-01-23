package engine

import (
	"fmt"
)

type Workflow struct {
	storage *Storage
}

func NewWorkflow(dbPath string) (*Workflow, error) {
	storage, err := NewStorage(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to init storage: %w", err)
	}
	return &Workflow{storage: storage}, nil
}

// Run kicks off the workflow closure.
// It initializes the context and handles the top-level error reporting.
func (w *Workflow) Run(workflowID string, fn func(*Context) error) error {
	ctx := NewContext(workflowID, w.storage)

	// purely aesthetic logging
	fmt.Printf("\n🚀 Starting workflow: %s\n", workflowID)
	fmt.Println("─────────────────────────────────────")

	// Run the user's workflow function
	err := fn(ctx)

	fmt.Println("─────────────────────────────────────")
	if err != nil {
		fmt.Printf("❌ Workflow FAILED: %v\n\n", err)
		return err
	}

	fmt.Printf("✅ Workflow COMPLETED.\n\n")
	return nil
}

func (w *Workflow) Close() error {
	return w.storage.Close()
}
