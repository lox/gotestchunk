package ciparallel

import (
	"os"
	"strconv"
)

// Chunk represents a parallel chunk configuration
type Chunk struct {
	Index int // current chunk (1-based)
	Total int // total number of chunks
}

// Detect returns the chunk configuration from CI environment variables
// Returns nil if no CI environment is detected
func Detect() *Chunk {
	// Try each CI system in order
	envMappings := []struct {
		chunkEnv  string
		chunksEnv string
		offset    int // some CI systems are 0-based, others 1-based
	}{
		{"CI_NODE_INDEX", "CI_NODE_TOTAL", 1},                           // Knapsack/TravisCI/GitLab
		{"CIRCLE_NODE_INDEX", "CIRCLE_NODE_TOTAL", 0},                   // CircleCI
		{"BITBUCKET_PARALLEL_STEP", "BITBUCKET_PARALLEL_STEP_COUNT", 1}, // Bitbucket
		{"BUILDKITE_PARALLEL_JOB", "BUILDKITE_PARALLEL_JOB_COUNT", 0},   // Buildkite
		{"SEMAPHORE_CURRENT_JOB", "SEMAPHORE_JOB_COUNT", 1},             // Semaphore
	}

	for _, mapping := range envMappings {
		if chunkStr, chunksStr := os.Getenv(mapping.chunkEnv), os.Getenv(mapping.chunksEnv); chunkStr != "" && chunksStr != "" {
			if total, err := strconv.Atoi(chunksStr); err == nil {
				if index, err := strconv.Atoi(chunkStr); err == nil {
					return &Chunk{
						Index: index + mapping.offset,
						Total: total,
					}
				}
			}
		}
	}
	return nil
}
