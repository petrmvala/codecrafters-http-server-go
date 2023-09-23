package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	httpMethodGet  = "GET"
	httpMethodPost = "POST"

	httpVersion_11 = "HTTP/1.1"

	httpStatusOK       = "200"
	httpStatusNotFound = "404"

	httpStatusTextOK       = "OK"
	httpStatusTextNotFound = "Not Found"

	httpHeaderContentType   = "Content-Type"
	httpHeaderContentLength = "Content-Length"
)

type httpMethod string

type httpVersion string

type httpStatusCode string

type httpStatusText string

type httpHeaderKey string

type httpRequest struct {
	startLine startLine
	headers   []httpHeader
	body      string
}

type httpResponse struct {
	statusLine statusLine
	headers    []httpHeader
	body       string
}

type startLine struct {
	httpMethod    httpMethod
	requestTarget string
	httpVersion   httpVersion
}

type statusLine struct {
	httpVersion    httpVersion
	httpStatusCode httpStatusCode
	httpStatusText httpStatusText
}

type httpHeader struct {
	name  httpHeaderKey
	value string
}

func validateHttpMethod(method string) (httpMethod, error) {
	switch method {
	case httpMethodGet, httpMethodPost:
		return httpMethod(method), nil
	default:
		return "", errors.New("invalid HTTP method")
	}
}

func validateHttpTarget(target string) bool {
	if m, _ := regexp.MatchString(`/(echo/\w+)?`, target); m {
		return true
	}
	return false
}

func validateHttpVersion(version string) (httpVersion, error) {
	if version == httpVersion_11 {
		return httpVersion(httpVersion_11), nil
	}
	return "", errors.New("invalid HTTP protocol version")
}

func parseRequest(r string) (*httpRequest, error) {
	lines := strings.Split(r, "\r\n")

	firstLine := strings.Split(lines[0], " ")
	method, target, protocol := firstLine[0], firstLine[1], firstLine[2]

	m, err := validateHttpMethod(method)
	if err != nil {
		return nil, err
	}

	t := validateHttpTarget(target)
	if !t {
		fmt.Println("The http target is invalid")
	}

	v, err := validateHttpVersion(protocol)
	if err != nil {
		return nil, err
	}

	return &httpRequest{
		startLine: startLine{
			httpMethod:    m,
			requestTarget: target,
			httpVersion:   v,
		},
		headers: nil,
		body:    "",
	}, nil

	// omitting other lines for now
}

func composeRequest(request *httpRequest) string {
	m := fmt.Sprintf("%s %s %s\r\n\r\n",
		request.startLine.httpMethod,
		request.startLine.requestTarget,
		request.startLine.httpVersion,
	)
	return m
}

func composeResponse(response *httpResponse) string {
	statusLine := fmt.Sprintf("%s %s %s\r\n",
		response.statusLine.httpVersion,
		response.statusLine.httpStatusCode,
		response.statusLine.httpStatusText,
	)

	headers := ""
	for _, v := range response.headers {
		headers += fmt.Sprintf("%s: %s\r\n", v.name, v.value)
	}

	return fmt.Sprintf("%s%s\r\n%s", statusLine, headers, response.body)
}

func handleEchoResponse(req *httpRequest) *httpResponse {
	return &httpResponse{
		statusLine: statusLine{
			httpVersion:    httpVersion(httpVersion_11),
			httpStatusCode: httpStatusCode(httpStatusOK),
			httpStatusText: httpStatusText(httpStatusTextOK),
		},
		headers: []httpHeader{
			{
				name:  httpHeaderKey(httpHeaderContentType),
				value: "text/plain",
			},
			{
				name:  httpHeaderKey(httpHeaderContentLength),
				value: strconv.Itoa(len(req.startLine.requestTarget[6:])),
			},
		},
		body: req.startLine.requestTarget[6:],
	}
}

func handleRootResponse() *httpResponse {
	return &httpResponse{
		statusLine: statusLine{
			httpVersion:    httpVersion(httpVersion_11),
			httpStatusCode: httpStatusCode(httpStatusOK),
			httpStatusText: httpStatusText(httpStatusTextOK),
		},
		headers: nil,
		body:    "",
	}
}

func handleDefaultResponse() *httpResponse {
	return &httpResponse{
		statusLine: statusLine{
			httpVersion:    httpVersion(httpVersion_11),
			httpStatusCode: httpStatusCode(httpStatusNotFound),
			httpStatusText: httpStatusText(httpStatusTextNotFound),
		},
		headers: nil,
		body:    "",
	}
}

func connHandler(conn net.Conn) {

	defer conn.Close()

	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading connection: ", err.Error())
		os.Exit(1)
	}

	req, err := parseRequest(string(buffer))
	if err != nil {
		fmt.Println("Error parsing request: ", err.Error())
		os.Exit(1)
	}
	fmt.Println("Received request:")
	fmt.Print(composeRequest(req))

	var res *httpResponse
	switch {
	case req.startLine.requestTarget == "/":
		res = handleRootResponse()
	case strings.HasPrefix(req.startLine.requestTarget, "/echo/"):
		res = handleEchoResponse(req)
	default:
		res = handleDefaultResponse()
	}

	out := composeResponse(res)

	_, err = conn.Write([]byte(out))
	if err != nil {
		fmt.Println("Error writing to connection: ", err.Error())
		os.Exit(1)
	}

}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		connHandler(conn)
	}
}
