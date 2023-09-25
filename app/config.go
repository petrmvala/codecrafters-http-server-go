package main

import (
	"flag"
	"log"
)

var Config config

type config struct {
	serveDir string
}

func configureServer() {
	serveDir := flag.String("directory", "/tmp/data/codecrafters.io/http-server-tester/", "Directory to serve files from")
	flag.Parse()

	Config.serveDir = *serveDir

	log.Println("server configured:", Config)
}
