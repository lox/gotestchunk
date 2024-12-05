package commands

import (
	"io"
	"testing"

	"github.com/lox/gotestchunk/pkg/testlist"
	"github.com/rs/zerolog"
)

func TestTestCmd_Run(t *testing.T) {
	tests := []struct {
		name      string
		chunks    int
		chunk     int
		pkgPath   string
		json      bool
		verbose   bool
		args      []string
		wantError bool
	}{
		{
			name:    "single chunk",
			chunks:  1,
			chunk:   1,
			pkgPath: "./pkg/example/...",
			json:    false,
		},
		{
			name:    "first of two chunks",
			chunks:  2,
			chunk:   1,
			pkgPath: "./pkg/example/...",
			json:    false,
		},
		{
			name:    "with extra args",
			chunks:  2,
			chunk:   1,
			pkgPath: "./pkg/example/...",
			args:    []string{"--", "-count=1", "-timeout=10s"},
			json:    false,
		},
		{
			name:    "second of two chunks with verbose",
			chunks:  2,
			chunk:   2,
			pkgPath: "./pkg/example/...",
			verbose: true,
			json:    false,
		},
		{
			name:      "invalid package",
			chunks:    1,
			chunk:     1,
			pkgPath:   "./does-not-exist",
			wantError: true,
		},
	}

	for _, tt := range tests {
		testlist.TestRunWithModuleRoot(t, tt.name, func(t *testing.T) {
			cmd := &TestCmd{
				Chunks:  tt.chunks,
				Chunk:   tt.chunk,
				Args:    append([]string{tt.pkgPath}, tt.args...),
				JSON:    tt.json,
				Verbose: tt.verbose,
			}

			logger := zerolog.New(io.Discard)
			err := cmd.Run(&logger)
			if (err != nil) != tt.wantError {
				t.Errorf("TestCmd.Run() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
