package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

const (
	headerAcceptEncoding  = "Accept-Encoding"
	headerAllow           = "Allow"
	headerContentEncoding = "Content-Encoding"
	headerContentType     = "Content-Type"
	headerContentLength   = "Content-Length"
	headerUserAgent       = "User-Agent"
)

func (h *headers) ToString() string {
	var b strings.Builder
	for hdr, val := range *h {
		var v string
		if hdr == headerContentLength {
			v = strconv.Itoa(val.(int))
		} else if hdr == headerAcceptEncoding {
			l := []string{}
			for e, _ := range val.(map[string]bool) {
				l = append(l, e)
			}
			v = strings.Join(l, ",")
		} else {
			v = val.(string)
		}
		fmt.Fprintf(&b, "%s: %s\r\n", hdr, v)
	}
	return b.String()
}

type headers map[string]any

func parseHeaders(data []string) headers {
	h := headers{}
	for _, l := range data {
		key, value, found := strings.Cut(l, ": ")
		if !found {
			log.Println("cannot parse header, skipping")
			continue
		}

		switch key {
		case headerAcceptEncoding: // https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Accept-Encoding
			e := map[string]bool{}
			for _, enc := range strings.Split(value, ",") {
				e[strings.TrimSpace(enc)] = true
			}
			h[key] = e
		case headerAllow:
			h[key] = value
		case headerContentEncoding:
			h[key] = value
		case headerContentType:
			h[key] = value
		case headerContentLength:
			len, err := strconv.Atoi(value)
			if err != nil {
				log.Println("cannot parse header, skipping")
				continue
			}
			h[key] = len
		case headerUserAgent:
			h[key] = value
		default:
			log.Println("invalid header, skipping:", key, value)
			continue
		}
	}
	return h
}
