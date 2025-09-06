package server

import (
	"errors"
	"log"
	"net"
	"strings"
)

const (
	version11 = "HTTP/1.1"
)

type PathHandler func(*Request) *Response

type Server struct {
	// Maximum content size of the request in bytes
	maxContentSize int
	address        string
	version        string
	paths          map[string]map[string]PathHandler
}

func NewServer(address string) *Server {
	return &Server{
		maxContentSize: 1024, // 1 KiB
		address:        address,
		version:        "HTTP/1.1",
		paths:          map[string]map[string]PathHandler{},
	}
}

func (s *Server) Run() error {
	l, err := net.Listen("tcp", s.address)
	if err != nil {
		return errors.New("failed to bind to port")
	}
	log.Println("started serving on", s.address)

	s.acceptLoop(l)

	return nil
}

func (s *Server) acceptLoop(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("error accepting connection:", err.Error())
			continue
		}
		log.Println("accepted connection:", conn.RemoteAddr().String())
		go s.handle(conn)
	}
}

func (s *Server) Register(path string, method string, handler PathHandler) {
	m, ok := s.paths[path]
	if !ok {
		s.paths[path] = map[string]PathHandler{method: handler}
		log.Println("path registered:", method, path)
		return
	}

	if _, ok := m[method]; !ok {
		m[method] = handler
		s.paths[path] = m
		log.Println("path registered:", method, path)
		return
	}

	log.Fatalln("path and method already exists:", method, path)
}

func (s *Server) handle(conn net.Conn) {
	buffer := make([]byte, 1024)

	defer conn.Close()
	_, err := conn.Read(buffer)
	if err != nil {
		log.Println("closing connection:", conn.RemoteAddr().String(), err.Error())
		return
	}

	writeAndClose := func(res Response) {
		_, err = conn.Write(res.Bytes())
		if err != nil {
			log.Println("error closing connection:", conn.RemoteAddr().String(), err.Error())
			return
		}
		conn.Close()
		log.Printf("connection closed [%s]: %s", res.status, conn.RemoteAddr().String())
	}

	req, err := parseRequest(buffer)
	if err != nil {
		log.Println("closing connection:", conn.RemoteAddr().String(), err.Error())
		return
	}

	//TODO: I don't like this
	res := Response{
		headers: headers{},
	}

	// The server only supports one level of nesting
	// If it ends without a slash, path needs to be matched exactly
	// If it ends with a slash, path can extend past slash
	pathBase := req.Target
	sl := strings.Index(req.Target[1:], "/")
	if sl != -1 {
		pathBase = req.Target[:sl+2]
	}
	// now is pathBase == / or /echo -> exact match, OR /echo/ -> prefix match (which is also exact match)

	methods, ok := s.paths[pathBase]
	if !ok {
		res.SetStatus(StatusNotFound)
		writeAndClose(res)
	}

	handler, ok := methods[req.Method]
	if !ok {
		// It is mandatory to set Allow header
		// https://www.rfc-editor.org/rfc/rfc9110#name-405-method-not-allowed
		res.SetStatus(StatusMethodNotAllowed)
		for m, _ := range methods {
			res.AddHeader(HeaderAllow, m)
		}
		writeAndClose(res)
	}

	res = *handler(req)
	writeAndClose(res)
}
