package main

import (
    "fmt"
    "log"
    "net"
    "net/http"
    "os"
    "time"

    "github.com/gorilla/mux"

    "basic/handlers"
)

func setupHandlers(r *mux.Router) {
    r.HandleFunc("/", handlers.HomeHandler).Methods("GET")
    r.HandleFunc("/home", handlers.HomeHandler).Methods("GET")
    r.HandleFunc("/content", handlers.ContentHandler).Methods("POST")
}

func main() {

    r := mux.NewRouter()
    setupHandlers(r)

    listenAddr := ":8080"

    server := &http.Server{
        Addr: listenAddr,
        Handler: r,
        IdleTimeout: 120 * time.Second,
        ReadHeaderTimeout: 2 * time.Second,
        WriteTimeout: -1,
        MaxHeaderBytes: 1 << 20, // 1 MB
    }

    listener, err := net.Listen("tcp", server.Addr); if err != nil {
        log.Printf("unable to create listener")
        os.Exit(1)
    }
    defer listener.Close()
    log.Printf("Server listening on %s\n", server.Addr)

    err = server.Serve(listener)

    if err != nil {
        if err == http.ErrServerClosed {
            log.Printf("server shutting down")
            os.Exit(0)
        } else {
            log.Printf("server error")
            os.Exit(1)
        }
    }

    err = fmt.Errorf("unexpected server shutdown")
    log.Printf("%s", err.Error())
    os.Exit(1)
}
