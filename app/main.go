package main

import (
	"flag"
	"log"

	"github.com/codecrafters-io/http-server-starter-go/app/server"
)

type ServeDir struct {
	directory string
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	serveDir := flag.String("directory", "/tmp/data/codecrafters.io/http-server-tester/", "Directory to serve files from")
	flag.Parse()

	dir := &ServeDir{
		directory: *serveDir,
	}

	s := server.NewServer("0.0.0.0:4221")

	s.Get("/", handleRootResponse)
	s.Get("/user-agent", handleUserAgent)
	s.Get("/echo/", handleEchoResponse)
	s.Get("/files/", dir.handleFileRequest())
	s.Post("/files/", dir.handleFilePost())

	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}
