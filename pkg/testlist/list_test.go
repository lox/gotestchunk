package testlist

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestList(t *testing.T) {
	moduleRoot, err := GetModuleRoot()
	if err != nil {
		t.Fatalf("failed to get module root: %v", err)
	}

	tests := []struct {
		name     string
		pkgPath  string
		want     []Test
		wantErr  bool
		contains []Test // partial list of tests that must be present
	}{
		{
			name:    "example package",
			pkgPath: moduleRoot + "/pkg/example",
			contains: []Test{
				{Package: "pkg/example", Name: "TestSimple"},
				{Package: "pkg/example", Name: "TestParallel"},
				{Package: "pkg/example", Name: "TestTableDriven"},
				{Package: "pkg/example", Name: "TestWithSetup"},
			},
		},
		{
			name:    "sub package",
			pkgPath: moduleRoot + "/pkg/example/sub",
			contains: []Test{
				{Package: "pkg/example/sub", Name: "TestMath"},
				{Package: "pkg/example/sub", Name: "TestDivideErrors"},
			},
		},
		{
			name:    "all packages",
			pkgPath: moduleRoot + "/pkg/example/...",
			contains: []Test{
				{Package: "pkg/example", Name: "TestSimple"},
				{Package: "pkg/example", Name: "TestParallel"},
				{Package: "pkg/example/sub", Name: "TestMath"},
			},
		},
		{
			name:    "invalid package",
			pkgPath: "./does-not-exist",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := List(tt.pkgPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Sort both slices for comparison
			sort.Slice(got, func(i, j int) bool {
				return got[i].String() < got[j].String()
			})

			// Check that all required tests are present
			for _, want := range tt.contains {
				found := false
				for _, g := range got {
					if g.Package == want.Package && g.Name == want.Name {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("List() missing test %v in output:\n%s", want, formatTests(got))
				}
			}
		})
	}
}

func TestTest_String(t *testing.T) {
	tests := []struct {
		name string
		test Test
		want string
	}{
		{
			name: "simple test",
			test: Test{Package: "pkg/example", Name: "TestSimple"},
			want: "pkg/example.TestSimple",
		},
		{
			name: "test in subpackage",
			test: Test{Package: "pkg/example/sub", Name: "TestMath"},
			want: "pkg/example/sub.TestMath",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.test.String(); got != tt.want {
				t.Errorf("Test.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to format tests for error messages
func formatTests(tests []Test) string {
	var b strings.Builder
	for _, t := range tests {
		b.WriteString(t.String())
		b.WriteString("\n")
	}
	return b.String()
}

func TestPackages(t *testing.T) {
	tests := []struct {
		name  string
		tests []Test
		want  []string
	}{
		{
			name: "single package",
			tests: []Test{
				{Package: "pkg/example", Name: "TestSimple"},
				{Package: "pkg/example", Name: "TestParallel"},
			},
			want: []string{"pkg/example"},
		},
		{
			name: "multiple packages",
			tests: []Test{
				{Package: "pkg/example", Name: "TestSimple"},
				{Package: "pkg/example/sub", Name: "TestMath"},
				{Package: "pkg/example", Name: "TestParallel"},
				{Package: "pkg/example/sub", Name: "TestDivide"},
			},
			want: []string{"pkg/example", "pkg/example/sub"},
		},
		{
			name:  "empty test list",
			tests: []Test{},
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Packages(tt.tests)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Packages() = %v, want %v", got, tt.want)
			}
		})
	}
}
