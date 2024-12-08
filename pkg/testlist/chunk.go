package testlist

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// TestTiming represents timing information for a single test
type TestTiming struct {
	Package string        `json:"package"`
	Test    string        `json:"test"`
	Time    time.Duration `json:"time"`
}

// LoadTimings loads and aggregates timing data from multiple files
func LoadTimings(pattern string) (map[string]time.Duration, error) {
	// Find all matching files
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("error finding timing files: %w", err)
	}

	// Map of test name to average duration
	timings := make(map[string]struct {
		total time.Duration
		count int
	})

	// Load and aggregate data from each file
	for _, file := range files {
		var fileTimings []TestTiming
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

// ChunkByTiming splits tests into chunks trying to balance total execution time
func ChunkByTiming(tests []Test, index, total int, timings map[string]time.Duration) ([]Test, error) {
	if index < 0 || index >= total {
		return nil, fmt.Errorf("chunk index %d out of bounds (total chunks: %d)", index, total)
	}

	// First sort tests to ensure consistent chunking
	Sort(tests)

	// Create chunks with timing information
	type testWithTiming struct {
		test Test
		time time.Duration
	}

	// Get timing for each test, using default if not found
	defaultTime := time.Second // default duration for tests without timing data
	testsWithTiming := make([]testWithTiming, len(tests))
	for i, test := range tests {
		key := test.Package + "." + test.Name
		time, ok := timings[key]
		if !ok {
			time = defaultTime
		}
		testsWithTiming[i] = testWithTiming{test, time}
	}

	// Sort by duration (longest first)
	sort.Slice(testsWithTiming, func(i, j int) bool {
		return testsWithTiming[i].time > testsWithTiming[j].time
	})

	// Create chunks and track their total times
	chunks := make([][]Test, total)
	chunkTimes := make([]time.Duration, total)

	// Distribute tests using a greedy algorithm
	for _, t := range testsWithTiming {
		// Find chunk with smallest total time
		minIndex := 0
		minTime := chunkTimes[0]
		for i := 1; i < total; i++ {
			if chunkTimes[i] < minTime {
				minIndex = i
				minTime = chunkTimes[i]
			}
		}

		chunks[minIndex] = append(chunks[minIndex], t.test)
		chunkTimes[minIndex] += t.time
	}

	return chunks[index], nil
}

// Chunks splits a slice of tests into n roughly equal chunks
func Chunks(tests []Test, total int) [][]Test {
	if total <= 0 {
		total = 1
	}
	if len(tests) == 0 {
		return make([][]Test, total)
	}

	// First sort tests to ensure consistent chunking
	Sort(tests)

	// Calculate chunk sizes
	chunkSize := int(math.Ceil(float64(len(tests)) / float64(total)))
	result := make([][]Test, 0, total)

	for i := 0; i < len(tests); i += chunkSize {
		end := i + chunkSize
		if end > len(tests) {
			end = len(tests)
		}
		result = append(result, tests[i:end])
	}

	// If we have fewer chunks than requested, pad with empty slices
	for len(result) < total {
		result = append(result, []Test{})
	}

	return result
}

// Chunk returns a specific chunk of tests given an index and total number of chunks
func Chunk(tests []Test, index, total int) ([]Test, error) {
	if index < 0 || index >= total {
		return nil, fmt.Errorf("chunk index %d out of bounds (total chunks: %d)", index, total)
	}

	chunks := Chunks(tests, total)
	return chunks[index], nil
}
