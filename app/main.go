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

	s := server.NewServer("4221")

	d := server.NewDistributor()
	d.Get("/", handleRootResponse)
	d.Get("/user-agent", handleUserAgent)
	d.Get("/echo/", handleEchoResponse)
	d.Get("/files/", dir.handleFileRequest())
	d.Post("/files/", dir.handleFilePost())

	s.Distributor = *d

	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}
