package main

import (
	"fmt"
	"log"
)

const (
	statusOK                  = "200"
	statusCreated             = "201"
	statusNotFound            = "404"
	statusMethodNotAllowed    = "405"
	statusInternalServerError = "500"
)

type response struct {
	version string
	status  string
	headers []header
	body    string
}

func newResponse() *response {
	return &response{
		version: version_11,
		headers: []header{},
	}
}

func (r *response) setHeader(header, value string) {
	r.headers = append(r.headers, *newHeader(header, value))
}

func (r *response) setStatus(status string) {
	r.status = status
}

func (r *response) Status() string {
	return r.status
}

func (r *response) setBody(body string) {
	r.body = body
}

func (r *response) ToString() string {
	headers := ""
	for _, h := range r.headers {
		headers += h.ToString()
	}
	return fmt.Sprintf("%s %s %s\r\n%s\r\n%s", r.version, r.status, statusText(r.status), headers, r.body)
}

func statusText(status string) string {
	s := ""
	switch status {
	case statusCreated:
		s = "Created"
	case statusMethodNotAllowed:
		s = "Method Not Allowed"
	case statusInternalServerError:
		s = "Internal Server Error"
	case statusNotFound:
		s = "Not Found"
	case statusOK:
		s = "OK"
	default:
		log.Fatalln("invalid status", status)
	}
	return s
}
