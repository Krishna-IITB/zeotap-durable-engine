package onboarding

import (
	"fmt"
	"time"

	"github.com/Krishna-IITB/zeotap-durable-engine/engine"
)

type Employee struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Laptop string `json:"laptop"`
	Email  string `json:"email"`
}

// RunOnboarding demonstrates a "Real World" workflow.
// It mixes sequential steps with parallel execution (Fan-out / Fan-in).
func RunOnboarding(ctx *engine.Context) error {

	// 1. Create Employee (Sequential)
	employee, err := engine.Step(ctx, "create_record", func() (Employee, error) {
		time.Sleep(1 * time.Second)
		emp := Employee{
			Name: "John Doe",
			ID:   "EMP001",
		}
		fmt.Printf("  📝 [HR] Created employee record: %s\n", emp.Name)
		return emp, nil
	})
	if err != nil {
		return err
	}

	// 2. Provisioning Phase (PARALLEL)
	// We want to request a laptop and setup email at the same time.
	// If one fails, the whole workflow pauses.

	laptopCh := make(chan string, 1)
	emailCh := make(chan string, 1)
	errCh := make(chan error, 2) // buffer size 2 to prevent blocking

	// Goroutine A: IT Hardware
	go func() {
		laptop, err := engine.Step(ctx, "provision_laptop", func() (string, error) {
			time.Sleep(2 * time.Second) // Simulate warehouse delay
			item := "MacBook Pro M2"
			fmt.Printf("  💻 [IT] Provisioned hardware: %s\n", item)
			return item, nil
		})
		if err != nil {
			errCh <- err
			return
		}
		laptopCh <- laptop
	}()

	// Goroutine B: IT Security
	go func() {
		email, err := engine.Step(ctx, "setup_access", func() (string, error) {
			time.Sleep(2 * time.Second) // Simulate sysadmin delay
			addr := employee.ID + "@company.com"
			fmt.Printf("  🔑 [SEC] Generated email: %s\n", addr)
			return addr, nil
		})
		if err != nil {
			errCh <- err
			return
		}
		emailCh <- email
	}()

	// Fan-in: Wait for both
	// We need 2 successes to proceed.
	completed := 0
	for completed < 2 {
		select {
		case err := <-errCh:
			return err // Fail fast if either branch explodes
		case l := <-laptopCh:
			employee.Laptop = l
			completed++
		case e := <-emailCh:
			employee.Email = e
			completed++
		}
	}

	// 3. Final Notification (Sequential)
	// Only runs after both parallel steps succeed.
	_, err = engine.Step(ctx, "send_welcome_email", func() (bool, error) {
		time.Sleep(1 * time.Second)
		fmt.Printf("  📧 [MAIL] Sending welcome pack to %s\n", employee.Email)
		fmt.Printf("  🎉 Onboarding fully complete for: %s\n", employee.Name)
		return true, nil
	})

	return err
}
