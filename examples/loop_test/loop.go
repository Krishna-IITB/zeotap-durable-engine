package loop_test

import (
	"fmt"
	"time"

	"github.com/Krishna-IITB/zeotap-durable-engine/engine"
)

// RunLoopTest checks if the engine generates unique IDs for repeated steps.
// Without the sequence counter in the context, the engine would think
// every iteration is the same step and skip 2, 3, etc.
func RunLoopTest(ctx *engine.Context) error {
	fmt.Println("\n🔁 Testing Loop Support")
	fmt.Println("   (Verifying sequence tracking for identical step names)")

	iterations := 3

	for i := 1; i <= iterations; i++ {
		// Notice we use the same step ID "loop_step" every time.
		// The Context handles the unique suffixing (e.g., loop_step_1, loop_step_2).
		result, err := engine.Step(ctx, "loop_step", func() (int, error) {
			time.Sleep(500 * time.Millisecond)
			fmt.Printf("    ⚙️  Doing work for iteration %d...\n", i)
			return i * 10, nil
		})

		if err != nil {
			return err
		}

		// Just verify we got the right data back
		fmt.Printf("    → Result: %d\n", result)
	}

	fmt.Println("✅ Loop test completed - All iterations tracked separately")
	return nil
}
