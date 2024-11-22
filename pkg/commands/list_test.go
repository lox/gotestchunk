package commands

import (
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/lox/gotestchunk/pkg/internal/testutil"
)

func TestListTests(t *testing.T) {
	moduleRoot, err := testutil.GetModuleRoot()
	if err != nil {
		t.Fatalf("failed to get module root: %v", err)
	}

	tests := []struct {
		name     string
		pkgPath  string
		want     []string
		wantErr  bool
		contains []string // partial list of tests that must be present
	}{
		{
			name:    "list tests in example package",
			pkgPath: filepath.Join(moduleRoot, "pkg/example"),
			contains: []string{
				"pkg/example.TestSimple",
				"pkg/example.TestParallel",
				"pkg/example.TestTableDriven",
			},
		},
		{
			name:    "list tests in sub package",
			pkgPath: filepath.Join(moduleRoot, "pkg/example/sub"),
			contains: []string{
				"pkg/example/sub.TestMath",
				"pkg/example/sub.TestDivideErrors",
			},
		},
		{
			name:    "list tests in all packages",
			pkgPath: filepath.Join(moduleRoot, "pkg/example/..."),
			contains: []string{
				"pkg/example.TestSimple",
				"pkg/example/sub.TestMath",
			},
		},
		{
			name:    "invalid package path",
			pkgPath: "./does-not-exist",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := listTests(tt.pkgPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("listTests() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Sort both slices for comparison
			sort.Strings(got)

			// Check that all required tests are present
			for _, want := range tt.contains {
				found := false
				for _, g := range got {
					if strings.HasSuffix(g, want) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("listTests() missing test %q in output:\n%s", want, strings.Join(got, "\n"))
				}
			}
		})
	}
}
