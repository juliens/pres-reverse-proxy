package main

import (
	"net/http"
	"fmt"
	"io"
)

func main() {
	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		server := "172.17.0.2"
		req.URL.Host = server
		req.URL.Scheme = "http"
		resp, err := http.DefaultTransport.(*http.Transport).RoundTrip(req)
		if err != nil {
			rw.WriteHeader(500)
			fmt.Println(err)
		}
		io.Copy(rw, resp.Body)
	})
	http.ListenAndServe(":8080", handler)
}
