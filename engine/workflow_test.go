package engine

import (
	"os"
	"testing"
)

// Helper to clean up DBs after tests
func cleanup(path string) {
	os.Remove(path)
}

func TestStepMemoization(t *testing.T) {
	dbPath := "test_memo.db"
	cleanup(dbPath)
	defer cleanup(dbPath)

	wf, err := NewWorkflow(dbPath)
	if err != nil {
		t.Fatal("Could not start engine:", err)
	}
	defer wf.Close()

	execCount := 0

	// Define the logic we want to run
	workLogic := func(ctx *Context) error {
		_, err := Step(ctx, "expensive_step", func() (int, error) {
			execCount++
			return 42, nil
		})
		return err
	}

	// 1. First run: Should execute
	if err := wf.Run("test_run_1", workLogic); err != nil {
		t.Fatal(err)
	}

	// 2. Second run: Should skip (memoization)
	// We use the same workflowID ("test_run_1") to trigger the cache hit
	if err := wf.Run("test_run_1", workLogic); err != nil {
		t.Fatal(err)
	}

	// Assertions
	if execCount != 1 {
		t.Errorf("Memoization failed! Expected function to run 1 time, ran %d times", execCount)
	}
}

func TestConcurrentWrites(t *testing.T) {
	dbPath := "test_concurrent.db"
	cleanup(dbPath)
	defer cleanup(dbPath)

	wf, err := NewWorkflow(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer wf.Close()

	// Simulating parallel steps (e.g., using goroutines inside a workflow)
	// This ensures SQLite doesn't lock up with SQLITE_BUSY
	err = wf.Run("concurrent_test", func(ctx *Context) error {
		errChan := make(chan error, 2)

		// Fire off Step A
		go func() {
			_, err := Step(ctx, "step_a", func() (string, error) {
				return "A", nil
			})
			errChan <- err
		}()

		// Fire off Step B
		go func() {
			_, err := Step(ctx, "step_b", func() (string, error) {
				return "B", nil
			})
			errChan <- err
		}()

		// Wait for both
		for i := 0; i < 2; i++ {
			if err := <-errChan; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		t.Errorf("Concurrent workflow failed: %v", err)
	}
}
