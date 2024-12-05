package commands

import (
	"fmt"

	"github.com/lox/gotestchunk/pkg/ciparallel"
	"github.com/lox/gotestchunk/pkg/testlist"
	"github.com/rs/zerolog"
)

type ListCmd struct {
	Package string `arg:"" optional:"" help:"Package to list tests from" default:"."`
	Chunks  int    `help:"Number of chunks to split tests into (defaults to CI value if available)" default:"1"`
	Chunk   int    `help:"Which chunk to output (1-based, defaults to CI value if available)" default:"1"`
	Format  string `help:"Output format (listTests|listPackages|runPattern)" default:"listTests" enum:"listTests,listPackages,runPattern"`
}

func (cmd *ListCmd) Validate() error {
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

func (cmd *ListCmd) Run(logger *zerolog.Logger) error {
	tests, err := testlist.List(cmd.Package)
	if err != nil {
		return fmt.Errorf("error listing tests: %w", err)
	}

	logger.Debug().
		Int("tests", len(tests)).
		Msg("Found tests")

	// If chunking is enabled, get the subset of tests for this chunk
	if cmd.Chunks > 1 {
		logger.Debug().
			Int("chunks", cmd.Chunks).
			Int("chunk", cmd.Chunk).
			Msg("Chunking tests")

		// Use 0-based index for testlist.Chunk
		chunkTests, err := testlist.Chunk(tests, cmd.Chunk-1, cmd.Chunks)
		if err != nil {
			return fmt.Errorf("error chunking tests: %w", err)
		}
		tests = chunkTests
	}

	output, err := testlist.Format(tests, cmd.Format)
	if err != nil {
		return err
	}

	if output != "" {
		fmt.Println(output)
	}

	return nil
}
