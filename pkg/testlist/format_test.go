package testlist

import (
	"testing"
)

func TestFormat(t *testing.T) {
	tests := []Test{
		{Package: "pkg/example", Name: "TestOne"},
		{Package: "pkg/example", Name: "TestTwo"},
		{Package: "pkg/other", Name: "TestThree"},
	}

	testCases := []struct {
		name     string
		format   string
		tests    []Test
		expected string
		wantErr  bool
	}{
		{
			name:     "listTests format",
			format:   "listTests",
			tests:    tests,
			expected: "TestOne\nTestTwo\nTestThree",
		},
		{
			name:     "listPackages format",
			format:   "listPackages",
			tests:    tests,
			expected: "./pkg/example\n./pkg/other",
		},
		{
			name:     "runPattern format",
			format:   "runPattern",
			tests:    tests,
			expected: "^(TestOne|TestTwo|TestThree)$",
		},
		{
			name:     "empty tests with runPattern",
			format:   "runPattern",
			tests:    []Test{},
			expected: "",
		},
		{
			name:    "unknown format",
			format:  "invalid",
			tests:   tests,
			wantErr: true,
		},
		{
			name:     "single test",
			format:   "runPattern",
			tests:    []Test{{Package: "pkg/example", Name: "TestOne"}},
			expected: "^(TestOne)$",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := Format(tc.tests, tc.format)

			if tc.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if output != tc.expected {
				t.Errorf("got %q, want %q", output, tc.expected)
			}
		})
	}
}
