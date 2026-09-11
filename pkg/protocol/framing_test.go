package protocol

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestFramingRoundtrip(t *testing.T) {
	// Arrange
	testData := []byte("Hello, Protobuf over TCP stream framing!")
	var buf bytes.Buffer

	// Act
	err := WriteMsg(&buf, testData)
	if err != nil {
		t.Fatalf("WriteMsg failed: %v", err)
	}

	received, err := ReadMsg(&buf)

	// Assert
	if err != nil {
		t.Fatalf("ReadMsg failed: %v", err)
	}
	if !bytes.Equal(received, testData) {
		t.Errorf("expected %q, got %q", string(testData), string(received))
	}
}

func TestFramingMultipleMessages(t *testing.T) {
	// Arrange
	msg1 := []byte("first message")
	msg2 := []byte("second message with more data")
	var buf bytes.Buffer

	// Act
	if err := WriteMsg(&buf, msg1); err != nil {
		t.Fatalf("WriteMsg msg1 failed: %v", err)
	}
	if err := WriteMsg(&buf, msg2); err != nil {
		t.Fatalf("WriteMsg msg2 failed: %v", err)
	}

	recv1, err1 := ReadMsg(&buf)
	recv2, err2 := ReadMsg(&buf)

	// Assert
	if err1 != nil || err2 != nil {
		t.Fatalf("ReadMsg failed: %v, %v", err1, err2)
	}
	if !bytes.Equal(recv1, msg1) || !bytes.Equal(recv2, msg2) {
		t.Errorf("messages do not match")
	}
}

func TestFramingEOFOnEmptyReader(t *testing.T) {
	// Arrange
	emptyBuf := bytes.NewReader([]byte{})

	// Act
	_, err := ReadMsg(emptyBuf)

	// Assert
	if !errors.Is(err, io.EOF) {
		t.Errorf("expected io.EOF, got %v", err)
	}
}

func TestFramingTruncatedPayload(t *testing.T) {
	// Arrange: 4 bytes length prefix indicating 10 bytes, but only 3 bytes provided
	truncated := []byte{0x00, 0x00, 0x00, 0x0A, 'a', 'b', 'c'}
	buf := bytes.NewReader(truncated)

	// Act
	_, err := ReadMsg(buf)

	// Assert
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("expected io.ErrUnexpectedEOF, got %v", err)
	}
}
