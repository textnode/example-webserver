package main

import (
    "fmt"
    "html"
    "io"
    "log"
    "net/http"
    "time"
)

/*
   Generate self-signed cert for SSL server...

   openssl genrsa -des3 -out server.key 1024
   openssl rsa -in server.key -out server.key.insecure
   openssl req -new -key server.key -out server.csr
   openssl x509 -req -days 3650 -in server.csr -signkey server.key -out server.crt
*/


type HandlerObject struct {
    value uint64
}

func NewHandlerObject(value uint64) *HandlerObject {
    return &HandlerObject{value: value}
}

func (self *HandlerObject) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "HandlerObject, %q", html.EscapeString(r.URL.Path))
}



type nopCloser struct {
    io.Reader
}

func (nopCloser) Close() error { return nil }

func main() {

    var handlerInstance *HandlerObject = NewHandlerObject(123)

    var httpMux *http.ServeMux = http.NewServeMux()
    httpServer := &http.Server{
        Addr:           ":8080",
        Handler:        httpMux,
        ReadTimeout:    2 * time.Second,
        WriteTimeout:   2 * time.Second,
        MaxHeaderBytes: 1 << 20,
    }

    var httpsMux *http.ServeMux = http.NewServeMux()
    httpsServer := &http.Server{
        Addr:           ":8443",
        Handler:        httpsMux,
        ReadTimeout:    10 * time.Second,
        WriteTimeout:   10 * time.Second,
        MaxHeaderBytes: 1 << 20,
    }

    httpMux.Handle("/foo", handlerInstance)
    httpMux.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
        log.Printf("%+v", r)
        r.Body = http.MaxBytesReader(w, nopCloser{r.Body}, 1024)
        fmt.Fprintf(w, "HandlerFunc, %q", html.EscapeString(r.URL.Path))
    })

    httpsMux.Handle("/foo", handlerInstance)
    httpsMux.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
        log.Printf("%+v", r)
        r.Body = http.MaxBytesReader(w, nopCloser{r.Body}, 65536)
        fmt.Fprintf(w, "HandlerFunc, %q", html.EscapeString(r.URL.Path))
    })

    go func() {
        log.Println("Before starting HTTPS listener...")
        err := httpsServer.ListenAndServeTLS("server.crt", "server.key.insecure")
        if err != nil {
            log.Fatal("HTTPS listener couldn't start")
        }
    }()

    log.Println("Before starting HTTP listener...")
    err := httpServer.ListenAndServe()
    if err != nil {
        log.Fatal("HTTP listener couldn't start")
    }
}
