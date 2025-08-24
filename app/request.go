package main

import (
	"errors"
	"strings"
)

type request struct {
	method  string
	target  string
	headers headers
	body    string
}

func parseRequest(r string) (*request, error) {
	req := &request{
		headers: headers{},
	}

	// If the string cannot be split, body is set to ""
	head, body, _ := strings.Cut(r, "\r\n\r\n")
	req.body = body

	headlines := strings.Split(head, "\r\n")
	if len(headlines) > 1 {
		req.headers = parseHeaders(headlines[1:])
	}

	firstLine := strings.Split(headlines[0], " ")
	if len(firstLine) != 3 {
		return nil, errors.New("error parsing head")
	}

	req.method = firstLine[0]
	req.target = firstLine[1]
	version := firstLine[2]
	if version != version11 {
		return nil, errors.New("invalid version")
	}
	if req.method != methodGet && req.method != methodPost {
		return nil, errors.New("invalid method")
	}

	return req, nil
}
