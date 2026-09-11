package protocol

import (
	"math"
	"testing"
)

const floatTolerance = 1e-9

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= floatTolerance
}

func TestParseRequestValid(t *testing.T) {
	// Arrange
	const raw = "CALC:0:10.5:+:4.5\n"

	// Act
	req, err := ParseRequest(raw)

	// Assert
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if req.Seq != 0 {
		t.Errorf("expected Seq 0, got %d", req.Seq)
	}
	if !almostEqual(req.Op1, 10.5) {
		t.Errorf("expected Op1 10.5, got %f", req.Op1)
	}
	if req.Op != "+" {
		t.Errorf("expected Op '+', got %q", req.Op)
	}
	if !almostEqual(req.Op2, 4.5) {
		t.Errorf("expected Op2 4.5, got %f", req.Op2)
	}
}

func TestParseRequestInvalidPrefix(t *testing.T) {
	// Arrange
	const raw = "SOMAR:0:10:+:5"

	// Act
	_, err := ParseRequest(raw)

	// Assert
	if err == nil {
		t.Fatal("expected error for invalid prefix, got nil")
	}
}

func TestParseRequestInvalidFieldCount(t *testing.T) {
	// Arrange
	const raw = "CALC:0:10:+"

	// Act
	_, err := ParseRequest(raw)

	// Assert
	if err == nil {
		t.Fatal("expected error for invalid field count, got nil")
	}
}

func TestParseRequestInvalidSeq(t *testing.T) {
	// Arrange
	const raw = "CALC:notanumber:10:+:5"

	// Act
	_, err := ParseRequest(raw)

	// Assert
	if err == nil {
		t.Fatal("expected error for invalid sequence number, got nil")
	}
}

func TestParseRequestInvalidOperands(t *testing.T) {
	// Arrange
	const raw = "CALC:0:invalid:+:5"

	// Act
	_, err := ParseRequest(raw)

	// Assert
	if err == nil {
		t.Fatal("expected error for invalid operand, got nil")
	}
}

func TestFormatRequest(t *testing.T) {
	// Arrange
	req := Request{Seq: 2, Op1: 3.5, Op: "*", Op2: 2.0}
	const expected = "CALC:2:3.5:*:2"

	// Act
	formatted := FormatRequest(req)

	// Assert
	if formatted != expected {
		t.Errorf("expected %q, got %q", expected, formatted)
	}
}

func TestFormatResponse(t *testing.T) {
	// Arrange
	const seq = 0
	const result = 15.0
	const expected = "RESULT:0:15.0"

	// Act
	formatted := FormatResponse(seq, result)

	// Assert
	if formatted != expected {
		t.Errorf("expected %q, got %q", expected, formatted)
	}
}

func TestFormatError(t *testing.T) {
	// Arrange
	const seq = 1
	const msg = "divisão por zero"
	const expected = "ERROR:1:divisão por zero"

	// Act
	formatted := FormatError(seq, msg)

	// Assert
	if formatted != expected {
		t.Errorf("expected %q, got %q", expected, formatted)
	}
}

func TestParseResponseSuccess(t *testing.T) {
	// Arrange
	const raw = "RESULT:0:15.0\n"

	// Act
	resp, err := ParseResponse(raw)

	// Assert
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp.IsError {
		t.Error("expected IsError to be false")
	}
	if resp.Seq != 0 {
		t.Errorf("expected Seq 0, got %d", resp.Seq)
	}
	if !almostEqual(resp.Result, 15.0) {
		t.Errorf("expected Result 15.0, got %f", resp.Result)
	}
}

func TestParseResponseError(t *testing.T) {
	// Arrange
	const raw = "ERROR:1:divisão por zero\r\n"

	// Act
	resp, err := ParseResponse(raw)

	// Assert
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !resp.IsError {
		t.Error("expected IsError to be true")
	}
	if resp.Seq != 1 {
		t.Errorf("expected Seq 1, got %d", resp.Seq)
	}
	if resp.ErrorMsg != "divisão por zero" {
		t.Errorf("expected error message 'divisão por zero', got %q", resp.ErrorMsg)
	}
}

func TestParseResponseInvalid(t *testing.T) {
	// Arrange
	const raw = "UNKNOWN:0:data"

	// Act
	_, err := ParseResponse(raw)

	// Assert
	if err == nil {
		t.Fatal("expected error for unknown prefix, got nil")
	}
}
