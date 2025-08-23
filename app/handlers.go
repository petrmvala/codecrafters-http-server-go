package main

import (
	"io"
	"log"
	"os"
)

// Respond with Status OK to GET /
func handleRootResponse(req *request) *response {
	res := newResponse()
	res.setStatus(statusOK)
	return res
}

// Respond with string echoed in body to GET /echo/<string>
func handleEchoResponse(req *request) *response {
	res := newResponse()

	str := req.Path()[6:]

	if enc, ok := req.headers[headerAcceptEncoding]; ok {
		if _, ok := enc.(map[string]bool)["gzip"]; ok {
			res.setHeader(headerContentEncoding, "gzip")
		}
	}

	res.setStatus(statusOK)
	res.setHeader(headerContentType, "text/plain")
	res.setHeader(headerContentLength, len(str))
	res.setBody(str)

	return res
}

// Respond with User Agent header value echoed in body to GET /user-agent
func handleUserAgent(req *request) *response {
	res := newResponse()

	ua, ok := req.headers[headerUserAgent]
	if !ok {
		res.setStatus(statusNotFound)
		return res
	}
	usrAg := ua.(string)

	res.setStatus(statusOK)
	res.setHeader(headerContentType, "text/plain")
	res.setHeader(headerContentLength, len(usrAg))
	res.setBody(usrAg)

	return res
}

// Respond with requested file served from directory to GET /files/<filename>
func handleFileRequest(req *request) *response {
	res := newResponse()

	filename := req.Path()[7:]

	data, err := os.ReadFile(Config.serveDir + "/" + filename)
	if err != nil {
		log.Println("file not found:", Config.serveDir+"/"+filename)
		res.setStatus(statusNotFound)

		return res
	}

	res.setStatus(statusOK)
	res.setHeader(headerContentType, "application/octet-stream")
	res.setHeader(headerContentLength, len(string(data)))
	res.setBody(string(data))

	return res
}

// Receive file and save it to directory via POST /files/<filename>
func handleFilePost(req *request) *response {
	res := newResponse()

	cl, ok := req.headers[headerContentLength]
	if !ok {
		log.Println("bad request: content-length header not received")
		res.setStatus(statusLengthRequired)
		return res
	}
	cBytes := cl.(int)

	if cBytes > Config.maxFileSizeBytes {
		log.Println("error: content too large")
		res.setStatus(statusContentTooLarge)
		return res
	}

	filename := req.Path()[7:]
	path := Config.serveDir + "/" + filename

	_, err := os.Stat(path)
	if err == nil {
		log.Println("error: file already exists")
		res.setStatus(statusForbiden)
		return res
	}

	file, err := os.Create(path)
	if err != nil {
		log.Println("error creating file:", err)
		res.setStatus(statusInternalServerError)
		return res
	}
	defer file.Close()

	content := req.body[:cBytes]
	_, err = io.WriteString(file, content)
	if err != nil {
		log.Println("error writing to file:", err)
		res.setStatus(statusInternalServerError)
		return res
	}

	log.Println(cBytes, " bytes written to ", path)
	res.setStatus(statusCreated)

	return res
}
