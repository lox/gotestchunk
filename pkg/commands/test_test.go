package commands

import (
	"os"
	"testing"

	"github.com/lox/gotestchunk/pkg/testlist"
	"github.com/rs/zerolog"
)

func TestTestCmd_Run(t *testing.T) {
	tests := []struct {
		name      string
		cmd       *TestCmd
		wantError bool
	}{
		{
			name: "single chunk",
			cmd: &TestCmd{
				Chunks: 1,
				Chunk:  1,
				Args:   []string{"./pkg/example/..."},
			},
		},
		{
			name: "first of two chunks",
			cmd: &TestCmd{
				Chunks: 2,
				Chunk:  1,
				Args:   []string{"./pkg/example/..."},
			},
		},
		{
			name: "second of two chunks with verbose",
			cmd: &TestCmd{
				Chunks:  2,
				Chunk:   2,
				Verbose: true,
				Args:    []string{"./pkg/example/..."},
			},
		},
		{
			name: "with extra args",
			cmd: &TestCmd{
				Chunks: 1,
				Chunk:  1,
				Args:   []string{"./pkg/example/...", "--", "-v", "-count=1"},
			},
		},
		{
			name: "invalid chunk index",
			cmd: &TestCmd{
				Chunks: 2,
				Chunk:  3,
				Args:   []string{"./pkg/example/..."},
			},
			wantError: true,
		},
		{
			name: "invalid package",
			cmd: &TestCmd{
				Chunks: 1,
				Chunk:  1,
				Args:   []string{"./does-not-exist"},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		testlist.TestRunWithModuleRoot(t, tt.name, func(t *testing.T) {
			logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).Level(zerolog.DebugLevel)
			err := tt.cmd.Run(&logger)
			if (err != nil) != tt.wantError {
				t.Errorf("TestCmd.Run() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestTestCmd_Validate(t *testing.T) {
	tests := []struct {
		name      string
		cmd       *TestCmd
		wantError bool
	}{
		{
			name: "valid chunks",
			cmd: &TestCmd{
				Chunks: 2,
				Chunk:  1,
			},
		},
		{
			name: "invalid chunks",
			cmd: &TestCmd{
				Chunks: 0,
				Chunk:  1,
			},
			wantError: true,
		},
		{
			name: "invalid chunk",
			cmd: &TestCmd{
				Chunks: 2,
				Chunk:  3,
			},
			wantError: true,
		},
		{
			name: "negative chunk",
			cmd: &TestCmd{
				Chunks: 2,
				Chunk:  -1,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("TestCmd.Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
