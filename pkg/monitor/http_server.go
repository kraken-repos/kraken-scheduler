package monitor

import (
	"fmt"
	"net/http"
)

func httpHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello eureka\n")
}

func StartHTTPServer() {
	go func() {
		http.HandleFunc("/", httpHandler)
		http.ListenAndServe(":443", nil)
	}()
}