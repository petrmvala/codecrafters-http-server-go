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
	headers []header
	body    string
}

func (r *request) isValid() bool {
	if _, ok := knownVersions[r.version]; !ok {
		return false
	}
	if _, ok := knownMethods[r.method]; !ok {
		return false
	}
	// target not validated
	for _, h := range r.headers {
		if !h.isValid() {
			return false
		}
	}
	return true
}

func (r *request) ToString() string {
	headers := ""
	for _, h := range r.headers {
		headers += h.ToString()
	}
	return fmt.Sprintf("%s %s %s\r\n%s\r\n%s", r.method, r.target, r.version, headers, r.body)
}

func (r *request) Header(header string) (string, error) {
	for _, v := range r.headers {
		if v.name == header {
			return v.value, nil
		}
	}
	return "", errors.New("header not found")
}

func (r *request) Path() string {
	return r.target
}

func (r *request) Method() string {
	return r.method
}

func parseRequest(r string) (*request, error) {
	req := &request{}

	// If the string cannot be split, body is set to ""
	head, body, _ := strings.Cut(r, "\r\n\r\n")
	req.body = body

	headlines := strings.Split(head, "\r\n")

	firstLine := strings.Split(headlines[0], " ")
	if len(firstLine) != 3 {
		return nil, errors.New("malformed request")
	}
	req.method, req.target, req.version = firstLine[0], firstLine[1], firstLine[2]

	// parse headers
	if len(headlines) > 1 {
		for _, h := range headlines[1:] {
			header, err := parseHeader(h)
			if err != nil {
				// ignore invalid headers, maybe we should act differently
				continue
			}
			req.headers = append(req.headers, *header)
		}
	}

	if !req.isValid() {
		return nil, errors.New("invalid request")
	}

	return req, nil
}
