package sub

// Multiply multiplies two numbers
func Multiply(a, b int) int {
	return a * b
}

// ErrDivideByZero is returned when attempting to divide by zero
var ErrDivideByZero = &DivideByZeroError{}

// DivideByZeroError represents a division by zero error
type DivideByZeroError struct{}

func (e *DivideByZeroError) Error() string {
	return "division by zero"
}

// Divide divides two numbers
func Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, ErrDivideByZero
	}
	return a / b, nil
}
