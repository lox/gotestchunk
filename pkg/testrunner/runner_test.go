package testrunner

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/lox/gotestchunk/pkg/testlist"
	"github.com/rs/zerolog"
)

// TestEventCollector collects test events for testing
type TestEventCollector struct {
	Events []TestEvent
}

func (c *TestEventCollector) HandleEvent(event TestEvent) error {
	c.Events = append(c.Events, event)
	return nil
}

func TestRunner(t *testing.T) {
	// Capture stdout and stderr
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	// Create event collector
	collector := &TestEventCollector{}

	// Run tests in example package
	runner := &Runner{
		Args:    []string{"../example/..."},
		Handler: collector,
		Stdout:  stdout,
		Stderr:  stderr,
	}

	// Run the tests
	if err := runner.Run(); err != nil {
		t.Fatalf("Runner.Run() error = %v", err)
	}

	// Verify we got test output
	if stdout.Len() == 0 {
		t.Error("expected test output, got none")
	}

	// Check we got events for all our example tests
	expectedTests := map[string]bool{
		"TestSimple":                      false,
		"TestParallel":                    false,
		"TestTableDriven":                 false,
		"TestTableDriven/small_number":    false,
		"TestTableDriven/zero":            false,
		"TestTableDriven/negative":        false,
		"TestTableDriven/large_number":    false,
		"TestWithSetup":                   false,
		"TestMath":                        false,
		"TestMath/Multiply":               false,
		"TestMath/Divide":                 false,
		"TestDivideErrors":                false,
		"TestDivideErrors/divide_by_zero": false,
		"TestDivideErrors/slow_division":  false,
		"ExampleMultiply":                 false,
	}

	// Track which tests are parent tests (they won't have elapsed times)
	parentTests := map[string]bool{
		"TestTableDriven":                 true,
		"TestMath":                        true,
		"TestDivideErrors":                true,
		"TestDivideErrors/divide_by_zero": true,
	}

	// Check each event
	for _, event := range collector.Events {
		if event.Action == "pass" && event.Test != "" {
			if _, ok := expectedTests[event.Test]; !ok {
				t.Errorf("unexpected test: %s", event.Test)
			}
			expectedTests[event.Test] = true
		}
	}

	// Check we got all expected tests
	var missing []string
	for test, found := range expectedTests {
		if !found {
			missing = append(missing, test)
		}
	}
	if len(missing) > 0 {
		t.Errorf("missing test events: %s", strings.Join(missing, ", "))
	}

	// Verify we got other important event types
	eventTypes := make(map[string]bool)
	for _, event := range collector.Events {
		eventTypes[event.Action] = true
	}

	requiredEvents := []string{"run", "pass", "output"}
	for _, required := range requiredEvents {
		if !eventTypes[required] {
			t.Errorf("missing required event type: %s", required)
		}
	}

	// Check that elapsed times are present for non-parent passed tests
	for _, event := range collector.Events {
		if event.Action == "pass" && event.Test != "" {
			// Skip parent tests and examples
			if parentTests[event.Test] || strings.HasPrefix(event.Test, "Example") {
				continue
			}
			if event.Elapsed == 0 {
				t.Errorf("test %s has no elapsed time", event.Test)
			}
		}
	}
}

// TestRunnerFailure tests that the runner properly handles test failures
func TestRunnerFailure(t *testing.T) {
	// Create a failing test by passing an invalid package path
	runner := &Runner{
		Args: []string{"../does-not-exist"},
	}

	if err := runner.Run(); err == nil {
		t.Error("Runner.Run() expected error for non-existent package")
	}
}

// TestRunnerInvalidArgs tests that the runner properly handles invalid arguments
func TestRunnerInvalidArgs(t *testing.T) {
	runner := &Runner{
		Args: []string{"--invalid-flag"},
	}

	if err := runner.Run(); err == nil {
		t.Error("Runner.Run() expected error for invalid flag")
	}
}

func TestRunnerWithJSON(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		handler   bool
		wantJSON  bool
		wantError bool
	}{
		{
			name:     "handler adds json",
			args:     []string{"./pkg/example/..."},
			handler:  true,
			wantJSON: true,
		},
		{
			name:     "explicit json without handler",
			args:     []string{"-json", "./pkg/example/..."},
			handler:  false,
			wantJSON: true,
		},
		{
			name:     "both handler and explicit json",
			args:     []string{"-json", "./pkg/example/..."},
			handler:  true,
			wantJSON: true,
		},
		{
			name:      "invalid package",
			args:      []string{"-json", "./does-not-exist"},
			wantJSON:  true,
			wantError: true,
		},
	}

	for _, tt := range tests {
		testlist.TestRunWithModuleRoot(t, tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			var handler EventHandler
			if tt.handler {
				handler = &TestEventCollector{}
			}

			// Create logger that writes to test output
			logger := zerolog.New(zerolog.NewTestWriter(t)).Level(zerolog.DebugLevel)

			runner := &Runner{
				Args:    tt.args,
				Handler: handler,
				Stdout:  stdout,
				Stderr:  stderr,
				Logger:  logger,
			}

			err := runner.Run()
			if (err != nil) != tt.wantError {
				t.Errorf("Run() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if tt.wantError {
				return
			}

			if tt.wantJSON {
				// Look for JSON events in output
				foundJSON := false
				scanner := bufio.NewScanner(bytes.NewReader(stdout.Bytes()))
				for scanner.Scan() {
					line := scanner.Text()
					var event TestEvent
					if err := json.Unmarshal([]byte(line), &event); err == nil {
						foundJSON = true
						break
					}
				}
				if !foundJSON {
					t.Error("No valid JSON events found in output")
				}
			}
		})
	}
}
