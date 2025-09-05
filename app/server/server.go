package server

import (
	"errors"
	"log"
	"net"
	"strings"
)

const (
	version11 = "HTTP/1.1"

	methodGet  = "GET"
	methodPost = "POST"
)

type PathHandler func(*Request) *Response

type Server struct {
	// Maximum content size of the request in bytes
	maxContentSize int
	address        string
	version        string
	pathGet        map[string]PathHandler
	pathPost       map[string]PathHandler
}

func NewServer(address string) *Server {
	return &Server{
		maxContentSize: 1024, // 1 KiB
		address:        address,
		version:        "HTTP/1.1",
		// one data structure is enough
		pathGet:  map[string]PathHandler{},
		pathPost: map[string]PathHandler{},
	}
}

func (s *Server) Run() error {
	l, err := net.Listen("tcp", s.address)
	if err != nil {
		return errors.New("failed to bind to port")
	}
	log.Println("started serving on ", s.address)

	s.acceptLoop(l)

	return nil
}

func (s *Server) acceptLoop(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("error accepting connection: ", err.Error())
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) RegisterGet(path string, handler PathHandler) {
	if _, ok := s.pathGet[path]; ok {
		log.Fatalln("invalid configuration: path already exists")
	}
	s.pathGet[path] = handler
	log.Println("path registered: GET", path)
}

func (s *Server) RegisterPost(path string, handler PathHandler) {
	if _, ok := s.pathPost[path]; ok {
		log.Fatalln("invalid configuration: path already exists")
	}
	s.pathPost[path] = handler
	log.Println("path registered: POST", path)
}

func (s *Server) handle(conn net.Conn) {
	buffer := make([]byte, 1024)

	defer conn.Close()
	_, err := conn.Read(buffer)
	if err != nil {
		log.Println("closing connection:", err.Error())
		return
	}

	// perhaps I should pass the bytes
	req, err := parseRequest(string(buffer))
	if err != nil {
		log.Println("closing connection:", err.Error())
		return
	}
	//TODO: I am not validating method nor target here, I should just accept from IP, this is misguiding
	log.Println("accepted connection:", req.Method, req.Target)

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

	getHandler, gok := s.pathGet[pathBase]
	postHandler, pok := s.pathPost[pathBase]

	if !gok && !pok {
		res.SetStatus(StatusNotFound)
	} else if (gok && req.Method != methodGet && !pok) || (!gok && pok && req.Method != methodPost) || (gok && req.Method != methodGet && pok && req.Method != methodPost) {
		// It is mandatory to set Allow header
		// https://www.rfc-editor.org/rfc/rfc9110#name-405-method-not-allowed
		res.SetStatus(StatusMethodNotAllowed)
		allowedMethods := []string{}
		if gok {
			allowedMethods = append(allowedMethods, methodGet)
		}
		if pok {
			allowedMethods = append(allowedMethods, methodPost)
		}
		res.SetHeader(HeaderAllow, strings.Join(allowedMethods, ", "))
	} else if gok && req.Method == methodGet {
		res = *getHandler(req)
	} else if pok && req.Method == methodPost {
		res = *postHandler(req)
	} else {
		log.Fatalln("server programming error")
	}

	_, err = conn.Write(res.Bytes())
	if err != nil {
		log.Println("closing connection:", err.Error())
		return
	}
	log.Printf("connection closed [%s]", res.status)
}
