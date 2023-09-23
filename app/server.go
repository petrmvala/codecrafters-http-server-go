package main

import (
	"net/http"
)

func rootHandler(w http.ResponseWriter, req *http.Request) {
	if req.RequestURI == "/" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func main() {
	http.HandleFunc("/", rootHandler)
	http.ListenAndServe(":4221", nil)
}
