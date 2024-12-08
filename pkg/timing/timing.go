package timing

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/lox/gotestchunk/pkg/testrunner"
)

// Test represents timing information for a single test
type Test struct {
	Package string        `json:"package"`
	Test    string        `json:"test"`
	Time    time.Duration `json:"time"`
}

// Collector collects test timing information
type Collector struct {
	Tests []Test
}

// HandleEvent processes a test event
func (c *Collector) HandleEvent(event testrunner.TestEvent) error {
	// Only process "pass" events with non-zero elapsed time
	if event.Action != "pass" || event.Test == "" || event.Elapsed == 0 {
		return nil
	}

	// Trim module path from package name
	pkg := event.Package
	if idx := strings.Index(pkg, "/pkg/"); idx != -1 {
		pkg = pkg[idx+1:]
	}

	c.Tests = append(c.Tests, Test{
		Package: pkg,
		Test:    event.Test,
		Time:    time.Duration(event.Elapsed * float64(time.Second)),
	})
	return nil
}

// LoadFromFiles loads and aggregates timing data from the given files
func LoadFromFiles(files []string) (map[string]time.Duration, error) {
	// Map of test name to average duration
	timings := make(map[string]struct {
		total time.Duration
		count int
	})

	// Load and aggregate data from each file
	for _, file := range files {
		var fileTimings []Test
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("error reading %s: %w", file, err)
		}

		if err := json.Unmarshal(data, &fileTimings); err != nil {
			return nil, fmt.Errorf("error parsing %s: %w", file, err)
		}

		for _, t := range fileTimings {
			key := t.Package + "." + t.Test
			entry := timings[key]
			entry.total += t.Time
			entry.count++
			timings[key] = entry
		}
	}

	// Calculate averages
	result := make(map[string]time.Duration)
	for key, timing := range timings {
		result[key] = timing.total / time.Duration(timing.count)
	}

	return result, nil
}

// WriteToFile writes test timing data to a JSON file
func WriteToFile(tests []Test, filename string) error {
	data, err := json.MarshalIndent(tests, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling timing data: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("error writing timing file: %w", err)
	}

	return nil
}
