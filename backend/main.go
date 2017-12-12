package main

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

func main() {
	mux := http.NewServeMux()

	mux.Handle("/trailer", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Trailer", "X-Trailer")
		rw.WriteHeader(http.StatusOK)
		rw.(http.Flusher).Flush()
		rw.Write([]byte("HELLO"))
		rw.Header().Set("X-Trailer", "X-Value")
		rw.(http.Flusher).Flush()
		rw.Write([]byte("HELLO"))

	}))

	mux.Handle("/stream", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("Begin"))
		rw.(http.Flusher).Flush()
		time.Sleep(3 * time.Second)
		rw.Write([]byte("End"))
		rw.(http.Flusher).Flush()
	}))

	mux.Handle("/demo.json", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("{}"))
	}))

	mux.Handle("/httpversion", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(req.Proto))
	}))

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Connection", "keep-alive")
		u, _ := url.Parse(req.URL.String())
		queryParams := u.Query()
		wait := queryParams.Get("wait")
		if len(wait) > 0 {
			duration, err := time.ParseDuration(wait)
			if err == nil {
				time.Sleep(duration)
			}
		}
		hostname, _ := os.Hostname()
		fmt.Fprintln(w, "Hostname:", hostname)
		ifaces, _ := net.Interfaces()
		for _, i := range ifaces {
			addrs, _ := i.Addrs()
			// handle err
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				fmt.Fprintln(w, "IP:", ip)
			}
		}
		req.Write(w)
		host, _, err := net.SplitHostPort(req.RemoteAddr)
		if err == nil {
			fmt.Fprintf(w, "Remote addr: %s", host)
		}
	}))

	log := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		scheme := "http"
		if req.TLS != nil {
			scheme = "https"
		}
		fmt.Printf("%s://%s%s\n", scheme, req.Host, req.RequestURI)
		mux.ServeHTTP(rw, req)
	})
	go http.ListenAndServe(":80", log)
	fmt.Println(http.ListenAndServeTLS(":443", "./cert.pem", "./key.pem", log))

}
