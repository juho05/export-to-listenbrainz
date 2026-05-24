package main

import (
	"encoding/json"
	"fmt"
	"io"
)

type jsonObjectArrayScanner[T any] struct {
	decoder *json.Decoder
}

func newJsonObjectArrayScanner[T any](reader io.Reader) (*jsonObjectArrayScanner[T], error) {
	decoder := json.NewDecoder(reader)
	// Read the open bracket '['
	t, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("read [ character: %w", err)
	}
	delim, ok := t.(json.Delim)
	if !ok || delim != '[' {
		return nil, fmt.Errorf("json object is not an array")
	}
	return &jsonObjectArrayScanner[T]{
		decoder: decoder,
	}, nil
}

func (s *jsonObjectArrayScanner[T]) nextObject() (T, error) {
	var obj T
	if !s.decoder.More() {
		// Read the closing bracket ']'
		t, err := s.decoder.Token()
		if err != nil {
			return obj, fmt.Errorf("read json object: %w", err)
		}
		delim, ok := t.(json.Delim)
		if !ok || delim != ']' {
			return obj, fmt.Errorf("read json object: unexpected character: %v", t)
		}
		return obj, io.EOF
	}
	err := s.decoder.Decode(&obj)
	if err != nil {
		return obj, fmt.Errorf("decode json object: %w", err)
	}
	return obj, nil
}
