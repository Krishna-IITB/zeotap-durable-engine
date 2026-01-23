// Zeotap Durable Execution Engine CLI
//
// Demonstrates a workflow engine that can survive process crashes
// and resume exactly where it left off using SQLite for state.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Krishna-IITB/zeotap-durable-engine/engine"
	"github.com/Krishna-IITB/zeotap-durable-engine/examples/conditional_test"
	"github.com/Krishna-IITB/zeotap-durable-engine/examples/loop_test"
	"github.com/Krishna-IITB/zeotap-durable-engine/examples/onboarding"
	"github.com/Krishna-IITB/zeotap-durable-engine/examples/zombie_test"
)

func main() {
	// Simple CLI args parsing
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	workflowID := os.Args[1]
	testType := "onboarding" // default
	if len(os.Args) >= 3 {
		testType = os.Args[2]
	}

	// ---------------------------------------------------------
	// Crash Simulation Handler
	// We listen for Ctrl+C so we can print a helpful message
	// instead of just exiting silently.
	// ---------------------------------------------------------
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n\n⚠️  Process Interrupted! (Crash Simulated)")
		fmt.Println("💡 Run the same command again to RESUME from the last saved step.")
		fmt.Println("   The DB has your back.")
		os.Exit(0)
	}()

	// Init Engine
	// Using a local file for easy cleanup/inspection
	dbPath := "./workflow.db"
	wf, err := engine.NewWorkflow(dbPath)
	if err != nil {
		fmt.Printf("❌ Failed to init engine: %v\n", err)
		os.Exit(1)
	}
	defer wf.Close()

	// Router
	var runErr error
	switch testType {
	case "loop":
		runErr = wf.Run(workflowID, loop_test.RunLoopTest)
	case "conditional":
		runErr = wf.Run(workflowID, conditional_test.RunConditionalTest)
	case "zombie":
		runErr = wf.Run(workflowID, zombie_test.RunZombieTest)
	case "onboarding":
		runErr = wf.Run(workflowID, onboarding.RunOnboarding)
	default:
		fmt.Printf("❌ Unknown test type: %s\n\n", testType)
		printUsage()
		os.Exit(1)
	}

	if runErr != nil {
		fmt.Printf("❌ Execution stopped: %v\n", runErr)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║     Zeotap Durable Execution Engine                      ║")
	fmt.Println("║     Assignment: Software Engineer Intern                 ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Usage: ./app <workflow_id> [test_type]")
	fmt.Println()
	fmt.Println("Available Workflows:")
	fmt.Println("  onboarding   - Employee onboarding with parallel provisioning")
	fmt.Println("  loop         - Demonstrates loop iteration tracking")
	fmt.Println("  conditional  - Tests branching logic (if/else)")
	fmt.Println("  zombie       - Crash-before-save recovery test")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  ./app user_123 onboarding     # Run employee onboarding")
	fmt.Println("  ./app test_001 loop           # Test loop support")
	fmt.Println("  ./app demo conditional        # Test conditional logic")
	fmt.Println()
	fmt.Println("Crash Recovery:")
	fmt.Println("  Press Ctrl+C during execution to simulate crash")
	fmt.Println("  Re-run same command to resume from last checkpoint")
	fmt.Println()
	fmt.Println("Database:")
	fmt.Println("  Location: ./workflow.db (SQLite)")
	fmt.Println("  Inspect: sqlite3 workflow.db 'SELECT * FROM steps;'")
	fmt.Println()
}
