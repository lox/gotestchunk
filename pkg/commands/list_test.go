package commands

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/lox/gotestchunk/pkg/testlist"
	"github.com/rs/zerolog"
)

func TestListCmd_Run(t *testing.T) {
	tests := []struct {
		name    string
		cmd     ListCmd
		want    string
		wantErr bool
	}{
		{
			name: "list all tests",
			cmd: ListCmd{
				Package: "./pkg/example/...",
				Format:  "listTests",
			},
			want: `TestSimple
TestParallel
TestTableDriven
TestWithSetup
TestMath
TestDivideErrors`,
		},
		{
			name: "list packages",
			cmd: ListCmd{
				Package: "./pkg/example/...",
				Format:  "listPackages",
			},
			want: `./pkg/example
./pkg/example/sub`,
		},
		{
			name: "run pattern",
			cmd: ListCmd{
				Package: "./pkg/example",
				Format:  "runPattern",
			},
			want: "^(TestSimple|TestParallel|TestTableDriven|TestWithSetup)$",
		},
		{
			name: "chunk tests",
			cmd: ListCmd{
				Package: "./pkg/example/...",
				Format:  "listTests",
				Chunks:  2,
				Chunk:   1,
			},
			want: `TestSimple
TestParallel
TestTableDriven`,
		},
		{
			name: "invalid package",
			cmd: ListCmd{
				Package: "./does-not-exist",
				Format:  "listTests",
			},
			wantErr: true,
		},
		{
			name: "invalid chunk",
			cmd: ListCmd{
				Package: "./pkg/example/...",
				Format:  "listTests",
				Chunks:  2,
				Chunk:   3,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		testlist.TestRunWithModuleRoot(t, tt.name, func(t *testing.T) {
			logger := zerolog.New(zerolog.NewTestWriter(t))

			output, err := captureOutput(func() error {
				return tt.cmd.Run(&logger)
			})

			if tt.wantErr {
				if err == nil {
					t.Error("ListCmd.Run() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ListCmd.Run() unexpected error: %v", err)
			}

			// Normalize output and expected output
			gotLines := normalizeOutput(output)
			wantLines := normalizeOutput(tt.want)

			// Compare output
			if !reflect.DeepEqual(gotLines, wantLines) {
				t.Errorf("ListCmd.Run() output mismatch:\ngot:\n%s\nwant:\n%s",
					strings.Join(gotLines, "\n"),
					strings.Join(wantLines, "\n"))
			}
		})
	}
}

// normalizeOutput splits output into lines, trims spaces, and sorts them
func normalizeOutput(s string) []string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	sort.Strings(lines)
	return lines
}

func TestListCmd_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cmd     ListCmd
		wantErr bool
	}{
		{
			name: "valid defaults",
			cmd: ListCmd{
				Chunks: 1,
				Chunk:  1,
			},
			wantErr: false,
		},
		{
			name: "valid chunks",
			cmd: ListCmd{
				Chunks: 3,
				Chunk:  2,
			},
			wantErr: false,
		},
		{
			name: "invalid chunks",
			cmd: ListCmd{
				Chunks: 0,
				Chunk:  1,
			},
			wantErr: true,
		},
		{
			name: "invalid chunk number",
			cmd: ListCmd{
				Chunks: 2,
				Chunk:  3,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListCmd.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
