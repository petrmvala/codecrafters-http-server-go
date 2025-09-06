package server

import (
	"fmt"
	"log"
	"strings"
)

const (
	HeaderAcceptEncoding  = "Accept-Encoding"
	HeaderAllow           = "Allow"
	HeaderConnection      = "Connection"
	HeaderContentEncoding = "Content-Encoding"
	HeaderContentType     = "Content-Type"
	HeaderContentLength   = "Content-Length"
	HeaderUserAgent       = "User-Agent"
)

type headers map[string][]string

func (h *headers) ToString() string {
	var b strings.Builder
	for hdr, val := range *h {
		v := strings.Join(val, ",")
		fmt.Fprintf(&b, "%s: %s\r\n", hdr, v)
	}
	return b.String()
}

func parseHeaders(data []string) headers {
	h := headers{}
	for _, hdrLine := range data {
		key, value, found := strings.Cut(hdrLine, ":")
		if !found {
			log.Println("cannot parse header:", hdrLine)
			continue
		}

		values := []string{}
		for _, v := range strings.Split(value, ",") {
			values = append(values, strings.TrimSpace(v))
		}

		hdrVal, ok := h[key]
		if !ok {
			hdrVal = []string{}
		}
		h[key] = append(hdrVal, values...)
	}
	return h
}
