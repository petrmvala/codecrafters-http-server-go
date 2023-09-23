package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
)

const (
	responseOK       = "HTTP/1.1 200 OK\r\n\r\n"
	responseNotFound = "HTTP/1.1 404 Not Found\r\n\r\n"
)

func connHandler(conn net.Conn) {

	defer conn.Close()

	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading connection: ", err.Error())
		os.Exit(1)
	}

	response := responseNotFound
	req := string(buffer)

	r, err := regexp.MatchString("GET / HTTP/1.1.*", req)
	if err != nil {
		fmt.Println("Error matching regex: ", err.Error())
	}

	if r {
		response = responseOK
	}

	_, err = conn.Write([]byte(response))
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
