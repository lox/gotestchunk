package testlist

import (
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestChunks(t *testing.T) {
	tests := []struct {
		name      string
		tests     []Test
		total     int
		wantSizes []int // expected size of each chunk
	}{
		{
			name:      "empty tests",
			tests:     []Test{},
			total:     3,
			wantSizes: []int{0, 0, 0},
		},
		{
			name: "single chunk",
			tests: []Test{
				{Package: "./pkg/example", Name: "TestSimple"},
				{Package: "./pkg/example", Name: "TestParallel"},
			},
			total:     1,
			wantSizes: []int{2},
		},
		{
			name: "even distribution",
			tests: []Test{
				{Package: "./pkg/example", Name: "TestSimple"},
				{Package: "./pkg/example", Name: "TestParallel"},
				{Package: "./pkg/example/sub", Name: "TestMath"},
				{Package: "./pkg/example/sub", Name: "TestDivideErrors"},
			},
			total:     2,
			wantSizes: []int{2, 2},
		},
		{
			name: "uneven distribution",
			tests: []Test{
				{Package: "./pkg/example", Name: "TestSimple"},
				{Package: "./pkg/example", Name: "TestParallel"},
				{Package: "./pkg/example/sub", Name: "TestMath"},
			},
			total:     2,
			wantSizes: []int{2, 1},
		},
		{
			name: "more chunks than tests",
			tests: []Test{
				{Package: "./pkg/example", Name: "TestSimple"},
			},
			total:     3,
			wantSizes: []int{1, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Chunks(tt.tests, tt.total)

			if len(got) != len(tt.wantSizes) {
				t.Errorf("Chunks() got %d chunks, want %d", len(got), len(tt.wantSizes))
			}

			for i, chunk := range got {
				if len(chunk) != tt.wantSizes[i] {
					t.Errorf("Chunks() chunk[%d] size = %d, want %d", i, len(chunk), tt.wantSizes[i])
				}
			}

			// Verify all tests are present and in order
			if len(tt.tests) > 0 {
				var allTests []Test
				for _, chunk := range got {
					allTests = append(allTests, chunk...)
				}
				if !reflect.DeepEqual(allTests, tt.tests) {
					t.Errorf("Chunks() combined chunks = %v, want %v", allTests, tt.tests)
				}
			}
		})
	}
}

func TestChunk(t *testing.T) {
	tests := []struct {
		name    string
		tests   []Test
		index   int
		total   int
		want    []Test
		wantErr bool
	}{
		{
			name: "get first chunk",
			tests: []Test{
				{Package: "./pkg/example", Name: "TestParallel"},
				{Package: "./pkg/example", Name: "TestSimple"},
				{Package: "./pkg/example/sub", Name: "TestDivideErrors"},
				{Package: "./pkg/example/sub", Name: "TestMath"},
			},
			index: 0,
			total: 2,
			want: []Test{
				{Package: "./pkg/example", Name: "TestParallel"},
				{Package: "./pkg/example", Name: "TestSimple"},
			},
		},
		{
			name: "get second chunk",
			tests: []Test{
				{Package: "./pkg/example", Name: "TestParallel"},
				{Package: "./pkg/example", Name: "TestSimple"},
				{Package: "./pkg/example/sub", Name: "TestDivideErrors"},
				{Package: "./pkg/example/sub", Name: "TestMath"},
			},
			index: 1,
			total: 2,
			want: []Test{
				{Package: "./pkg/example/sub", Name: "TestDivideErrors"},
				{Package: "./pkg/example/sub", Name: "TestMath"},
			},
		},
		{
			name: "get empty chunk",
			tests: []Test{
				{Package: "./pkg/example", Name: "TestSimple"},
			},
			index: 1,
			total: 2,
			want:  []Test{},
		},
		{
			name:    "invalid index",
			tests:   []Test{{Package: "./pkg/example", Name: "TestSimple"}},
			index:   2,
			total:   2,
			wantErr: true,
		},
		{
			name:    "negative index",
			tests:   []Test{{Package: "./pkg/example", Name: "TestSimple"}},
			index:   -1,
			total:   1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Chunk(tt.tests, tt.index, tt.total)
			if (err != nil) != tt.wantErr {
				t.Errorf("Chunk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Chunk() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChunkByTiming(t *testing.T) {
	tests := []struct {
		name     string
		tests    []Test
		timings  map[string]time.Duration
		index    int
		total    int
		want     []Test
		wantErr  bool
		errMatch string
	}{
		{
			name: "balanced by timing",
			tests: []Test{
				{Package: "pkg/a", Name: "Test1"},
				{Package: "pkg/a", Name: "Test2"},
				{Package: "pkg/b", Name: "Test3"},
				{Package: "pkg/b", Name: "Test4"},
			},
			timings: map[string]time.Duration{
				"pkg/a.Test1": 100 * time.Millisecond,
				"pkg/a.Test2": 200 * time.Millisecond,
				"pkg/b.Test3": 300 * time.Millisecond,
				"pkg/b.Test4": 400 * time.Millisecond,
			},
			index: 0,
			total: 2,
			want: []Test{
				{Package: "pkg/b", Name: "Test4"}, // 400ms
				{Package: "pkg/a", Name: "Test1"}, // 100ms
				// Total: 500ms (other chunk gets Test2+Test3 = 500ms)
			},
		},
		{
			name: "missing timing uses default",
			tests: []Test{
				{Package: "pkg/a", Name: "Test1"},
				{Package: "pkg/a", Name: "Test2"},
			},
			timings: map[string]time.Duration{
				"pkg/a.Test1": 100 * time.Millisecond,
				// Test2 missing, should use default
			},
			index: 0,
			total: 2,
			want: []Test{
				{Package: "pkg/a", Name: "Test2"}, // 1s (default)
			},
		},
		{
			name: "single chunk gets all tests",
			tests: []Test{
				{Package: "pkg/a", Name: "Test1"},
				{Package: "pkg/a", Name: "Test2"},
			},
			timings: map[string]time.Duration{
				"pkg/a.Test1": 100 * time.Millisecond,
				"pkg/a.Test2": 200 * time.Millisecond,
			},
			index: 0,
			total: 1,
			want: []Test{
				{Package: "pkg/a", Name: "Test1"},
				{Package: "pkg/a", Name: "Test2"},
			},
		},
		{
			name:  "empty tests list",
			tests: []Test{},
			timings: map[string]time.Duration{
				"pkg/a.Test1": 100 * time.Millisecond,
			},
			index: 0,
			total: 2,
			want:  []Test{},
		},
		{
			name: "index out of bounds",
			tests: []Test{
				{Package: "pkg/a", Name: "Test1"},
			},
			timings: map[string]time.Duration{
				"pkg/a.Test1": 100 * time.Millisecond,
			},
			index:    2,
			total:    2,
			wantErr:  true,
			errMatch: "chunk index 2 out of bounds",
		},
		{
			name: "negative index",
			tests: []Test{
				{Package: "pkg/a", Name: "Test1"},
			},
			timings: map[string]time.Duration{
				"pkg/a.Test1": 100 * time.Millisecond,
			},
			index:    -1,
			total:    2,
			wantErr:  true,
			errMatch: "chunk index -1 out of bounds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ChunkByTiming(tt.tests, tt.index, tt.total, tt.timings)
			if tt.wantErr {
				if err == nil {
					t.Error("ChunkByTiming() expected error")
				} else if tt.errMatch != "" && !strings.Contains(err.Error(), tt.errMatch) {
					t.Errorf("ChunkByTiming() error = %v, want match %v", err, tt.errMatch)
				}
				return
			}
			if err != nil {
				t.Fatalf("ChunkByTiming() unexpected error: %v", err)
			}

			// For empty slices, just check length
			if len(tt.want) == 0 {
				if len(got) != 0 {
					t.Errorf("ChunkByTiming() got %d tests, want empty slice", len(got))
				}
				return
			}

			// Sort both slices for consistent comparison
			sortTests(got)
			sortTests(tt.want)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ChunkByTiming() = %v, want %v", got, tt.want)
			}
		})
	}
}

// sortTests sorts tests by package and name for consistent comparison
func sortTests(tests []Test) {
	sort.Slice(tests, func(i, j int) bool {
		if tests[i].Package != tests[j].Package {
			return tests[i].Package < tests[j].Package
		}
		return tests[i].Name < tests[j].Name
	})
}
