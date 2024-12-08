package testrunner

import (
	"bytes"
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
	Dir      string          // Directory to run tests in
	Args     []string        // Arguments to pass to go test
	Handlers []EventHandler  // Handlers for test events
	Logger   *zerolog.Logger // Optional logger for debug output
	Stdout   io.Writer       // Writer for JSON output, defaults to os.Stdout
}

// AddHandler adds an event handler to the runner
func (r *Runner) AddHandler(handler EventHandler) {
	r.Handlers = append(r.Handlers, handler)
}

// Run executes go test with the given arguments and processes events
func (r *Runner) Run() error {
	if r.Dir != "" {
		if err := os.Chdir(r.Dir); err != nil {
			return fmt.Errorf("error changing directory: %w", err)
		}
	}

	args := []string{"test", "-json"}

	// Add the rest of the arguments, filtering out any -json flags
	for _, arg := range r.Args {
		if arg != "-json" {
			args = append(args, arg)
		}
	}

	r.Logger.Debug().
		Strs("args", args).
		Msg("Running go test")

	// Set up command
	cmd := exec.Command("go", args...)

	// Create pipe for stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %w", err)
	}

	// Capture stderr to buffer
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting command: %w", err)
	}

	// Default Stdout to os.Stdout if not set
	if r.Stdout == nil {
		r.Stdout = os.Stdout
	}

	// Process events
	done := make(chan error, 1)
	go func() {
		defer close(done)
		decoder := json.NewDecoder(stdout)
		encoder := json.NewEncoder(r.Stdout)

		for decoder.More() {
			var event TestEvent
			if err := decoder.Decode(&event); err != nil {
				done <- fmt.Errorf("error decoding test output: %w", err)
				return
			}

			// Process event through all handlers
			for _, handler := range r.Handlers {
				if err := handler.HandleEvent(event); err != nil {
					done <- fmt.Errorf("error handling event: %w", err)
					return
				}
			}

			// Write the event to Stdout
			if err := encoder.Encode(event); err != nil {
				done <- fmt.Errorf("error encoding event: %w", err)
				return
			}
		}
		done <- nil
	}()

	// Wait for command to finish
	if err := cmd.Wait(); err != nil {
		stderrOutput := stderr.String()
		r.Logger.Error().
			Str("stderr", stderrOutput).
			Int("exit_code", cmd.ProcessState.ExitCode()).
			Msg("Test command failed")
		return fmt.Errorf("test command failed: %w", err)
	}

	// Wait for event processing to finish
	if err := <-done; err != nil {
		return err
	}

	return nil
}
