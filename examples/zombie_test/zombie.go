package zombie_test

import (
	"fmt"
	"time"

	"github.com/Krishna-IITB/zeotap-durable-engine/engine"
)

// RunZombieTest simulates the "Crash after execution, before save" problem.
// This is the hardest thing to solve in distributed systems.
func RunZombieTest(ctx *engine.Context) error {
	fmt.Println("\n💀 Testing Zombie Step Problem")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("Goal: Ensure we don't re-run a step if the DB save failed,")
	fmt.Println("      OR ensure the user knows it's re-running.")

	// 1. Sanity check step
	_, err := engine.Step(ctx, "step_before_zombie", func() (string, error) {
		time.Sleep(500 * time.Millisecond)
		fmt.Println("    ✓ Setup step completed normally")
		return "ok", nil
	})
	if err != nil {
		return err
	}

	// 2. The Dangerous Step
	// We enable a flag in the engine that forces a sleep right before the DB commit.
	fmt.Println("\n🔴 ENABLING ZOMBIE TEST MODE")
	fmt.Println("   (Prepare to Ctrl+C when you see the warning!)")

	engine.TestZombieDelay = true

	_, err = engine.Step(ctx, "zombie_step", func() (string, error) {
		time.Sleep(500 * time.Millisecond)
		fmt.Println("    💀 ZOMBIE STEP: Logic executed! (DB Save pending...)")
		return "I am alive", nil
	})

	// Reset flag immediately so we don't mess up future steps
	engine.TestZombieDelay = false

	if err != nil {
		return err
	}

	// 3. Recovery verification
	// If we restart, we should reach here.
	_, err = engine.Step(ctx, "step_after_zombie", func() (string, error) {
		time.Sleep(500 * time.Millisecond)
		fmt.Println("    ✓ Recovery successful: Final step executed.")
		return "done", nil
	})

	fmt.Println("✅ Zombie test passed.")
	return err
}
