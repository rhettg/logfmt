package logfmt

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

// EndOfRecord indicates that no more keys or values exist to decode in the
// current record. Use Decoder.ScanRecord to advance to the next record.
var EndOfRecord = errors.New("end of record")

// A Decoder reads and decodes logfmt records from an input stream.
type Decoder struct {
	pos     int
	key     []byte
	value   []byte
	lineNum int
	s       *bufio.Scanner
	err     error
}

// NewDecoder returns a new decoder that reads from r.
//
// The decoder introduces its own buffering and may read data from r beyond
// the logfmt records requested.
func NewDecoder(r io.Reader) *Decoder {
	dec := &Decoder{
		s: bufio.NewScanner(r),
	}
	return dec
}

// ScanRecord advances the Decoder to the next record, which can then be
// parsed with the ScanKey and ScanValue methods. It returns false when
// decoding stops, either by reaching the end of the input or an error. After
// ScanRecord returns false, the Err method will return any error that
// occurred during decoding, except that if it was io.EOF, Err will return
// nil.
func (dec *Decoder) ScanRecord() bool {
	if dec.err != nil {
		return false
	}
	if !dec.s.Scan() {
		dec.err = dec.s.Err()
		return false
	}
	dec.lineNum++
	dec.pos = 0
	return true
}

func (dec *Decoder) ScanKeyval() bool {
	dec.key, dec.value = nil, nil
	if dec.err != nil {
		return false
	}

	line := dec.s.Bytes()

	// garbage
	for p, c := range line[dec.pos:] {
		if c > ' ' {
			dec.pos += p
			goto key
		}
	}
	dec.pos = len(line)
	return false

key:
	start := dec.pos
	for p, c := range line[dec.pos:] {
		switch {
		case c == '=':
			dec.pos += p
			if dec.pos > start {
				dec.key = line[start:dec.pos]
			}
			if dec.key == nil {
				dec.unexpectedByte(c)
				return false
			}
			goto equal
		case c == '"':
			dec.pos += p
			dec.unexpectedByte(c)
			return false
		case c <= ' ':
			dec.pos += p
			if dec.pos > start {
				dec.key = line[start:dec.pos]
			}
			return true
		}
	}
	dec.pos = len(line)
	if dec.pos > start {
		dec.key = line[start:dec.pos]
	}
	return true

equal:
	dec.pos++
	if dec.pos >= len(line) {
		return true
	}
	switch c := line[dec.pos]; {
	case c <= ' ':
		return true
	case c == '"':
		goto qvalue
	}

	// value
	start = dec.pos
	for p, c := range line[dec.pos:] {
		switch {
		case c == '=' || c == '"':
			dec.pos += p
			dec.unexpectedByte(c)
			return false
		case c <= ' ':
			dec.pos += p
			if dec.pos > start {
				dec.value = line[start:dec.pos]
			}
			return true
		}
	}
	dec.pos = len(line)
	if dec.pos > start {
		dec.value = line[start:dec.pos]
	}
	return true

qvalue:
	const (
		untermQuote  = "unterminated quoted value"
		invalidQuote = "invalid quoted value"
	)

	hasEsc, esc := false, false
	start = dec.pos
	for p, c := range line[dec.pos+1:] {
		switch {
		case esc:
			esc = false
		case c == '\\':
			hasEsc, esc = true, true
		case c == '"':
			dec.pos += p + 2
			if hasEsc {
				v, ok := unquoteBytes(line[start:dec.pos])
				if !ok {
					dec.syntaxError(invalidQuote)
					return false
				}
				dec.value = v
			} else {
				start++
				end := dec.pos - 1
				if end > start {
					dec.value = line[start:end]
				}
			}
			return true
		}
	}
	dec.pos = len(line)
	dec.syntaxError(untermQuote)
	return false
}

func (dec *Decoder) Key() []byte {
	return dec.key
}

func (dec *Decoder) Value() []byte {
	return dec.value
}

func (dec *Decoder) Err() error {
	return dec.err
}

// func (dec *Decoder) DecodeValue() ([]byte, error) {
// }

func (dec *Decoder) syntaxError(msg string) {
	dec.err = &SyntaxError{
		Msg:  msg,
		Line: dec.lineNum,
		Pos:  dec.pos + 1,
	}
}

func (dec *Decoder) unexpectedByte(c byte) {
	dec.err = &SyntaxError{
		Msg:  fmt.Sprintf("unexpected %q", c),
		Line: dec.lineNum,
		Pos:  dec.pos + 1,
	}
}

type SyntaxError struct {
	Msg  string
	Line int
	Pos  int
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("logfmt syntax error at pos %d on line %d: %s", e.Pos, e.Line, e.Msg)
}
