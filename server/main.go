package main

import (
	"net/http"
	"bytes"
	"strings"
	"os/exec"
	"log"
	"fmt"
	"io"
	"net"
	"time"
	"crypto/tls"
	"golang.org/x/net/http2"
)

// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Connection",
	"Proxy-Connection", // non-standard but still sent by libcurl and rejected by e.g. google
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",      // canonicalized version of "TE"
	"Trailer", // not Trailers per URL above; http://www.rfc-editor.org/errata_search.php?eid=4522
	"Transfer-Encoding",
	"Upgrade",
}

func main() {
	cmd := exec.Command("docker", "inspect", "-f {{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", "backend")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	server := strings.TrimSpace(out.String())
	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		//serverURL, _ := url.Parse("http://"+server)
		//revproxy := httputil.NewSingleHostReverseProxy(serverURL)
		//revproxy.ServeHTTP(rw, req)

		req.URL.Scheme = "https"
		req.URL.Host = server


		//Connection HopByHop
		for _, header := range strings.Split(req.Header.Get("Connection"), ",") {
			req.Header.Del(header)
		}

		//HopByHop
		for _, header := range hopHeaders {
			req.Header.Del(header)
		}

		//XForwardedFor
		ip, _, _ := net.SplitHostPort(req.RemoteAddr)
		xf := req.Header.Get("X-Forwarded-For")
		if xf != "" {
			xf += ","
		}
		xf += ip
		req.Header.Set("X-Forwarded-For", xf)

		//XForwardedHost XForwardedPort
		host, port, _ := net.SplitHostPort(req.Host)
		req.Header.Set("X-Forwarded-Host", host)
		req.Header.Set("X-Forwarded-Port", port)

		////TLS
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		////HTTP2
		http2.ConfigureTransport(http.DefaultTransport.(*http.Transport))

		resp, err := http.DefaultTransport.RoundTrip(req)
		if err != nil {
			rw.WriteHeader(500)
			fmt.Printf("Error on roundtrip: %s", err)
			rw.Write([]byte(fmt.Sprintf("error: %q", err)))
			return
		}

		//Connection HopByHop
		for _, header := range strings.Split(resp.Header.Get("Connection"), ",") {
			resp.Header.Del(header)
		}

		//HopByHop
		for _, header := range hopHeaders {
			resp.Header.Del(header)
		}

		//Copy resp Headers
		for key, value := range resp.Header {
			for _, val := range value {
				rw.Header().Add(key, val)
			}
		}

		//Trailer
		trailerKeys := make([]string, 0, len(resp.Trailer))
		for k := range resp.Trailer {
			trailerKeys = append(trailerKeys, k)
		}
		rw.Header().Add("Trailer", strings.Join(trailerKeys, ", "))

		//Write resp statusCode
		rw.WriteHeader(resp.StatusCode)

		//Flusher for stream
		done := make(chan bool)
		go func() {
			t := time.Tick(100 * time.Millisecond)
			for {
				select {
				case <-t:
					if rw == nil {
						return
					}
					flusher, ok := rw.(http.Flusher)
					if ok {
						flusher.Flush()
					} else {
						return
					}
				case <-done:
					return
				}

			}
		}()

		// Body copy
		io.Copy(rw, resp.Body)

		//Flush end
		done <- true

		//Trailer read after body readed
		for key, values := range resp.Trailer {
			for _, value := range values {
				rw.Header().Add(key, value)
			}
		}

	})
	//http.ListenAndServe(":8080", handler)
	http.ListenAndServeTLS(":8080", "./cert.pem", "key.pem", handler)
}
