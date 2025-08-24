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
	paths    map[string]map[string]pathHandler
	pathGet  map[string]pathHandler
	pathPost map[string]pathHandler
}

func newDistributor() *distributor {
	paths := make(map[string]map[string]pathHandler)
	return &distributor{
		paths:    paths,
		pathGet:  map[string]pathHandler{},
		pathPost: map[string]pathHandler{},
	}
}

func (d *distributor) registerPath(path, method string, handler pathHandler) {
	if _, pathExists := d.paths[path]; !pathExists {
		d.paths[path] = map[string]pathHandler{
			method: handler,
		}
	} else if _, methodExists := d.paths[path][method]; !methodExists {
		d.paths[path][method] = handler
	} else {
		log.Fatalln("method", method, "already registered for path", path)
	}
	log.Println("registered method", method, "for path", path)
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

	matched := false

	// if req.method == methodGet {
	// 	for path, handler := range d.pathGet {
	// 		globMatch := false
	// 		if string(path[len(path)-1]) == "*" {
	// 			globMatch = true
	// 		}
	// 		if !globMatch && path == req.target { // path /foo, perform exact match
	// 			matched = true
	// 		} else if globMatch && len(path) <= len(req.target) && strings.HasPrefix(req.target, path[:len(path)-1]) { // path in form /foo*, perform prefix match
	// 			matched = true
	// 		} else {
	// 			continue
	// 		}
	// 		res = *handler(req)
	// 		break
	// 	}
	// } else if req.method == methodPost {

	// }

	// The server only supports one level of nesting
	// If it ends without a slash, path needs to be matched exactly
	// If it ends with a slash, path can extend past slash
	pathBase := req.target
	log.Println("pathBase", pathBase)
	sl := strings.Index(req.target[1:], "/")
	if sl != -1 {
		pathBase = req.target[:sl+2]
		log.Println("second slash found, pathBase", pathBase)
	}
	// now is pathBase == / or /echo -> exact match, OR /echo/ -> prefix match (which is also exact match)

	for path, methods := range d.paths {
		if path != pathBase {
			log.Println("path", path, "doesn't match pathBase", pathBase)
			continue
		}
		matched = true

		handler, ok := methods[req.method]
		if ok {
			res = *handler(req)
			break
		}

		// It is mandatory to set Allow header
		// https://www.rfc-editor.org/rfc/rfc9110#name-405-method-not-allowed
		keys := make([]string, 0, len(methods))
		for m := range methods {
			keys = append(keys, m)
		}
		allowedMethods := strings.Join(keys, ", ")

		res.setStatus(statusMethodNotAllowed)
		res.setHeader(statusMethodNotAllowed, allowedMethods)

		break
	}

	if !matched {
		res.setStatus(statusNotFound)
	}

	_, err = conn.Write([]byte(res.ToString()))
	if err != nil {
		log.Fatalln("error writing to connection: ", err.Error())
	}
	log.Printf("server[%s]", res.status)
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
	d.registerPath("/", methodGet, handleRootResponse)
	d.registerPath("/user-agent", methodGet, handleUserAgent)
	d.registerPath("/echo/", methodGet, handleEchoResponse)
	d.registerPath("/files/", methodGet, handleFileRequest)
	d.registerPath("/files/", methodPost, handleFilePost)
	// d.get("/", handleRootResponse)
	// d.get("/user-agent", handleUserAgent)
	// d.get("/echo/*", handleEchoResponse)
	// d.get("/files/*", handleFileRequest)
	// d.post("/files/*", handleFilePost)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("error accepting connection: ", err.Error())
			continue
		}
		go d.handle(conn)
	}
}
