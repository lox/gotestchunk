package testrunner

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/rs/zerolog"
)

// EventHandler processes test events from go test -json output
type EventHandler interface {
	HandleEvent(event TestEvent) error
}

// TestEvent represents a single event from go test -json output
type TestEvent struct {
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test"`
	Elapsed float64 `json:"Elapsed"`
	Output  string  `json:"Output"`
}

// Runner executes go test and processes the output
type Runner struct {
	Args    []string       // Arguments to pass to go test
	Handler EventHandler   // Handler for test events
	Stdout  io.Writer      // Where to write test output (defaults to os.Stdout)
	Stderr  io.Writer      // Where to write test errors (defaults to os.Stderr)
	Logger  zerolog.Logger // Optional logger for debug output
}

// Run executes go test with the given arguments and processes events
func (r *Runner) Run() error {
	args := []string{"test"}

	// Check if -json is already in the arguments
	hasJSON := false
	for _, arg := range r.Args {
		if arg == "-json" {
			hasJSON = true
			break
		}
	}

	// Add -json if we have a handler and it's not already present
	if r.Handler != nil && !hasJSON {
		args = append(args, "-json")
	}

	// Add the rest of the arguments
	args = append(args, r.Args...)

	r.Logger.Debug().
		Strs("args", args).
		Bool("hasJSON", hasJSON).
		Bool("hasHandler", r.Handler != nil).
		Msg("Running go test")

	// Set up command
	cmd := exec.Command("go", args...)

	// Create pipe for stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %w", err)
	}

	// Set stderr
	if r.Stderr != nil {
		cmd.Stderr = r.Stderr
	} else {
		cmd.Stderr = os.Stderr
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting command: %w", err)
	}

	// Process events in a goroutine if we have a handler or JSON output is requested
	var done chan error
	if r.Handler != nil || hasJSON {
		done = make(chan error, 1)
		go func() {
			defer close(done)
			decoder := json.NewDecoder(stdout)

			for decoder.More() {
				var event TestEvent
				if err := decoder.Decode(&event); err != nil {
					done <- fmt.Errorf("error decoding test output: %w", err)
					return
				}

				// Write output if requested
				if r.Stdout != nil && event.Output != "" {
					if _, err := fmt.Fprint(r.Stdout, event.Output); err != nil {
						done <- fmt.Errorf("error writing output: %w", err)
						return
					}
				}

				// Process event if we have a handler
				if r.Handler != nil {
					if err := r.Handler.HandleEvent(event); err != nil {
						done <- fmt.Errorf("error handling event: %w", err)
						return
					}
				}
			}
			done <- nil
		}()
	} else if r.Stdout != nil {
		// If no JSON handling needed, just copy stdout directly
		_, err := io.Copy(r.Stdout, stdout)
		if err != nil {
			return fmt.Errorf("error copying stdout: %w", err)
		}
	}

	// Wait for command to finish
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("test command failed: %w", err)
	}

	// Wait for event processing to finish if we're handling JSON
	if done != nil {
		if err := <-done; err != nil {
			return err
		}
	}

	return nil
}
