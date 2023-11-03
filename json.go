package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"unicode"
)

type jsonObjectArrayScanner struct {
	reader           *bufio.Reader
	openBracketCount int
	openParenCount   int
	buffer           *bytes.Buffer
}

func newJsonObjectArrayScanner(reader io.Reader) (*jsonObjectArrayScanner, error) {
	scanner := &jsonObjectArrayScanner{
		reader: bufio.NewReader(reader),
		buffer: new(bytes.Buffer),
	}
	for {
		r, _, err := scanner.reader.ReadRune()
		if err != nil {
			return nil, fmt.Errorf("read [ character: %w", err)
		}
		if unicode.IsSpace(r) {
			continue
		}
		if r == '[' {
			scanner.openBracketCount++
			break
		}
		return nil, errors.New("json object is not an array")
	}
	return scanner, nil
}

func (s *jsonObjectArrayScanner) nextObject() ([]byte, error) {
	for {
		r, _, err := s.reader.ReadRune()
		if err != nil {
			return nil, fmt.Errorf("read json object: %w", err)
		}
		if unicode.IsSpace(r) || r == ',' {
			continue
		}
		if r == '{' {
			s.buffer.WriteRune(r)
			s.openParenCount++
			break
		}
		if r == ']' {
			s.openBracketCount--
			if s.openBracketCount == 0 {
				return nil, io.EOF
			}
		}
		return nil, fmt.Errorf("read json object: unexpected character: %c", r)
	}

	for s.openParenCount > 0 {
		r, _, err := s.reader.ReadRune()
		if err != nil {
			return nil, fmt.Errorf("read json object: %w", err)
		}
		if r == '{' {
			s.openParenCount++
		}
		if r == '}' {
			s.openParenCount--
		}
		s.buffer.WriteRune(r)
	}

	bytes := make([]byte, len(s.buffer.Bytes()))
	copy(bytes, s.buffer.Bytes())
	s.buffer.Reset()
	return bytes, nil
}
