package server

import (
	"errors"
	"strings"
)

type Request struct {
	Method  string
	Target  string
	Headers headers
	Body    string
}

func parseRequest(r string) (*Request, error) {
	req := &Request{
		Headers: headers{},
	}

	// If the string cannot be split, body is set to ""
	head, body, _ := strings.Cut(r, "\r\n\r\n")
	req.Body = body

	headlines := strings.Split(head, "\r\n")
	if len(headlines) > 1 {
		req.Headers = parseHeaders(headlines[1:])
	}

	firstLine := strings.Split(headlines[0], " ")
	if len(firstLine) != 3 {
		return nil, errors.New("error parsing head")
	}

	req.Method = firstLine[0]
	req.Target = firstLine[1]
	version := firstLine[2]
	if version != version11 {
		return nil, errors.New("invalid version")
	}
	if req.Method != methodGet && req.Method != methodPost {
		return nil, errors.New("invalid method")
	}

	return req, nil
}
