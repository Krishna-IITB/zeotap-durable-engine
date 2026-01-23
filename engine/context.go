package engine

import (
	"sync"
	"sync/atomic"
)

// Context holds the state for a single workflow execution run.
// It keeps track of the sequence ID to ensure steps are deterministic.
type Context struct {
	WorkflowID string
	storage    *Storage

	// sequenceID is used to generate unique keys for steps based on order.
	// We use atomic operations here just in case, though usually steps run sequentially.
	sequenceID int64

	// mu protects the shared resources if we ever extend this context
	mu sync.Mutex
}

func NewContext(workflowID string, storage *Storage) *Context {
	return &Context{
		WorkflowID: workflowID,
		storage:    storage,
		sequenceID: 0,
	}
}

// getNextSequence increments the counter. Crucial for ensuring
// that "Step 2" is always identified as "Step 2" in the DB.
func (c *Context) getNextSequence() int64 {
	return atomic.AddInt64(&c.sequenceID, 1)
}
