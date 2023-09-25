package main

import (
	"errors"
	"fmt"
	"log"
)

const (
	httpVersion_11 = "HTTP/1.1"

	httpMethodGet  = "GET"
	httpMethodPost = "POST"

	httpStatusOK                  = "200"
	HttpStatusCreated             = "201"
	httpStatusNotFound            = "404"
	httpStatusMethodNotAllowed    = "405"
	httpStatusInternalServerError = "500"

	httpStatusTextOK                  = "OK"
	HttpStatusTextCreated             = "Created"
	httpStatusTextNotFound            = "Not Found"
	httpStatusTextMethodNotAllowed    = "Method Not Allowed"
	httpStatusTextInternalServerError = "Internal Server Error"

	httpHeaderAllow         = "Allow"
	httpHeaderContentType   = "Content-Type"
	httpHeaderContentLength = "Content-Length"
	httpHeaderUserAgent     = "User-Agent"
)

type httpMethod string

type httpHeader struct {
	name  string
	value string
}

func newHeader(header, value string) *httpHeader {
	var h httpHeader

	h.name = header
	h.value = value

	return &h
}

func (h *httpHeader) isValid() bool {
	// Header values are not validated
	switch h.name {
	case httpHeaderAllow, httpHeaderContentLength, httpHeaderContentType, httpHeaderUserAgent:
		return true
	}
	return false
}

func (h *httpHeader) ToString() string {
	return fmt.Sprintf("%s: %s\r\n", h.name, h.value)
}

type startLine struct {
	httpMethod    string
	requestTarget string
	httpVersion   string
}

func (s *startLine) isValid() bool {
	// Target value is not validated
	if s.httpVersion != httpVersion_11 {
		return false // we don't support different versions
	}

	switch s.httpMethod {
	case httpMethodGet, httpMethodPost:
		return true
	default:
		return false
	}
}

func (s *startLine) ToString() string {
	return fmt.Sprintf("%s %s %s\r\n", s.httpMethod, s.requestTarget, s.httpVersion)
}

type httpRequest struct {
	startLine startLine
	headers   []httpHeader
	body      string
}

func (r *httpRequest) isValid() bool {
	// We don't validate body
	if !r.startLine.isValid() {
		return false
	}

	for _, h := range r.headers {
		if !h.isValid() {
			return false
		}
	}

	return true
}

func (r *httpRequest) ToString() string {
	head := r.startLine.ToString()

	for _, h := range r.headers {
		head += h.ToString()
	}

	return fmt.Sprintf("%s\r\n%s", head, r.body)
}

func (r *httpRequest) Header(header string) (string, error) {
	for _, v := range r.headers {
		if v.name == header {
			return v.value, nil
		}
	}
	return "", errors.New("header not found")
}

func (r *httpRequest) Path() string {
	return r.startLine.requestTarget
}

func (r *httpRequest) Method() string {
	return r.startLine.httpMethod
}

type statusLine struct {
	httpVersion    string
	httpStatusCode string
	httpStatusText string
}

func newStatusLine(status string) *statusLine {
	sl := statusLine{httpVersion: httpVersion_11}

	switch status {
	case HttpStatusCreated:
		sl.httpStatusCode = HttpStatusCreated
		sl.httpStatusText = HttpStatusTextCreated
	case httpStatusMethodNotAllowed:
		sl.httpStatusCode = httpStatusMethodNotAllowed
		sl.httpStatusText = httpStatusTextMethodNotAllowed
	case httpStatusInternalServerError:
		sl.httpStatusCode = httpStatusInternalServerError
		sl.httpStatusText = httpStatusTextInternalServerError
	case httpStatusNotFound:
		sl.httpStatusCode = httpStatusNotFound
		sl.httpStatusText = httpStatusTextNotFound
	case httpStatusOK:
		sl.httpStatusCode = httpStatusOK
		sl.httpStatusText = httpStatusTextOK
	default:
		log.Fatalln("invalid status", status)
	}

	return &sl
}

func (s *statusLine) ToString() string {
	return fmt.Sprintf("%s %s %s\r\n", s.httpVersion, s.httpStatusCode, s.httpStatusText)
}

type httpResponse struct {
	statusLine statusLine
	headers    []httpHeader
	body       string
}

func (r *httpResponse) setHeader(header, value string) {
	r.headers = append(r.headers, *newHeader(header, value))
}

func (r *httpResponse) setStatus(status string) {
	r.statusLine = *newStatusLine(status)
}

func (r *httpResponse) setBody(body string) {
	r.body = body
}

func (r *httpResponse) ToString() string {
	head := r.statusLine.ToString()

	for _, h := range r.headers {
		head += h.ToString()
	}

	return fmt.Sprintf("%s\r\n%s", head, r.body)
}

func (r *httpResponse) Status() string {
	return r.statusLine.httpStatusCode
}
