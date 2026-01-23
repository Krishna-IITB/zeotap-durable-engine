package conditional_test

import (
	"fmt"
	"time"

	"github.com/Krishna-IITB/zeotap-durable-engine/engine"
)

// RunConditionalTest proves that the engine handles branching logic correctly.
// If we crash inside a branch, we need to make sure we don't accidentally
// execute the OTHER branch upon resumption.
func RunConditionalTest(ctx *engine.Context) error {
	fmt.Println("\n🔀 Testing Conditional Logic Support")

	// Step 1: Decide which path to take
	condition, err := engine.Step(ctx, "check_condition", func() (bool, error) {
		time.Sleep(500 * time.Millisecond)
		fmt.Println("    🤔 Evaluating condition...")
		return true, nil // Hardcoded to true for this demo
	})
	if err != nil {
		return err
	}

	// Step 2: Branch Execution
	// The engine sequence ID ensures these steps don't conflict even though
	// they might appear at the same "depth" in the code.
	if condition {
		_, err = engine.Step(ctx, "true_branch", func() (string, error) {
			time.Sleep(500 * time.Millisecond)
			fmt.Println("    ✓ [TRUE PATH] executed")
			return "true_executed", nil
		})
	} else {
		_, err = engine.Step(ctx, "false_branch", func() (string, error) {
			time.Sleep(500 * time.Millisecond)
			fmt.Println("    ✓ [FALSE PATH] executed")
			return "false_executed", nil
		})
	}

	if err != nil {
		return err
	}

	fmt.Println("✅ Conditional test completed")
	return nil
}
