package commands

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/lox/gotestchunk/pkg/internal/testutil"
)

func TestChunkCmd_Validate(t *testing.T) {
	tests := []struct {
		name    string
		chunks  int
		chunk   int
		wantErr bool
	}{
		{
			name:    "valid single chunk",
			chunks:  1,
			chunk:   1,
			wantErr: false,
		},
		{
			name:    "valid multi chunk",
			chunks:  4,
			chunk:   2,
			wantErr: false,
		},
		{
			name:    "invalid chunks zero",
			chunks:  0,
			chunk:   1,
			wantErr: true,
		},
		{
			name:    "invalid chunk zero",
			chunks:  2,
			chunk:   0,
			wantErr: true,
		},
		{
			name:    "invalid chunk too large",
			chunks:  2,
			chunk:   3,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &ChunkCmd{
				Chunks: tt.chunks,
				Chunk:  tt.chunk,
			}
			err := cmd.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ChunkCmd.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Test ChunkPackagesCmd validation too
			cmdPkg := &ChunkPackagesCmd{
				Chunks: tt.chunks,
				Chunk:  tt.chunk,
			}
			err = cmdPkg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ChunkPackagesCmd.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMakeRunPattern(t *testing.T) {
	tests := []struct {
		name     string
		tests    []string
		want     string
		contains []string // strings that must be in the pattern
	}{
		{
			name: "simple tests",
			tests: []string{
				"pkg/example.TestSimple",
				"pkg/example.TestParallel",
			},
			contains: []string{
				"TestSimple",
				"TestParallel",
				"^(",
				")$",
			},
		},
		{
			name: "tests from different packages",
			tests: []string{
				"pkg/example.TestSimple",
				"pkg/example/sub.TestMath",
			},
			contains: []string{
				"TestSimple",
				"TestMath",
				"|",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := makeRunPattern(tt.tests)

			// Check required substrings
			for _, want := range tt.contains {
				if !strings.Contains(pattern, want) {
					t.Errorf("makeRunPattern() = %v, should contain %v", pattern, want)
				}
			}
		})
	}
}

func TestChunkCommands(t *testing.T) {
	moduleRoot, err := testutil.GetModuleRoot()
	if err != nil {
		t.Fatalf("failed to get module root: %v", err)
	}

	// First ensure we have enough tests to chunk
	tests, err := listTests(moduleRoot + "/pkg/example/...")
	if err != nil {
		t.Fatalf("failed to list tests: %v", err)
	}
	if len(tests) < 3 {
		t.Fatal("not enough tests to chunk meaningfully")
	}

	// Test different chunking scenarios
	scenarios := []struct {
		name      string
		chunks    int
		chunk     int
		pkgPath   string
		wantTests int // approximate number of tests we expect
	}{
		{
			name:      "single chunk",
			chunks:    1,
			chunk:     1,
			pkgPath:   moduleRoot + "/pkg/example/...",
			wantTests: len(tests),
		},
		{
			name:      "first of two chunks",
			chunks:    2,
			chunk:     1,
			pkgPath:   moduleRoot + "/pkg/example/...",
			wantTests: len(tests) / 2,
		},
		{
			name:      "second of two chunks",
			chunks:    2,
			chunk:     2,
			pkgPath:   moduleRoot + "/pkg/example/...",
			wantTests: len(tests) / 2,
		},
	}

	for _, tt := range scenarios {
		t.Run(tt.name+" pattern", func(t *testing.T) {
			cmd := &ChunkCmd{
				Chunks:   tt.chunks,
				Chunk:    tt.chunk,
				Packages: []string{tt.pkgPath},
			}

			output := captureOutput(t, func() error {
				return cmd.Run()
			})

			// Verify the pattern format
			if !strings.HasPrefix(output, "^(") || !strings.HasSuffix(output, ")$") {
				t.Errorf("ChunkCmd.Run() output = %v, want pattern matching ^(...)$", output)
			}

			// Count the number of tests in the pattern
			testCount := strings.Count(output, "|") + 1
			if testCount < tt.wantTests/2 || testCount > tt.wantTests*2 {
				t.Errorf("ChunkCmd.Run() got %d tests, want approximately %d", testCount, tt.wantTests)
			}
		})

		t.Run(tt.name+" packages", func(t *testing.T) {
			cmd := &ChunkPackagesCmd{
				Chunks:   tt.chunks,
				Chunk:    tt.chunk,
				Packages: []string{tt.pkgPath},
			}

			output := captureOutput(t, func() error {
				return cmd.Run()
			})

			// Verify package paths
			if !strings.Contains(output, "pkg/example") {
				t.Errorf("ChunkPackagesCmd.Run() output = %v, should contain package paths", output)
			}
		})
	}
}

// Helper function to capture stdout
func captureOutput(t *testing.T, fn func() error) string {
	stdout := &strings.Builder{}
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := fn()
	if err != nil {
		t.Fatalf("command error = %v", err)
	}

	w.Close()
	os.Stdout = oldStdout
	io.Copy(stdout, r)
	return stdout.String()
}
