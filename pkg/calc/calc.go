package calc

import (
	"errors"
	"fmt"
)

// Standard error definitions for calculator operations.
var (
	ErrDivisionByZero  = errors.New("divisão por zero")
	ErrInvalidOperator = errors.New("operação inválida")
)

// Compute evaluates a binary arithmetic operation on two floating-point numbers.
//
// Usage:
//
//	res, err := calc.Compute(10, "+", 5)
func Compute(op1 float64, op string, op2 float64) (float64, error) {
	switch op {
	case "+":
		return op1 + op2, nil
	case "-":
		return op1 - op2, nil
	case "*":
		return op1 * op2, nil
	case "/":
		if op2 == 0 {
			return 0, fmt.Errorf("%w: cannot divide %v by zero", ErrDivisionByZero, op1)
		}
		return op1 / op2, nil
	default:
		return 0, fmt.Errorf("%w: unsupported operator %q", ErrInvalidOperator, op)
	}
}
