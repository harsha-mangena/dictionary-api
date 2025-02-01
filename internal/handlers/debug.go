package handlers

import (
    "context"
    "encoding/json"
    "log"
    "net/http"

    "go.mongodb.org/mongo-driver/bson"
)

func (h *Handler) DebugWord(w http.ResponseWriter, r *http.Request) {
    var rawDoc bson.M
    err := h.db.FindOne(context.Background(), bson.M{}).Decode(&rawDoc)
    if err != nil {
        log.Printf("Error fetching document: %v", err)
        respondWithError(w, http.StatusInternalServerError, "Database error")
        return
    }

    log.Printf("Document structure: %+v", rawDoc)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(rawDoc)
}