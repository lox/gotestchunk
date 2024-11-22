package example

import (
	"testing"
	"time"
)

// Simple test
func TestSimple(t *testing.T) {
	t.Log("running simple test")
	time.Sleep(100 * time.Millisecond)
}

// Parallel test
func TestParallel(t *testing.T) {
	t.Parallel()
	t.Log("running parallel test")
	time.Sleep(200 * time.Millisecond)
}

// Table test with subtests
func TestTableDriven(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"small number", 5, 25},
		{"zero", 0, 0},
		{"negative", -2, 4},
		{"large number", 100, 10000},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			time.Sleep(150 * time.Millisecond)
			result := tt.input * tt.input
			if result != tt.expected {
				t.Errorf("got %d, want %d", result, tt.expected)
			}
		})
	}
}

// Test with setup and cleanup
func TestWithSetup(t *testing.T) {
	// Setup
	t.Log("setting up")
	time.Sleep(50 * time.Millisecond)

	t.Cleanup(func() {
		t.Log("cleaning up")
		time.Sleep(50 * time.Millisecond)
	})

	t.Log("running test")
	time.Sleep(100 * time.Millisecond)
}
