package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/lox/gotestchunk/pkg/commands"
)

var cli struct {
	List          commands.ListCmd          `cmd:"" help:"List tests in packages"`
	Chunk         commands.ChunkCmd         `cmd:"" help:"Output test pattern for a chunk"`
	ChunkPackages commands.ChunkPackagesCmd `cmd:"" help:"Output package paths for a chunk"`
	Test          commands.TestCmd          `cmd:"" help:"Run tests for a specific chunk"`
}

func main() {
	ctx := kong.Parse(&cli,
		kong.Name("gotestchunk"),
		kong.Description("A tool for listing and chunking Go tests"),
		kong.UsageOnError(),
	)
	err := ctx.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
