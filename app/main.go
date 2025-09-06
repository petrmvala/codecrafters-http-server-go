package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"log"
	"strconv"

	"github.com/codecrafters-io/http-server-starter-go/app/server"
)

type ServeDir struct {
	directory string
}

func WithGzip(req *server.Request) *server.Response {
	res := handleEchoResponse(req)

	gzipHdr := false
	if enc, ok := req.Headers[server.HeaderAcceptEncoding]; ok {
		for _, e := range enc {
			if e == "gzip" {
				gzipHdr = true
				break
			}
		}
	}
	if !gzipHdr {
		return res
	}

	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	_, err := w.Write(res.Body)
	if err != nil {
		log.Println("failed to compress:", err.Error())
		return res
	}
	if err := w.Close(); err != nil {
		log.Println("failed to close:", err.Error())
		return res
	}

	res.SetHeader(server.HeaderContentLength, strconv.Itoa(buf.Len()))
	res.SetHeader(server.HeaderContentEncoding, "gzip")
	res.SetBody(buf.Bytes())
	return res
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	serveDir := flag.String("directory", "/tmp/data/codecrafters.io/http-server-tester/", "Directory to serve files from")
	flag.Parse()

	dir := &ServeDir{
		directory: *serveDir,
	}

	s := server.NewServer("0.0.0.0:4221")

	s.Register("/", "GET", handleRootResponse)
	s.Register("/user-agent", "GET", handleUserAgent)
	s.Register("/echo/", "GET", WithGzip)
	s.Register("/files/", "GET", dir.handleFileRequest())
	s.Register("/files/", "POST", dir.handleFilePost())

	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}
