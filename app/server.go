package main

import (
	"flag"
	"log"
	"net"
	"strings"
)

const (
	version11 = "HTTP/1.1"

	methodGet  = "GET"
	methodPost = "POST"
)

type pathHandler func(*request) *response

type distributor struct {
	pathGet  map[string]pathHandler
	pathPost map[string]pathHandler
}

func newDistributor() *distributor {
	return &distributor{
		pathGet:  map[string]pathHandler{},
		pathPost: map[string]pathHandler{},
	}
}

func (d *distributor) get(path string, handler pathHandler) {
	if _, ok := d.pathGet[path]; ok {
		log.Fatalln("invalid configuration: path already exists")
	}
	d.pathGet[path] = handler
	log.Println("path registered: GET", path)
}

func (d *distributor) post(path string, handler pathHandler) {
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
	log.Println("accepted connection:", req.method, req.target)

	//TODO: I don't like this
	res := response{
		headers: headers{},
	}

	// The server only supports one level of nesting
	// If it ends without a slash, path needs to be matched exactly
	// If it ends with a slash, path can extend past slash
	pathBase := req.target
	sl := strings.Index(req.target[1:], "/")
	if sl != -1 {
		pathBase = req.target[:sl+2]
	}
	// now is pathBase == / or /echo -> exact match, OR /echo/ -> prefix match (which is also exact match)

	getHandler, gok := d.pathGet[pathBase]
	postHandler, pok := d.pathPost[pathBase]

	if !gok && !pok {
		res.setStatus(statusNotFound)
	} else if (gok && req.method != methodGet && !pok) || (!gok && pok && req.method != methodPost) || (gok && req.method != methodGet && pok && req.method != methodPost) {
		// It is mandatory to set Allow header
		// https://www.rfc-editor.org/rfc/rfc9110#name-405-method-not-allowed
		res.setStatus(statusMethodNotAllowed)
		allowedMethods := []string{}
		if gok {
			allowedMethods = append(allowedMethods, methodGet)
		}
		if pok {
			allowedMethods = append(allowedMethods, methodPost)
		}
		res.setHeader(headerAllow, strings.Join(allowedMethods, ", "))
	} else if gok && req.method == methodGet {
		res = *getHandler(req)
	} else if pok && req.method == methodPost {
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

var Config config

type config struct {
	serveDir         string
	maxFileSizeBytes int
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	serveDir := flag.String("directory", "/tmp/data/codecrafters.io/http-server-tester/", "Directory to serve files from")
	maxFileSizeBytes := flag.Int("max-file-size", 1000000, "Max accepted file size in Bytes [1MB]")
	flag.Parse()

	Config = config{
		serveDir:         *serveDir,
		maxFileSizeBytes: *maxFileSizeBytes,
	}

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		log.Fatalln("failed to bind to port 4221")
	}
	log.Println("started serving on port 4221")

	d := newDistributor()
	d.get("/", handleRootResponse)
	d.get("/user-agent", handleUserAgent)
	d.get("/echo/", handleEchoResponse)
	d.get("/files/", handleFileRequest)
	d.post("/files/", handleFilePost)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("error accepting connection: ", err.Error())
			continue
		}
		go d.handle(conn)
	}
}
