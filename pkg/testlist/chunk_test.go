package testlist

import (
	"reflect"
	"testing"
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
