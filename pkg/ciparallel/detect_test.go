package ciparallel

import (
	"os"
	"testing"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name       string
		env        map[string]string
		wantChunk  int
		wantChunks int
		wantNotNil bool
	}{
		{
			name: "gitlab ci",
			env: map[string]string{
				"CI_NODE_INDEX": "0",
				"CI_NODE_TOTAL": "4",
			},
			wantChunk:  1,
			wantChunks: 4,
			wantNotNil: true,
		},
		{
			name: "circle ci",
			env: map[string]string{
				"CIRCLE_NODE_INDEX": "1",
				"CIRCLE_NODE_TOTAL": "3",
			},
			wantChunk:  1,
			wantChunks: 3,
			wantNotNil: true,
		},
		{
			name: "buildkite",
			env: map[string]string{
				"BUILDKITE_PARALLEL_JOB":       "2",
				"BUILDKITE_PARALLEL_JOB_COUNT": "5",
			},
			wantChunk:  2,
			wantChunks: 5,
			wantNotNil: true,
		},
		{
			name:       "no ci env",
			env:        map[string]string{},
			wantNotNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set test environment
			for k, v := range tt.env {
				os.Setenv(k, v)
			}

			got := Detect()
			if (got != nil) != tt.wantNotNil {
				t.Errorf("Detect() = %v, want not nil: %v", got, tt.wantNotNil)
				return
			}

			if got != nil {
				if got.Index != tt.wantChunk {
					t.Errorf("Detect() chunk = %v, want %v", got.Index, tt.wantChunk)
				}
				if got.Total != tt.wantChunks {
					t.Errorf("Detect() chunks = %v, want %v", got.Total, tt.wantChunks)
				}
			}
		})
	}
}
