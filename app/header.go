package main

import (
	"errors"
	"fmt"
	"strings"
)

const (
	headerAllow         = "Allow"
	headerContentType   = "Content-Type"
	headerContentLength = "Content-Length"
	headerUserAgent     = "User-Agent"
)

var knownHeaders = map[string]bool{
	headerAllow:         true,
	headerContentType:   true,
	headerContentLength: true,
	headerUserAgent:     true,
}

type header struct {
	name  string
	value string
}

func newHeader(key, val string) *header {
	return &header{
		name:  key,
		value: val,
	}
}

func (h *header) isValid() bool {
	if _, ok := knownHeaders[h.name]; !ok {
		return false
	}
	return true
}

func (h *header) ToString() string {
	return fmt.Sprintf("%s: %s\r\n", h.name, h.value)
}

func parseHeader(header string) (*header, error) {

	key, value, found := strings.Cut(header, ": ")
	if !found {
		return nil, errors.New("invalid header")
	}

	h := newHeader(key, value)
	if !h.isValid() {
		return nil, errors.New("invalid header")
	}

	return h, nil
}
