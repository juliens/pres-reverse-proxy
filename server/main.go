package main

import (
	"fmt"
	"golang.org/x/net/http2"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func main() {
	proxyURL, err := url.Parse("https://172.17.0.2")
	if err != nil {
		log.Fatal(err)
	}

	proxy := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		req.Host = proxyURL.Host
		req.URL.Scheme = proxyURL.Scheme
		req.URL.Host = proxyURL.Host
		req.RequestURI = ""

		s, _, _ := net.SplitHostPort(req.RemoteAddr)
		req.Header.Set("X-Forwarded-For", s)

		for _, value := range strings.Split(
			req.Header.Get("Connection"), ",",
		) {
			req.Header.Del(value)
		}

		for _, value := range hopHeaders {
			req.Header.Del(value)
		}

		http2.ConfigureTransport(
			http.DefaultTransport.(*http.Transport))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(rw, err)
			return
		}

		for _, value := range strings.Split(
			resp.Header.Get("Connection"), ",",
		) {
			resp.Header.Del(value)
		}

		for _, value := range hopHeaders {
			resp.Header.Del(value)
		}

		for key, values := range resp.Header {
			for _, value := range values {
				rw.Header().Set(key, value)
			}
		}

		done := make(chan bool)
		go func() {
			for {
				select {
				case <-time.Tick(10 * time.Millisecond):
					rw.(http.Flusher).Flush()
				case <-done:
					return
				}
			}
		}()

		trailerKeys := []string{}
		for key := range resp.Trailer {
			trailerKeys = append(trailerKeys, key)
		}

		rw.Header().Set("Trailer",
			strings.Join(trailerKeys, ","),
			)

		rw.WriteHeader(resp.StatusCode)
		io.Copy(rw, resp.Body)

		for key, values := range resp.Trailer {
			for _, value := range values {
				rw.Header().Set(key, value)
			}
		}

		close(done)

	})
	// http.ListenAndServe(":8080", proxy)
	http.ListenAndServeTLS(":8080","cert.pem","key.pem", proxy)
}
