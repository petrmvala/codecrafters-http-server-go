package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"os"

	"github.com/codecrafters-io/http-server-starter-go/app/server"
)

// Respond with Status OK to GET /
func handleRootResponse(req *server.Request) *server.Response {
	res := server.NewResponse()
	res.SetStatus(server.StatusOK)
	return res
}

// Respond with string echoed in body to GET /echo/<string>
func handleEchoResponse(req *server.Request) *server.Response {
	res := server.NewResponse()
	str := req.Target[6:]

	res.SetStatus(server.StatusOK)
	res.SetHeader(server.HeaderContentType, "text/plain")

	if enc, ok := req.Headers[server.HeaderAcceptEncoding]; ok {
		if _, ok := enc.(map[string]bool)["gzip"]; ok {
			var buf bytes.Buffer
			w := gzip.NewWriter(&buf)
			_, err := w.Write([]byte(str))
			if err != nil {
				log.Println("aborting compression:", err.Error())
			}
			if err := w.Close(); err != nil {
				log.Println("aborting compression:", err.Error())
			}
			res.SetHeader(server.HeaderContentLength, buf.Len())
			res.SetHeader(server.HeaderContentEncoding, "gzip")
			res.SetBody(buf.Bytes())
			return res
		}
	}
	res.SetHeader(server.HeaderContentLength, len(str))
	res.SetBody([]byte(str))
	return res
}

// Respond with User Agent header value echoed in body to GET /user-agent
func handleUserAgent(req *server.Request) *server.Response {
	res := server.NewResponse()

	ua, ok := req.Headers[server.HeaderUserAgent]
	if !ok {
		res.SetStatus(server.StatusNotFound)
		return res
	}
	usrAg := ua.(string)

	res.SetStatus(server.StatusOK)
	res.SetHeader(server.HeaderContentType, "text/plain")
	res.SetHeader(server.HeaderContentLength, len(usrAg))
	res.SetBody([]byte(usrAg))

	return res
}

// Respond with requested file served from directory to GET /files/<filename>
func (d *ServeDir) handleFileRequest() server.PathHandler {
	return func(req *server.Request) *server.Response {
		res := server.NewResponse()

		filename := req.Target[7:]

		data, err := os.ReadFile(d.directory + "/" + filename)
		if err != nil {
			log.Println("file not found:", d.directory+"/"+filename)
			res.SetStatus(server.StatusNotFound)

			return res
		}

		res.SetStatus(server.StatusOK)
		res.SetHeader(server.HeaderContentType, "application/octet-stream")
		res.SetHeader(server.HeaderContentLength, len(data))
		res.SetBody(data)

		return res
	}
}

// Receive file and save it to directory via POST /files/<filename>
func (d *ServeDir) handleFilePost() server.PathHandler {
	return func(req *server.Request) *server.Response {
		res := server.NewResponse()

		cl, ok := req.Headers[server.HeaderContentLength]
		if !ok {
			log.Println("bad request: content-length header not received")
			res.SetStatus(server.StatusLengthRequired)
			return res
		}
		cBytes := cl.(int)

		// if cBytes > Config.maxFileSizeBytes {
		// 	log.Println("error: content too large")
		// 	res.setStatus(statusContentTooLarge)
		// 	return res
		// }

		filename := req.Target[7:]
		path := d.directory + "/" + filename

		_, err := os.Stat(path)
		if err == nil {
			log.Println("error: file already exists")
			res.SetStatus(server.StatusForbiden)
			return res
		}

		file, err := os.Create(path)
		if err != nil {
			log.Println("error creating file:", err)
			res.SetStatus(server.StatusInternalServerError)
			return res
		}
		defer file.Close()

		content := req.Body[:cBytes]
		_, err = io.WriteString(file, content)
		if err != nil {
			log.Println("error writing to file:", err)
			res.SetStatus(server.StatusInternalServerError)
			return res
		}

		log.Println(cBytes, " bytes written to ", path)
		res.SetStatus(server.StatusCreated)

		return res
	}
}
