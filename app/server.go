package main

import (
	"log"
	"net"
	"strings"
)

type pathHandler func(*request) *response

// The distributor maps paths and methods to their handlers
//
//	{
//	  "/":     {"GET": getRoothandler}
//	  "/path": {
//			"GET": getPathhandler,
//			"PUT": putPathhandler,
//	  }
//	  "/foo/*": {"GET": getFooHandler}
//	}
type distributor struct {
	paths map[string]map[string]pathHandler
}

func newDistributor() *distributor {
	var d distributor
	d.paths = make(map[string]map[string]pathHandler)

	return &d
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

func (d *distributor) handle(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		log.Fatalln("error reading connection: ", err.Error())
	}

	req, err := parseRequest(string(buffer))
	if err != nil {
		log.Fatalln("error parsing request: ", err.Error())
	}
	log.Printf("client[%s]: %s", req.Method(), req.Path())

	res := response{
		version: version_11,
		headers: headers{},
	}
	matched := false

	for path, methods := range d.paths {
		matchAll := false
		if string(path[len(path)-1]) == "*" {
			matchAll = true
		}

		// path in form /foo, perform exact match
		if !matchAll && path == req.Path() {
			matched = true

			// path in form /foo*, perform prefix match
		} else if matchAll && len(path) <= len(req.Path()) &&
			strings.HasPrefix(req.Path(), path[:len(path)-1]) {
			matched = true
		}

		// match method
		if matched {
			handler, ok := methods[req.Method()]
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
	}

	if !matched {
		res.setStatus(statusNotFound)
	}

	_, err = conn.Write([]byte(res.ToString()))
	if err != nil {
		log.Fatalln("error writing to connection: ", err.Error())
	}
	log.Printf("server[%s]", res.Status())
}

func main() {

	configure()

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		log.Fatalln("failed to bind to port 4221")
	}
	log.Println("server started serving on port 4221")

	d := newDistributor()
	d.registerPath("/", methodGet, handleRootResponse)
	d.registerPath("/user-agent", methodGet, handleUserAgent)
	d.registerPath("/echo/*", methodGet, handleEchoResponse)
	d.registerPath("/files/*", methodGet, handleFileRequest)
	d.registerPath("/files/*", methodPost, handleFilePost)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("error accepting connection: ", err.Error())
			continue
		}
		go d.handle(conn)
	}
}
