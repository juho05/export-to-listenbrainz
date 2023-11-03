package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"unicode"
)

type jsonObjectArrayScanner[T any] struct {
	reader           *bufio.Reader
	openBracketCount int
	openParenCount   int
	buffer           *bytes.Buffer
}

func newJsonObjectArrayScanner[T any](reader io.Reader) (*jsonObjectArrayScanner[T], error) {
	scanner := &jsonObjectArrayScanner[T]{
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

func (s *jsonObjectArrayScanner[T]) nextObject() (T, error) {
	var obj T
	for {
		r, _, err := s.reader.ReadRune()
		if err != nil {
			return obj, fmt.Errorf("read json object: %w", err)
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
				return obj, io.EOF
			}
		}
		return obj, fmt.Errorf("read json object: unexpected character: %c", r)
	}

	for s.openParenCount > 0 {
		r, _, err := s.reader.ReadRune()
		if err != nil {
			return obj, fmt.Errorf("read json object: %w", err)
		}
		if r == '{' {
			s.openParenCount++
		}
		if r == '}' {
			s.openParenCount--
		}
		s.buffer.WriteRune(r)
	}

	err := json.Unmarshal(s.buffer.Bytes(), &obj)
	s.buffer.Reset()
	if err != nil {
		return obj, fmt.Errorf("decode json object: %w", err)
	}
	return obj, nil
}
