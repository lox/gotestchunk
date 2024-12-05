package commands

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// captureOutput captures stdout during the execution of a function
func captureOutput(fn func() error) (string, error) {
	stdout := &strings.Builder{}
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", fmt.Errorf("failed to create pipe: %w", err)
	}
	os.Stdout = w

	// Create a channel to signal when copying is done
	done := make(chan error)
	go func() {
		_, err := io.Copy(stdout, r)
		done <- err
	}()

	// Run the function that generates output
	err = fn()

	// Close the write end of the pipe
	w.Close()

	// Wait for copying to complete
	if copyErr := <-done; copyErr != nil {
		return "", fmt.Errorf("error copying output: %w", copyErr)
	}

	// Restore stdout
	os.Stdout = oldStdout

	if err != nil {
		return "", err
	}

	return stdout.String(), nil
}
