package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/lox/gotestchunk/pkg/commands"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var cli struct {
	Debug bool `short:"d" help:"Enable debug logging"`

	List commands.ListCmd `cmd:"" help:"List tests in packages"`
	Test commands.TestCmd `cmd:"" help:"Run tests for a specific chunk" default:"withargs"`
}

func main() {
	logger := log.Output(zerolog.ConsoleWriter{
		Out: os.Stderr,
	})

	if cli.Debug {
		logger = logger.Level(zerolog.DebugLevel)
	} else {
		logger = logger.Level(zerolog.InfoLevel)
	}

	ctx := kong.Parse(&cli,
		kong.Name("gotestchunk"),
		kong.Description("A tool for listing and chunking Go tests"),
		kong.UsageOnError(),
	)
	err := ctx.Run(&logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
