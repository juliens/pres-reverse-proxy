package main

import (
	"crypto/tls"
	"fmt"
	"github.com/xenolf/lego/log"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func init() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}

func main() {
	demoURL, err := url.Parse("https://172.17.0.2")
	if err != nil {
		log.Fatal(err)
	}
	proxy := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		req.Host = demoURL.Host
		req.URL.Host = demoURL.Host
		req.URL.Scheme = demoURL.Scheme
		req.RequestURI = ""

		host, _, _ := net.SplitHostPort(req.RemoteAddr)
		req.Header.Set("X-Forwarded-For", host)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(rw, err)
			return
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
				case <-time.Tick(time.Millisecond * 10):
					rw.(http.Flusher).Flush()
				case <-done:
					return
				}
			}
		}()

		var trailerKeys []string
		for key := range resp.Trailer {
			trailerKeys = append(trailerKeys, key)
		}

		rw.Header().Set("Trailer", strings.Join(trailerKeys, ","))

		rw.WriteHeader(resp.StatusCode)
		io.Copy(rw, resp.Body)
		close(done)

		for key, values := range resp.Trailer {
			for _, value := range values {
				rw.Header().Set(key, value)
			}
		}

	})
	http.ListenAndServeTLS(":8080", "cert.pem", "key.pem", proxy)
}
