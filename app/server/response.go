package server

import (
	"bytes"
	"fmt"
	"log"
)

const (
	StatusOK      = "200"
	StatusCreated = "201"

	StatusBadRequest       = "400"
	StatusForbiden         = "403"
	StatusNotFound         = "404"
	StatusMethodNotAllowed = "405"
	StatusLengthRequired   = "411"
	StatusContentTooLarge  = "413"

	StatusInternalServerError = "500"
)

type Response struct {
	status  string
	headers headers
	Body    []byte
}

func NewResponse() *Response {
	return &Response{
		headers: headers{},
	}
}

func (r *Response) AddHeader(header string, value string) {
	val, ok := r.headers[header]
	if !ok {
		val = []string{}
	}
	r.headers[header] = append(val, value)
}

func (r *Response) SetHeader(header string, value string) {
	r.headers[header] = []string{value}
}

func (r *Response) SetStatus(status string) {
	r.status = status
}

func (r *Response) SetBody(body []byte) {
	r.Body = body
}

func (r *Response) Bytes() []byte {
	head := fmt.Sprintf("%s %s %s\r\n%s", version11, r.status, statusText(r.status), r.headers.ToString())
	b := [][]byte{[]byte(head), r.Body}
	return bytes.Join(b, []byte("\r\n"))
}

func statusText(status string) string {
	s := ""
	switch status {
	case StatusOK:
		s = "OK"
	case StatusCreated:
		s = "Created"
	case StatusBadRequest:
		s = "Bad Request"
	case StatusForbiden:
		s = "Forbidden"
	case StatusNotFound:
		s = "Not Found"
	case StatusMethodNotAllowed:
		s = "Method Not Allowed"
	case StatusLengthRequired:
		s = "Length Required"
	case StatusContentTooLarge:
		s = "Content Too Large"
	case StatusInternalServerError:
		s = "Internal Server Error"
	default:
		log.Fatalln("invalid status", status)
	}
	return s
}
