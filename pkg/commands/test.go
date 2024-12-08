package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/lox/gotestchunk/pkg/ciparallel"
	"github.com/lox/gotestchunk/pkg/testlist"
	"github.com/lox/gotestchunk/pkg/testrunner"
	"github.com/lox/gotestchunk/pkg/timing"
	"github.com/rs/zerolog"
)

type TestCmd struct {
	Chunks      int      `help:"Number of chunks to split tests into (defaults to CI value if available)" default:"1"`
	Chunk       int      `help:"Which chunk to output (1-based, defaults to CI value if available)" default:"1"`
	Count       int      `help:"Number of times to run each test" default:"0"`
	JSON        bool     `help:"Output test results in JSON format" default:"false"`
	Verbose     bool     `short:"v" help:"Verbose output" default:"false"`
	Gotestsum   bool     `help:"Use gotestsum if available" default:"false"`
	Args        []string `arg:"" optional:"" passthrough:"" help:"Packages to test, followed by optional -- and test arguments"`
	WriteTiming string   `help:"Write test timing information to this JSON file" default:""`
	ReadTiming  string   `help:"Read test timing information from files matching this glob pattern" default:""`
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
	logger.Debug().
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

	logger.Debug().
		Strs("packages", packages).
		Strs("testArgs", testArgs).
		Msg("Split arguments")

	// Get all tests
	tests, listErr := testlist.List(packages...)
	if listErr != nil {
		return fmt.Errorf("error listing tests: %w", listErr)
	}

	logger.Debug().
		Int("tests", len(tests)).
		Msg("Found tests")

	// Load timing data if glob pattern provided
	var timings map[string]time.Duration
	if cmd.ReadTiming != "" {
		// Find all matching files
		pattern := cmd.ReadTiming
		if !filepath.IsAbs(pattern) {
			var err error
			pattern, err = filepath.Abs(pattern)
			if err != nil {
				return fmt.Errorf("error getting absolute path: %w", err)
			}
		}

		// Use doublestar to find all matching files
		fs := os.DirFS(filepath.Dir(pattern))
		matches, err := doublestar.Glob(fs, filepath.Base(pattern))
		if err != nil {
			return fmt.Errorf("error finding timing files: %w", err)
		}

		// Convert matches to full paths
		files := make([]string, len(matches))
		for i, match := range matches {
			files[i] = filepath.Join(filepath.Dir(pattern), match)
		}

		if len(files) == 0 {
			logger.Warn().
				Str("pattern", cmd.ReadTiming).
				Msg("No timing files found")
		} else {
			timings, err = timing.LoadFromFiles(files)
			if err != nil {
				return fmt.Errorf("error loading timing data: %w", err)
			}
			logger.Info().
				Str("pattern", cmd.ReadTiming).
				Int("files", len(files)).
				Int("timings", len(timings)).
				Msg("Loaded test timing information")
		}
	}

	// Get tests for this chunk
	var chunkTests []testlist.Test
	var chunkErr error
	if timings != nil {
		chunkTests, chunkErr = testlist.ChunkByTiming(tests, cmd.Chunk-1, cmd.Chunks, timings)
	} else {
		chunkTests, chunkErr = testlist.Chunk(tests, cmd.Chunk-1, cmd.Chunks)
	}
	if chunkErr != nil {
		return fmt.Errorf("error getting chunk: %w", chunkErr)
	}

	if len(chunkTests) == 0 {
		return fmt.Errorf("no tests in chunk %d", cmd.Chunk)
	}

	logger.Info().
		Str("chunk", strconv.Itoa(cmd.Chunk)).
		Int("tests", len(chunkTests)).
		Msg("Found chunk tests")

	// Build go test command args
	goTestArgs := []string{}
	if cmd.JSON || cmd.WriteTiming != "" {
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

	logger.Debug().
		Strs("args", goTestArgs).
		Msg("Running go test")

	// If timing file is requested, we need to capture and parse the output
	if cmd.WriteTiming != "" {
		// Run tests and collect timing data
		tests, err := timing.RunTests(goTestArgs)
		if err != nil {
			return fmt.Errorf("error running tests: %w", err)
		}

		// Write timing data to file if we got any results
		if len(tests) > 0 {
			if err := timing.WriteToFile(tests, cmd.WriteTiming); err != nil {
				return err
			}

			logger.Info().
				Str("file", cmd.WriteTiming).
				Int("tests", len(tests)).
				Msg("Wrote test timing information")
		}

		return nil
	}

	// Original direct execution path
	runner := &testrunner.Runner{
		Args:   goTestArgs,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	return runner.Run()
}
