package sub

import (
	"fmt"
	"testing"
	"time"
)

// Test group using t.Run
func TestMath(t *testing.T) {
	t.Run("Multiply", func(t *testing.T) {
		t.Parallel()
		if got := Multiply(2, 3); got != 6 {
			t.Errorf("Multiply(2, 3) = %v, want 6", got)
		}
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("Divide", func(t *testing.T) {
		t.Parallel()
		got, err := Divide(6, 2)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if got != 3 {
			t.Errorf("Divide(6, 2) = %v, want 3", got)
		}
		time.Sleep(100 * time.Millisecond)
	})
}

// Example test
func ExampleMultiply() {
	result := Multiply(4, 5)
	fmt.Println(result)
	// Output: 20
}

// Benchmark test
func BenchmarkMultiply(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Multiply(2, 3)
	}
}

// Test with error cases
func TestDivideErrors(t *testing.T) {
	cases := []struct {
		name     string
		a, b     int
		want     int
		wantErr  bool
		errType  error
		skipSlow bool
	}{
		{
			name:    "divide by zero",
			a:       10,
			b:       0,
			wantErr: true,
			errType: ErrDivideByZero,
		},
		{
			name:     "slow division",
			a:        100,
			b:        2,
			want:     50,
			skipSlow: true,
		},
	}

	for _, tc := range cases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			if tc.skipSlow && testing.Short() {
				t.Skip("skipping slow test")
			}
			t.Parallel()

			got, err := Divide(tc.a, tc.b)
			if (err != nil) != tc.wantErr {
				t.Errorf("Divide(%d, %d) error = %v, wantErr %v", tc.a, tc.b, err, tc.wantErr)
				return
			}
			if tc.wantErr {
				if err != tc.errType {
					t.Errorf("Divide(%d, %d) error = %v, want %v", tc.a, tc.b, err, tc.errType)
				}
				return
			}
			if got != tc.want {
				t.Errorf("Divide(%d, %d) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
			time.Sleep(150 * time.Millisecond)
		})
	}
}
