package testlist

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// GetModuleRoot returns the absolute path to the module root
func GetModuleRoot() (string, error) {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	// Walk up until we find go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find go.mod")
		}
		dir = parent
	}
}

// TestRunWithModuleRoot runs the given test function after changing to the module root directory,
// and restores the original working directory afterwards.
func TestRunWithModuleRoot(t *testing.T, name string, f func(t *testing.T)) {
	t.Helper()

	t.Run(name, func(t *testing.T) {
		// Save current directory
		origDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("failed to get working directory: %v", err)
		}

		moduleRoot, err := GetModuleRoot()
		if err != nil {
			t.Fatalf("failed to get module root: %v", err)
		}

		// Change to module root
		if err := os.Chdir(moduleRoot); err != nil {
			t.Fatalf("failed to change to module root: %v", err)
		}

		// Restore original directory after test
		defer func() {
			if err := os.Chdir(origDir); err != nil {
				t.Errorf("failed to restore working directory: %v", err)
			}
		}()

		f(t)
	})
}
