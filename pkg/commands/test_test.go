package commands

import (
	"testing"

	"github.com/lox/gotestchunk/pkg/internal/testutil"
)

func TestTestCmd_Run(t *testing.T) {
	moduleRoot, err := testutil.GetModuleRoot()
	if err != nil {
		t.Fatalf("failed to get module root: %v", err)
	}

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
			pkgPath: moduleRoot + "/pkg/example/...",
		},
		{
			name:    "first of two chunks",
			chunks:  2,
			chunk:   1,
			pkgPath: moduleRoot + "/pkg/example/...",
			json:    true,
		},
		{
			name:    "with extra args",
			chunks:  2,
			chunk:   1,
			pkgPath: moduleRoot + "/pkg/example/...",
			args:    []string{"-count=1", "-timeout=10s"},
		},
		{
			name:    "second of two chunks with verbose",
			chunks:  2,
			chunk:   2,
			pkgPath: moduleRoot + "/pkg/example/...",
			verbose: true,
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
		t.Run(tt.name, func(t *testing.T) {
			cmd := &TestCmd{
				Chunks:   tt.chunks,
				Chunk:    tt.chunk,
				Packages: []string{tt.pkgPath},
				JSON:     tt.json,
				Verbose:  tt.verbose,
				Args:     tt.args,
			}

			err := cmd.Run()
			if (err != nil) != tt.wantError {
				t.Errorf("TestCmd.Run() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
