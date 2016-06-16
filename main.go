package main

import (
	"log"
	"net/http"
	"strconv"
)

type handlerError struct {
	Error   error
	Message string
	Code    int
}

type httpHandler func(http.ResponseWriter, *http.Request) *handlerError

func (fn httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := fn(w, r); e != nil {
		http.Error(w, e.Message, e.Code)
	}
}

func helloworld(rw http.ResponseWriter, request *http.Request) *handlerError {
	rw.Write([]byte("Hello world!"))
	return nil
}

func main() {
	http.Handle("/helloworld", httpHandler(helloworld))

	PORT := 3000
	log.Printf("Listening on :%d", PORT)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(PORT), nil))
}
