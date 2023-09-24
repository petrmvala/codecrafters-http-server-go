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
	httpHeaderUserAgent     = "User-Agent"
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

func parseHeader(header string) (*httpHeader, error) {
	key, value, found := strings.Cut(header, ": ")
	if !found {
		return nil, errors.New("invalid header")
	}
	switch key {
	case httpHeaderContentLength, httpHeaderContentType, httpHeaderUserAgent:
		return &httpHeader{name: httpHeaderKey(key), value: value}, nil
	default:
		return nil, errors.New("invalid header")
	}
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

	// If the string cannot be split, body is set to ""
	header, body, _ := strings.Cut(r, "\r\n\r\n")

	headlines := strings.Split(header, "\r\n")

	// parse headers
	var headers []httpHeader
	if len(headlines) > 1 {
		for _, h := range headlines[1:] {
			header, err := parseHeader(h)
			if err != nil {
				// ignore invalid headers, maybe we should act differently
				continue
			}
			headers = append(headers, *header)
		}
	}

	firstLine := strings.Split(headlines[0], " ")
	if len(firstLine) != 3 {
		return nil, errors.New("malformed request")
	}
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
		headers: headers,
		body:    body,
	}, nil
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

func handleUserAgent(req *httpRequest) *httpResponse {
	var ua string
	for _, v := range req.headers {
		if v.name == httpHeaderUserAgent {
			ua = v.value
			break
		}
	}

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
				value: strconv.Itoa(len(ua)),
			},
		},
		body: ua,
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
	case req.startLine.requestTarget == "/user-agent":
		res = handleUserAgent(req)
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
