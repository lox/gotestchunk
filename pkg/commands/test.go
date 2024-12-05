package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/lox/gotestchunk/pkg/ciparallel"
	"github.com/lox/gotestchunk/pkg/testlist"
	"github.com/rs/zerolog"
)

type TestCmd struct {
	Chunks    int      `help:"Number of chunks to split tests into (defaults to CI value if available)" default:"1"`
	Chunk     int      `help:"Which chunk to output (1-based, defaults to CI value if available)" default:"1"`
	Count     int      `help:"Number of times to run each test" default:"0"`
	JSON      bool     `help:"Output test results in JSON format" default:"false"`
	Verbose   bool     `short:"v" help:"Verbose output" default:"false"`
	Gotestsum bool     `help:"Use gotestsum if available" default:"false"`
	Args      []string `arg:"" optional:"" passthrough:"" help:"Packages to test, followed by optional -- and test arguments"`
}

func (cmd *TestCmd) Validate() error {
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

func (cmd *TestCmd) Run(logger *zerolog.Logger) error {
	logger.Info().
		Strs("args", cmd.Args).
		Msg("Running test command")

	// Split Args into packages and test args at --
	var packages []string
	var testArgs []string

	foundSeparator := false
	for i, arg := range cmd.Args {
		if arg == "--" {
			packages = cmd.Args[:i]
			testArgs = cmd.Args[i+1:]
			foundSeparator = true
			break
		}
	}

	// If no -- found, all args are packages
	if !foundSeparator {
		packages = cmd.Args
	}
	if len(packages) == 0 {
		packages = []string{"./..."}
	}

	// Get all tests
	tests, err := testlist.List(packages...)
	if err != nil {
		return fmt.Errorf("error listing tests: %w", err)
	}

	logger.Info().
		Int("tests", len(tests)).
		Msg("Found tests")

	// Get tests for this chunk using testlist.Chunk
	chunkTests, err := testlist.Chunk(tests, cmd.Chunk-1, cmd.Chunks) // -1 because Chunk is 0-based
	if err != nil {
		return fmt.Errorf("error getting chunk: %w", err)
	}

	if len(chunkTests) == 0 {
		return fmt.Errorf("no tests in chunk %d", cmd.Chunk)
	}

	logger.Info().
		Str("chunk", strconv.Itoa(cmd.Chunk)).
		Int("tests", len(chunkTests)).
		Msg("Found chunk tests")

	// Build go test command args
	goTestArgs := []string{"test"}
	if cmd.JSON {
		goTestArgs = append(goTestArgs, "-json")
	}
	if cmd.Verbose {
		goTestArgs = append(goTestArgs, "-v")
	}
	if cmd.Count > 0 {
		goTestArgs = append(goTestArgs, "-count="+strconv.Itoa(cmd.Count))
	}

	// Add any test args before the test pattern
	if len(testArgs) > 0 {
		goTestArgs = append(goTestArgs, testArgs...)
	}

	// Add test pattern
	pattern, err := testlist.Format(chunkTests, "runPattern")
	if err != nil {
		return fmt.Errorf("error formatting test pattern: %w", err)
	}
	if pattern != "" {
		goTestArgs = append(goTestArgs, "-run="+pattern)
	}

	// Add packages
	for _, pkg := range testlist.Packages(chunkTests) {
		goTestArgs = append(goTestArgs, "./"+pkg)
	}

	logger.Info().
		Strs("args", goTestArgs).
		Msg("Running go test")

	// Run go test directly
	goCmd := exec.Command("go", goTestArgs...)
	goCmd.Stdout = os.Stdout
	goCmd.Stderr = os.Stderr
	return goCmd.Run()
}
