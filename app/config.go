package main

import (
	"flag"
	"log"
)

var Config config

type config struct {
	serveDir         string
	maxFileSizeBytes int
}

func configure() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	serveDir := flag.String("directory", "/tmp/data/codecrafters.io/http-server-tester/", "Directory to serve files from")
	maxFileSizeBytes := flag.Int("max-file-size", 1000000, "Max accepted file size in Bytes [1MB]")
	flag.Parse()

	Config = config{
		serveDir:         *serveDir,
		maxFileSizeBytes: *maxFileSizeBytes,
	}

	log.Println("server configured:", Config)
}
