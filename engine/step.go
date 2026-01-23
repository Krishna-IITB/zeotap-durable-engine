package engine

import (
	"encoding/json"
	"fmt"
	"time"
)

// Dev Note: This is a hack for integration testing.
// Allows us to kill the process mid-step to verify DB persistence.
var TestZombieDelay = false

// Step is the core wrapper. It handles the "check DB -> run -> save DB" loop.
// T is generic so we can return actual types (int, string, structs) instead of interface{}.
func Step[T any](ctx *Context, id string, fn func() (T, error)) (T, error) {
	var zero T // return this on error

	// 1. Deterministic Key Generation
	// We combine the user-friendly name (id) with the sequence number.
	seq := ctx.getNextSequence()
	stepKey := fmt.Sprintf("%s_%d", id, seq)

	// 2. Check for Replay (Memoization)
	ctx.mu.Lock()
	record, err := ctx.storage.GetStep(ctx.WorkflowID, stepKey)
	ctx.mu.Unlock()

	if err != nil {
		return zero, fmt.Errorf("DB error checking step '%s': %w", id, err)
	}

	// If we found a completed record, unmarshal and skip execution.
	if record != nil && record.Status == "completed" {
		var result T
		if err := json.Unmarshal([]byte(record.Output), &result); err != nil {
			return zero, fmt.Errorf("corrupt data in DB for step '%s': %w", id, err)
		}

		fmt.Printf("✓ [SKIP] Step '%s' (key: %s) already done.\n", id, stepKey)
		return result, nil
	}

	// 3. Execution
	// If we are here, it's a fresh run or a retry.
	fmt.Printf("→ [RUN]  Executing step '%s' (key: %s)...\n", id, stepKey)

	result, err := fn()
	if err != nil {
		return zero, fmt.Errorf("logic failed in step '%s': %w", id, err)
	}

	// DEBUG: Simulate a crash right before saving to test atomicity/resuming.
	if TestZombieDelay {
		fmt.Printf("\n⚠️  [TEST MODE] Sleeping 3s. KILL PROCESS NOW (Ctrl+C) to test zombie recovery!\n\n")
		time.Sleep(3 * time.Second)
	}

	// 4. Persistence
	// Save the result so we don't run this again.
	ctx.mu.Lock()
	err = ctx.storage.SaveStep(ctx.WorkflowID, stepKey, "completed", result)
	ctx.mu.Unlock()

	if err != nil {
		return zero, fmt.Errorf("FATAL: could not save step result: %w", err)
	}

	fmt.Printf("✓ [DONE] Step '%s' saved.\n", id)
	return result, nil
}
