package timing

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lox/gotestchunk/pkg/testrunner"
)

func TestCollector(t *testing.T) {
	tests := []struct {
		name     string
		event    testrunner.TestEvent
		wantTest *Test
	}{
		{
			name: "pass event with elapsed time",
			event: testrunner.TestEvent{
				Action:  "pass",
				Package: "github.com/lox/gotestchunk/pkg/example",
				Test:    "TestSimple",
				Elapsed: 1.5,
			},
			wantTest: &Test{
				Package: "pkg/example",
				Test:    "TestSimple",
				Time:    1500 * time.Millisecond,
			},
		},
		{
			name: "pass event with no elapsed time",
			event: testrunner.TestEvent{
				Action:  "pass",
				Package: "github.com/lox/gotestchunk/pkg/example",
				Test:    "TestParent",
				Elapsed: 0,
			},
			wantTest: nil,
		},
		{
			name: "non-pass event",
			event: testrunner.TestEvent{
				Action:  "run",
				Package: "github.com/lox/gotestchunk/pkg/example",
				Test:    "TestSimple",
				Elapsed: 1.5,
			},
			wantTest: nil,
		},
		{
			name: "pass event with no test name",
			event: testrunner.TestEvent{
				Action:  "pass",
				Package: "github.com/lox/gotestchunk/pkg/example",
				Test:    "",
				Elapsed: 1.5,
			},
			wantTest: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := &Collector{}
			if err := collector.HandleEvent(tt.event); err != nil {
				t.Fatalf("HandleEvent() error = %v", err)
			}

			if tt.wantTest == nil {
				if len(collector.Tests) > 0 {
					t.Errorf("HandleEvent() collected test when it shouldn't: %+v", collector.Tests[0])
				}
				return
			}

			if len(collector.Tests) != 1 {
				t.Fatalf("HandleEvent() collected %d tests, want 1", len(collector.Tests))
			}

			got := collector.Tests[0]
			if got.Package != tt.wantTest.Package {
				t.Errorf("Package = %v, want %v", got.Package, tt.wantTest.Package)
			}
			if got.Test != tt.wantTest.Test {
				t.Errorf("Test = %v, want %v", got.Test, tt.wantTest.Test)
			}
			if got.Time != tt.wantTest.Time {
				t.Errorf("Time = %v, want %v", got.Time, tt.wantTest.Time)
			}
		})
	}
}

func TestFileOperations(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "timing-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test data
	tests := []Test{
		{
			Package: "pkg/example",
			Test:    "TestSimple",
			Time:    100 * time.Millisecond,
		},
		{
			Package: "pkg/example",
			Test:    "TestParallel",
			Time:    200 * time.Millisecond,
		},
	}

	// Write test data to files
	files := []string{
		filepath.Join(tmpDir, "timing-1.json"),
		filepath.Join(tmpDir, "timing-2.json"),
	}
	for _, file := range files {
		if err := WriteToFile(tests, file); err != nil {
			t.Fatalf("WriteToFile() error = %v", err)
		}
	}

	// Load and verify the data
	timings, err := LoadFromFiles(files)
	if err != nil {
		t.Fatalf("LoadFromFiles() error = %v", err)
	}

	// Check we got the expected timings
	expectedTests := map[string]time.Duration{
		"pkg/example.TestSimple":   100 * time.Millisecond,
		"pkg/example.TestParallel": 200 * time.Millisecond,
	}

	if len(timings) != len(expectedTests) {
		t.Errorf("got %d timings, want %d", len(timings), len(expectedTests))
	}

	for key, want := range expectedTests {
		got, ok := timings[key]
		if !ok {
			t.Errorf("missing timing for %s", key)
			continue
		}
		if got != want {
			t.Errorf("timing for %s = %v, want %v", key, got, want)
		}
	}
}

func TestLoadFromFilesError(t *testing.T) {
	// Try to load non-existent files
	_, err := LoadFromFiles([]string{"does-not-exist.json"})
	if err == nil {
		t.Error("LoadFromFiles() expected error for non-existent file")
	}

	// Try to load invalid JSON
	tmpFile := filepath.Join(t.TempDir(), "invalid.json")
	if err := os.WriteFile(tmpFile, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err = LoadFromFiles([]string{tmpFile})
	if err == nil {
		t.Error("LoadFromFiles() expected error for invalid JSON")
	}
}
