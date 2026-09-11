package protocol

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	PrefixCalc             = "CALC"
	PrefixResult           = "RESULT"
	PrefixError            = "ERROR"
	FieldDelimiter         = ":"
	ExpectedRequestFields  = 5
	ExpectedResponseFields = 3
)

var (
	ErrInvalidMessageFormat = errors.New("invalid message format")
	ErrInvalidSequence      = errors.New("invalid sequence number")
	ErrInvalidOperand       = errors.New("invalid operand")
)

// Request represents a parsed mathematical calculation request.
type Request struct {
	Seq int
	Op1 float64
	Op  string
	Op2 float64
}

// Response represents a parsed server response (either success or error).
type Response struct {
	Seq      int
	Result   float64
	ErrorMsg string
	IsError  bool
}

// ParseRequest parses a raw message into a structured Request.
//
// Usage:
//
//	req, err := protocol.ParseRequest("CALC:0:10:+:5")
func ParseRequest(raw string) (Request, error) {
	clean := strings.TrimSpace(raw)
	parts := strings.Split(clean, FieldDelimiter)
	if len(parts) != ExpectedRequestFields {
		return Request{}, fmt.Errorf("%w: expected %d fields, got %d", ErrInvalidMessageFormat, ExpectedRequestFields, len(parts))
	}
	if parts[0] != PrefixCalc {
		return Request{}, fmt.Errorf("%w: invalid prefix %q, expected %q", ErrInvalidMessageFormat, parts[0], PrefixCalc)
	}

	seq, err := strconv.Atoi(parts[1])
	if err != nil {
		return Request{}, fmt.Errorf("%w: %q is not an integer", ErrInvalidSequence, parts[1])
	}
	op1, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return Request{}, fmt.Errorf("%w: op1 %q is not a number", ErrInvalidOperand, parts[2])
	}
	op2, err := strconv.ParseFloat(parts[4], 64)
	if err != nil {
		return Request{}, fmt.Errorf("%w: op2 %q is not a number", ErrInvalidOperand, parts[4])
	}

	return Request{
		Seq: seq,
		Op1: op1,
		Op:  parts[3],
		Op2: op2,
	}, nil
}

// FormatRequest serializes a Request struct into the textual wire format.
//
// Usage:
//
//	wire := protocol.FormatRequest(req)
func FormatRequest(req Request) string {
	return fmt.Sprintf("%s%s%d%s%g%s%s%s%g",
		PrefixCalc, FieldDelimiter,
		req.Seq, FieldDelimiter,
		req.Op1, FieldDelimiter,
		req.Op, FieldDelimiter,
		req.Op2,
	)
}

// FormatResponse serializes a successful calculation response.
//
// Usage:
//
//	wire := protocol.FormatResponse(0, 15.0)
func FormatResponse(seq int, result float64) string {
	formatted := strconv.FormatFloat(result, 'f', -1, 64)
	if !strings.Contains(formatted, ".") {
		formatted += ".0"
	}
	return fmt.Sprintf("%s%s%d%s%s", PrefixResult, FieldDelimiter, seq, FieldDelimiter, formatted)
}

// FormatError serializes an error response.
//
// Usage:
//
//	wire := protocol.FormatError(1, "divisão por zero")
func FormatError(seq int, msg string) string {
	return fmt.Sprintf("%s%s%d%s%s", PrefixError, FieldDelimiter, seq, FieldDelimiter, msg)
}

// ParseResponse parses a server response wire message into a Response struct.
//
// Usage:
//
//	resp, err := protocol.ParseResponse("RESULT:0:15.0")
func ParseResponse(raw string) (Response, error) {
	clean := strings.TrimSpace(raw)
	parts := strings.Split(clean, FieldDelimiter)
	if len(parts) < ExpectedResponseFields {
		return Response{}, fmt.Errorf("%w: expected at least %d fields, got %d", ErrInvalidMessageFormat, ExpectedResponseFields, len(parts))
	}

	seq, err := strconv.Atoi(parts[1])
	if err != nil {
		return Response{}, fmt.Errorf("%w: %q is not an integer", ErrInvalidSequence, parts[1])
	}

	switch parts[0] {
	case PrefixResult:
		res, parseErr := strconv.ParseFloat(parts[2], 64)
		if parseErr != nil {
			return Response{}, fmt.Errorf("%w: invalid result %q", ErrInvalidOperand, parts[2])
		}
		return Response{Seq: seq, Result: res, IsError: false}, nil
	case PrefixError:
		msg := strings.Join(parts[2:], FieldDelimiter)
		return Response{Seq: seq, ErrorMsg: msg, IsError: true}, nil
	default:
		return Response{}, fmt.Errorf("%w: unknown response prefix %q", ErrInvalidMessageFormat, parts[0])
	}
}
