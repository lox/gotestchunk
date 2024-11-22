package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/lox/gotestchunk/pkg/ciparallel"
)

// makeRunPattern converts a list of test names into a -run pattern
func makeRunPattern(tests []string) string {
	testNames := make([]string, 0, len(tests))
	for _, test := range tests {
		parts := strings.Split(test, ".")
		if len(parts) == 2 {
			testNames = append(testNames, parts[1])
		}
	}
	return fmt.Sprintf("^(%s)$", strings.Join(testNames, "|"))
}

type ChunkCmd struct {
	Chunks   int      `help:"Number of chunks to split tests into (defaults to CI value if available)" default:"1"`
	Chunk    int      `help:"Which chunk to output (1-based, defaults to CI value if available)" default:"1"`
	Packages []string `arg:"" optional:"" help:"Packages to list tests from" type:"path"`
}

func (cmd *ChunkCmd) Validate() error {
	// Check for CI environment variables first
	if chunk := ciparallel.Detect(); chunk != nil {
		cmd.Chunk = chunk.Index
		cmd.Chunks = chunk.Total
	}

	if cmd.Chunks < 1 {
		return fmt.Errorf("chunks must be >= 1")
	}
	if cmd.Chunk < 1 || cmd.Chunk > cmd.Chunks {
		return fmt.Errorf("chunk must be between 1 and chunks")
	}
	return nil
}

func (cmd *ChunkCmd) Run() error {
	// Default to current directory if no package specified
	pkgPath := "."
	if len(cmd.Packages) > 0 {
		pkgPath = cmd.Packages[0]
	}

	// Get all tests
	tests, err := listTests(pkgPath)
	if err != nil {
		return fmt.Errorf("error listing tests: %w", err)
	}

	// Sort tests for deterministic chunking
	sort.Strings(tests)

	// Calculate chunk size and bounds
	chunkSize := (len(tests) + cmd.Chunks - 1) / cmd.Chunks
	start := (cmd.Chunk - 1) * chunkSize
	end := start + chunkSize
	if end > len(tests) {
		end = len(tests)
	}

	// Get tests for this chunk
	chunkTests := tests[start:end]
	if len(chunkTests) == 0 {
		return fmt.Errorf("no tests in chunk %d", cmd.Chunk)
	}

	// Output just the pattern
	fmt.Print(makeRunPattern(chunkTests))
	return nil
}

type ChunkPackagesCmd struct {
	Chunks   int      `help:"Number of chunks to split tests into (defaults to CI value if available)" default:"1"`
	Chunk    int      `help:"Which chunk to output (1-based, defaults to CI value if available)" default:"1"`
	Packages []string `arg:"" optional:"" help:"Packages to list tests from" type:"path"`
}

func (cmd *ChunkPackagesCmd) Validate() error {
	// Check for CI environment variables first
	if chunk := ciparallel.Detect(); chunk != nil {
		cmd.Chunk = chunk.Index
		cmd.Chunks = chunk.Total
	}

	if cmd.Chunks < 1 {
		return fmt.Errorf("chunks must be >= 1")
	}
	if cmd.Chunk < 1 || cmd.Chunk > cmd.Chunks {
		return fmt.Errorf("chunk must be between 1 and chunks")
	}
	return nil
}

func (cmd *ChunkPackagesCmd) Run() error {
	// Default to current directory if no package specified
	pkgPath := "."
	if len(cmd.Packages) > 0 {
		pkgPath = cmd.Packages[0]
	}

	// Get all tests
	tests, err := listTests(pkgPath)
	if err != nil {
		return fmt.Errorf("error listing tests: %w", err)
	}

	// Sort tests for deterministic chunking
	sort.Strings(tests)

	// Calculate chunk size and bounds
	chunkSize := (len(tests) + cmd.Chunks - 1) / cmd.Chunks
	start := (cmd.Chunk - 1) * chunkSize
	end := start + chunkSize
	if end > len(tests) {
		end = len(tests)
	}

	// Get tests for this chunk
	chunkTests := tests[start:end]
	if len(chunkTests) == 0 {
		return fmt.Errorf("no tests in chunk %d", cmd.Chunk)
	}

	// Get unique packages
	packages := make(map[string]bool)
	for _, test := range chunkTests {
		parts := strings.Split(test, ".")
		if len(parts) == 2 {
			packages[parts[0]] = true
		}
	}

	// Output package paths
	for pkg := range packages {
		fmt.Printf("./%s ", pkg)
	}
	return nil
}
