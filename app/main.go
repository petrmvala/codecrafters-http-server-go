package main

import (
	"flag"
	"log"
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

	s := NewServer("4221")

	d := newDistributor()
	d.get("/", handleRootResponse)
	d.get("/user-agent", handleUserAgent)
	d.get("/echo/", handleEchoResponse)
	d.get("/files/", dir.handleFileRequest())
	d.post("/files/", dir.handleFilePost())

	s.distributor = *d

	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}
