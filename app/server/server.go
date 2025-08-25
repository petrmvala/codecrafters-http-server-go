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

type Server struct {
	maxFilesizeBytes int
	port             string
	serveDir         string
	version          string
	Distributor      distributor
}

func NewServer(port string) *Server {
	return &Server{
		maxFilesizeBytes: 1000000, // 1 MB
		port:             port,
		serveDir:         "/tmp/data/codecrafters.io/http-server-tester/",
		version:          "HTTP/1.1",
		Distributor:      distributor{},
	}
}

func (s *Server) Run() error {
	l, err := net.Listen("tcp", "0.0.0.0:"+s.port)
	if err != nil {
		return errors.New("failed to bind to port")
	}
	log.Println("started serving on port", s.port)

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
		go s.Distributor.handle(conn)
	}
}

type PathHandler func(*Request) *Response

type distributor struct {
	pathGet  map[string]PathHandler
	pathPost map[string]PathHandler
}

func NewDistributor() *distributor {
	return &distributor{
		pathGet:  map[string]PathHandler{},
		pathPost: map[string]PathHandler{},
	}
}

func (d *distributor) Get(path string, handler PathHandler) {
	if _, ok := d.pathGet[path]; ok {
		log.Fatalln("invalid configuration: path already exists")
	}
	d.pathGet[path] = handler
	log.Println("path registered: GET", path)
}

func (d *distributor) Post(path string, handler PathHandler) {
	if _, ok := d.pathPost[path]; ok {
		log.Fatalln("invalid configuration: path already exists")
	}
	d.pathPost[path] = handler
	log.Println("path registered: POST", path)
}

func (d *distributor) handle(conn net.Conn) {
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

	getHandler, gok := d.pathGet[pathBase]
	postHandler, pok := d.pathPost[pathBase]

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
