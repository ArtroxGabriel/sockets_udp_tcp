package calc

import (
	"errors"
	"math"
	"testing"
)

const floatTolerance = 1e-9

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= floatTolerance
}

func TestComputeAddition(t *testing.T) {
	// Arrange
	const op1 = 10.5
	const op2 = 4.5
	const expected = 15.0

	// Act
	result, err := Compute(op1, "+", op2)

	// Assert
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !almostEqual(result, expected) {
		t.Errorf("expected %f, got %f", expected, result)
	}
}

func TestComputeSubtraction(t *testing.T) {
	// Arrange
	const op1 = 20.0
	const op2 = 7.5
	const expected = 12.5

	// Act
	result, err := Compute(op1, "-", op2)

	// Assert
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !almostEqual(result, expected) {
		t.Errorf("expected %f, got %f", expected, result)
	}
}

func TestComputeMultiplication(t *testing.T) {
	// Arrange
	const op1 = 3.5
	const op2 = 2.0
	const expected = 7.0

	// Act
	result, err := Compute(op1, "*", op2)

	// Assert
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !almostEqual(result, expected) {
		t.Errorf("expected %f, got %f", expected, result)
	}
}

func TestComputeDivision(t *testing.T) {
	// Arrange
	const op1 = 15.0
	const op2 = 3.0
	const expected = 5.0

	// Act
	result, err := Compute(op1, "/", op2)

	// Assert
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !almostEqual(result, expected) {
		t.Errorf("expected %f, got %f", expected, result)
	}
}

func TestComputeDivisionByZero(t *testing.T) {
	// Arrange
	const op1 = 8.0
	const op2 = 0.0

	// Act
	_, err := Compute(op1, "/", op2)

	// Assert
	if err == nil {
		t.Fatal("expected error for division by zero, got nil")
	}
	if !errors.Is(err, ErrDivisionByZero) {
		t.Errorf("expected ErrDivisionByZero, got %v", err)
	}
}

func TestComputeInvalidOperator(t *testing.T) {
	// Arrange
	const op1 = 10.0
	const op2 = 5.0
	const invalidOp = "%"

	// Act
	_, err := Compute(op1, invalidOp, op2)

	// Assert
	if err == nil {
		t.Fatal("expected error for invalid operator, got nil")
	}
	if !errors.Is(err, ErrInvalidOperator) {
		t.Errorf("expected ErrInvalidOperator, got %v", err)
	}
}
