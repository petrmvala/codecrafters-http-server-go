package main

import (
	"errors"
	"fmt"
	"strings"
)

type request struct {
	version string
	method  string
	target  string
	headers headers
	body    string
}

func newRequest() *request {
	return &request{
		version: version_11,
		headers: headers{},
	}
}

func (r *request) isValid() bool {
	if _, ok := knownVersions[r.version]; !ok {
		return false
	}
	if _, ok := knownMethods[r.method]; !ok {
		return false
	}
	// target not validated
	// headers are validated during parsing
	return true
}

func (r *request) ToString() string {
	return fmt.Sprintf("%s %s %s\r\n%s\r\n%s", r.method, r.target, r.version, r.headers.ToString(), r.body)
}

func (r *request) Path() string {
	return r.target
}

func (r *request) Method() string {
	return r.method
}

func parseRequest(r string) (*request, error) {
	req := newRequest()

	// If the string cannot be split, body is set to ""
	head, body, _ := strings.Cut(r, "\r\n\r\n")
	req.body = body

	headlines := strings.Split(head, "\r\n")

	firstLine := strings.Split(headlines[0], " ")
	if len(firstLine) != 3 {
		return nil, errors.New("malformed request")
	}
	req.method, req.target, req.version = firstLine[0], firstLine[1], firstLine[2]

	if len(headlines) > 1 {
		req.headers = parseHeaders(headlines[1:])
	}

	if !req.isValid() {
		return nil, errors.New("invalid request")
	}

	return req, nil
}
