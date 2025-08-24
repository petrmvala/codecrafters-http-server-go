package main

import (
	"bytes"
	"fmt"
	"log"
)

const (
	statusOK      = "200"
	statusCreated = "201"

	statusBadRequest       = "400"
	statusForbiden         = "403"
	statusNotFound         = "404"
	statusMethodNotAllowed = "405"
	statusLengthRequired   = "411"
	statusContentTooLarge  = "413"

	statusInternalServerError = "500"
)

type response struct {
	status  string
	headers headers
	body    []byte
}

func newResponse() *response {
	return &response{
		headers: headers{},
	}
}

func (r *response) setHeader(header string, value any) {
	r.headers[header] = value
}

func (r *response) setStatus(status string) {
	r.status = status
}

func (r *response) setBody(body []byte) {
	r.body = body
}

func (r *response) ToString() string {
	return fmt.Sprintf("%s %s %s\r\n%s\r\n%s", version11, r.status, statusText(r.status), r.headers.ToString(), r.body)
}

func (r *response) Bytes() []byte {
	head := fmt.Sprintf("%s %s %s\r\n%s", version11, r.status, statusText(r.status), r.headers.ToString())
	b := [][]byte{[]byte(head), r.body}
	return bytes.Join(b, []byte("\r\n"))
}

func statusText(status string) string {
	s := ""
	switch status {
	case statusOK:
		s = "OK"
	case statusCreated:
		s = "Created"
	case statusBadRequest:
		s = "Bad Request"
	case statusForbiden:
		s = "Forbidden"
	case statusNotFound:
		s = "Not Found"
	case statusMethodNotAllowed:
		s = "Method Not Allowed"
	case statusLengthRequired:
		s = "Length Required"
	case statusContentTooLarge:
		s = "Content Too Large"
	case statusInternalServerError:
		s = "Internal Server Error"
	default:
		log.Fatalln("invalid status", status)
	}
	return s
}
