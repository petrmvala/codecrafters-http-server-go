package main

import (
	"log"
	"os"
	"strconv"
)

// Respond with Status OK to GET /
func handleRootResponse(req *httpRequest) *httpResponse {
	res := &httpResponse{}

	res.setStatus(httpStatusOK)

	return res
}

// Respond with string echoed in body to GET /echo/<string>
func handleEchoResponse(req *httpRequest) *httpResponse {
	res := &httpResponse{}

	str := req.Path()[6:]

	res.setStatus(httpStatusOK)
	res.setHeader(httpHeaderContentType, "text/plain")
	res.setHeader(httpHeaderContentLength, strconv.Itoa(len(str)))
	res.setBody(str)

	return res
}

// Respond with User Agent header value echoed in body to GET /user-agent
func handleUserAgent(req *httpRequest) *httpResponse {
	res := &httpResponse{}

	ua, err := req.Header(httpHeaderUserAgent)
	if err != nil {
		res.setStatus(httpStatusNotFound)
		return res
	}

	res.setStatus(httpStatusOK)
	res.setHeader(httpHeaderContentType, "text/plain")
	res.setHeader(httpHeaderContentLength, strconv.Itoa(len(ua)))
	res.setBody(ua)

	return res
}

// Respond with requested file served from directory to GET /files/<filename>
func handleFileRequest(req *httpRequest) *httpResponse {
	res := &httpResponse{}

	filename := req.Path()[7:]

	data, err := os.ReadFile(Config.serveDir + "/" + filename)
	if err != nil {
		log.Println("file not found:", Config.serveDir+"/"+filename)
		res.setStatus(httpStatusNotFound)

		return res
	}

	res.setStatus(httpStatusOK)
	res.setHeader(httpHeaderContentType, "application/octet-stream")
	res.setHeader(httpHeaderContentLength, strconv.Itoa(len(string(data))))
	res.setBody(string(data))

	return res
}

// Receive file and save it to directory via POST /files/<filename>
func handleFilePost(req *httpRequest) *httpResponse {
	res := &httpResponse{}

	filename := req.Path()[7:]

	//TODO: check if file already exists

	_, err := os.Stat(Config.serveDir + "/" + filename)
	if err == nil {
		log.Fatalln("file already exists, exiting for safety")
	}

	err = os.WriteFile(Config.serveDir+"/"+filename, []byte(req.body), 0666)
	if err != nil {
		log.Fatal(err)
	}

	res.setStatus(HttpStatusCreated)

	return res
}
