package testrunner

import (
	"bytes"
	"io"
	"path/filepath"
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
	collector := &TestEventCollector{}

	moduleRoot, err := testlist.GetModuleRoot()
	if err != nil {
		t.Fatalf("testlist.GetModuleRoot() error = %v", err)
	}

	logger := zerolog.New(zerolog.NewTestWriter(t)).
		Level(zerolog.DebugLevel)

	// Run tests in example package
	runner := &Runner{
		Dir:    moduleRoot,
		Args:   []string{filepath.Join(moduleRoot, "pkg/example/...")},
		Logger: &logger,
	}
	runner.AddHandler(collector)

	// Run the tests
	if err := runner.Run(); err != nil {
		t.Fatalf("Runner.Run() error = %v", err)
	}

	// Verify we got test output
	if len(collector.Events) == 0 {
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
	logger := zerolog.New(zerolog.NewTestWriter(t)).
		Level(zerolog.DebugLevel)

	// Create a failing test by passing an invalid package path
	runner := &Runner{
		Args:   []string{"../does-not-exist"},
		Logger: &logger,
	}

	if err := runner.Run(); err == nil {
		t.Error("Runner.Run() expected error for non-existent package")
	}
}

// TestRunnerInvalidArgs tests that the runner properly handles invalid arguments
func TestRunnerInvalidArgs(t *testing.T) {
	logger := zerolog.New(zerolog.NewTestWriter(t)).
		Level(zerolog.DebugLevel)

	runner := &Runner{
		Args:   []string{"--invalid-flag"},
		Logger: &logger,
	}

	if err := runner.Run(); err == nil {
		t.Error("Runner.Run() expected error for invalid flag")
	}
}

func TestRunnerWithJSON(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		stdout    io.Writer
		handler   bool
		wantError bool
	}{
		{
			name:    "with handler and stdout",
			args:    []string{"./pkg/example/..."},
			stdout:  &bytes.Buffer{},
			handler: true,
		},
		{
			name:    "with stdout only",
			args:    []string{"./pkg/example/..."},
			stdout:  &bytes.Buffer{},
			handler: false,
		},
		{
			name:    "without handler or stdout",
			args:    []string{"./pkg/example/..."},
			handler: false,
		},
		{
			name:      "invalid package",
			args:      []string{"./does-not-exist"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		testlist.TestRunWithModuleRoot(t, tt.name, func(t *testing.T) {
			logger := zerolog.New(zerolog.NewTestWriter(t)).
				Level(zerolog.DebugLevel)

			var collector *TestEventCollector
			runner := &Runner{
				Args:   tt.args,
				Logger: &logger,
				Stdout: tt.stdout,
			}

			if tt.handler {
				collector = &TestEventCollector{}
				runner.AddHandler(collector)
			}

			err := runner.Run()
			if (err != nil) != tt.wantError {
				t.Errorf("Run() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if tt.wantError {
				return
			}

			if tt.handler && len(collector.Events) == 0 {
				t.Error("No events collected")
			}

			if tt.stdout != nil {
				var stdout bytes.Buffer
				if _, ok := tt.stdout.(*bytes.Buffer); ok {
					stdout = *tt.stdout.(*bytes.Buffer)
				} else {
					t.Errorf("Stdout not a buffer")
				}

				if stdout.Len() == 0 {
					t.Error("No JSON output written")
				}
			}
		})
	}
}
