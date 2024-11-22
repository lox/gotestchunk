package commands

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/lox/gotestchunk/pkg/ciparallel"
)

type TestCmd struct {
	Chunks        int      `help:"Number of chunks to split tests into (defaults to CI value if available)" default:"1"`
	Chunk         int      `help:"Which chunk to output (1-based, defaults to CI value if available)" default:"1"`
	Count         int      `help:"Number of times to run each test" default:"0"`
	JSON          bool     `help:"Output test results in JSON format" default:"false"`
	Verbose       bool     `short:"v" help:"Verbose output" default:"false"`
	Packages      []string `short:"p" required:"" help:"Packages to list tests from"`
	Gotestsum     bool     `help:"Use gotestsum if available" default:"false"`
	GotestsumArgs []string `name:"gotestsum-args" help:"Arguments to pass to gotestsum (before --)"`
	Args          []string `arg:"" optional:"" help:"Extra arguments to pass to go test (after --)"`
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

func (cmd *TestCmd) Run() error {
	// Default to current directory if no package specified
	pkgPath := "."
	if len(cmd.Packages) > 0 {
		pkgPath = cmd.Packages[0]
	}

	// Get module name first
	modCmd := exec.Command("go", "list", "-m")
	modOutput, err := modCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get module name: %w", err)
	}
	moduleName := strings.TrimSpace(string(modOutput))

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

	// Get unique packages and test pattern
	packages := make(map[string]bool)
	testNames := make([]string, 0, len(chunkTests))
	for _, test := range chunkTests {
		parts := strings.Split(test, ".")
		if len(parts) == 2 {
			packages[parts[0]] = true
			testNames = append(testNames, parts[1])
		}
	}

	// Build go test command args
	goArgs := []string{"test", "-json"}
	if cmd.Verbose {
		goArgs = append(goArgs, "-v")
	}
	if cmd.Count > 0 {
		goArgs = append(goArgs, "-count="+strconv.Itoa(cmd.Count))
	}

	// Add any extra args before the test pattern
	goArgs = append(goArgs, cmd.Args...)

	// Add test pattern
	pattern := fmt.Sprintf("^(%s)$", strings.Join(testNames, "|"))
	goArgs = append(goArgs, "-run="+pattern)

	// Add packages with full import paths
	for pkg := range packages {
		goArgs = append(goArgs, moduleName+"/"+pkg)
	}

	if cmd.Gotestsum {
		// Check if gotestsum is available
		if _, err := exec.LookPath("gotestsum"); err == nil {
			// Build gotestsum command
			args := []string{}
			if len(cmd.GotestsumArgs) > 0 {
				args = append(args, cmd.GotestsumArgs...)
			}
			args = append(args, "--raw-command", "--", "go")
			args = append(args, goArgs...)

			goCmd := exec.Command("gotestsum", args...)
			goCmd.Stdout = os.Stdout
			goCmd.Stderr = os.Stderr
			return goCmd.Run()
		} else {
			fmt.Fprintf(os.Stderr, "Warning: gotestsum not found in PATH, falling back to go test\n")
		}
	}

	// Run go test directly if gotestsum is not requested or not available
	goCmd := exec.Command("go", goArgs...)
	goCmd.Stdout = os.Stdout
	goCmd.Stderr = os.Stderr
	return goCmd.Run()
}
