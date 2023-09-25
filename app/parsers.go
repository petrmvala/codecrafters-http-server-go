package main

import (
	"errors"
	"strings"
)

func parseHeader(header string) (*httpHeader, error) {

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

func parseRequest(r string) (*httpRequest, error) {
	req := &httpRequest{}

	// If the string cannot be split, body is set to ""
	head, body, _ := strings.Cut(r, "\r\n\r\n")
	req.body = body

	headlines := strings.Split(head, "\r\n")

	firstLine := strings.Split(headlines[0], " ")
	if len(firstLine) != 3 {
		return nil, errors.New("malformed request")
	}
	req.startLine = startLine{firstLine[0], firstLine[1], firstLine[2]}

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
