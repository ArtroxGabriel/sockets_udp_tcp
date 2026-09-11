package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	PrefixLengthBytes = 4
	MaxMessageSize    = 10 * 1024 * 1024 // 10 MB sanity limit
)

var (
	ErrMessageTooLarge = errors.New("message exceeds maximum allowed size")
)

// WriteMsg encodes and writes a length-prefixed binary message to the writer.
// Format: 4-byte big-endian uint32 payload length followed by the payload bytes.
//
// Usage:
//
//	err := protocol.WriteMsg(conn, payload)
func WriteMsg(w io.Writer, payload []byte) error {
	length := uint32(len(payload))
	if length > MaxMessageSize {
		return fmt.Errorf("%w: length %d > max %d", ErrMessageTooLarge, length, MaxMessageSize)
	}

	header := make([]byte, PrefixLengthBytes)
	binary.BigEndian.PutUint32(header, length)

	if _, err := w.Write(header); err != nil {
		return fmt.Errorf("failed to write length prefix: %w", err)
	}
	if _, err := w.Write(payload); err != nil {
		return fmt.Errorf("failed to write payload: %w", err)
	}
	return nil
}

// ReadMsg reads a 4-byte length prefix and then reads the exact payload from the reader.
//
// Usage:
//
//	data, err := protocol.ReadMsg(conn)
func ReadMsg(r io.Reader) ([]byte, error) {
	header := make([]byte, PrefixLengthBytes)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(header)
	if length > MaxMessageSize {
		return nil, fmt.Errorf("%w: length %d > max %d", ErrMessageTooLarge, length, MaxMessageSize)
	}

	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}
	return payload, nil
}
