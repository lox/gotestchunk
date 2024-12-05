package testlist

import (
	"fmt"
	"math"
)

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
