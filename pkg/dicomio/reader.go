package dicomio

import (
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
)

var (
	ErrorInsufficientBytesLeft = errors.New("not enough bytes left until buffer limit to complete this operation")
)

type Reader interface {
	io.Reader
	ReadUInt16() (uint16, error)
	ReadUInt32() (uint32, error)
	ReadInt16() (int16, error)
	ReadInt32() (int32, error)
	ReadString(n uint32) (string, error)
	Skip(n int64) error
	PushLimit(n int64)
	PopLimit()
	IsLimitExhausted() bool
	BytesLeftUntilLimit() int64
}

type reader struct {
	in         io.Reader
	bo         binary.ByteOrder
	limit      int64
	bytesRead  int64
	limitStack []int64
}

func NewReader(in io.Reader, bo binary.ByteOrder, limit int64) (Reader, error) {
	return &reader{
		in:        in,
		bo:        bo,
		limit:     limit,
		bytesRead: 0,
	}, nil
}

func (r *reader) BytesLeftUntilLimit() int64 {
	return r.limit - r.bytesRead
}

// Read
func (r *reader) Read(p []byte) (int, error) {
	// Check if we've hit the limit
	if r.BytesLeftUntilLimit() <= 0 {
		if len(p) == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}

	// If asking for more than we have left, just return whatever we've got left
	// TODO: return a special kind of error if this situation occurs to inform the caller
	if int64(len(p)) > r.BytesLeftUntilLimit() {
		p = p[:r.BytesLeftUntilLimit()]
	}
	n, err := r.in.Read(p)
	if n >= 0 {
		r.bytesRead += int64(n)
	}
	return n, err
}

// TODO: Implement the rest of these interface functions
func (r *reader) ReadUInt16() (uint16, error) {
	var out uint16
	err := binary.Read(r, r.bo, &out)
	return out, err
}

func (r *reader) ReadUInt32() (uint32, error) {
	var out uint32
	err := binary.Read(r, r.bo, &out)
	return out, err
}

func (r *reader) ReadInt16() (int16, error) {
	var out int16
	err := binary.Read(r, r.bo, &out)
	return out, err
}

func (r *reader) ReadInt32() (int32, error) {
	var out int32
	err := binary.Read(r, r.bo, &out)
	return out, err
}
func (r *reader) ReadString(n uint32) (string, error) {
	data := make([]byte, n)
	_, err := io.ReadFull(r, data)
	// TODO: add support for different coding systems
	return string(data), err
}
func (r *reader) Skip(n int64) error {
	if r.BytesLeftUntilLimit() < n {
		// not enough left to skip
		return ErrorInsufficientBytesLeft
	}

	_, err := io.CopyN(ioutil.Discard, r, n)

	return err
}

// PushLimit creates a limit n bytes from the current position
func (r *reader) PushLimit(n int64) {
	newLimit := r.bytesRead + n
	if newLimit > r.limit {
		// error
	}

	// Add current limit to the stack
	r.limitStack = append(r.limitStack, r.limit)
	r.limit = newLimit
}
func (r *reader) PopLimit() {
	if r.bytesRead < r.limit {
		// didn't read all the way to the limit, so skip over what's left.
		_ = r.Skip(r.limit - r.bytesRead)
	}
	last := len(r.limitStack) - 1
	r.limit = r.limitStack[last]
	r.limitStack = r.limitStack[:last]
}

func (r *reader) IsLimitExhausted() bool {
	return r.BytesLeftUntilLimit() <= 0
}
