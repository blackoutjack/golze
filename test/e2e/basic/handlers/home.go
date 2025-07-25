package handlers

import (
    "log"
    "net/http"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {

    _, err := w.Write([]byte("WELCOME TO THE HOME PAGE"))
    if err != nil {
        log.Printf("failed to write response", "error", err)
    }
}

