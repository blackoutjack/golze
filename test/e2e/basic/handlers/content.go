
package handlers

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
)

type ContentRequest struct {
    RecordKind string `json:"kind"`
    RecordData []string `json:"records"`
}

func ContentHandler(w http.ResponseWriter, r *http.Request) {

    reqContent := r.Body

    data := &ContentRequest{}
    err := json.NewDecoder(reqContent).Decode(data); if err != nil {
        fmt.Printf("failed to parse request body as json: %w", err)
    }

    resp := fmt.Sprintf("YOU SENT %d %s RECORDS, THANKS!",
        len(data.RecordData), data.RecordKind,)

    _, err = w.Write([]byte(resp))
    if err != nil {
        log.Printf("failed to write response", "error", err)
    }
}
