package handlers

import (
    "context"
    "encoding/json"
    "log"
    "math/rand"
    "net/http"
    "time"

    "github.com/go-redis/redis/v8"
    "github.com/gorilla/mux"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type Handler struct {
    db    *mongo.Collection
    redis *redis.Client
}

func NewHandler(db *mongo.Collection, redisClient *redis.Client) *Handler {
    return &Handler{
        db:    db,
        redis: redisClient,
    }
}

func (h *Handler) SearchWords(w http.ResponseWriter, r *http.Request) {
    log.Printf("Handling SearchWords request")
    query := r.URL.Query().Get("q")
    if query == "" {
        respondWithError(w, http.StatusBadRequest, "Search query is required")
        return
    }

    filter := bson.M{
        "word": bson.M{
            "$regex":   query,
            "$options": "i",
        },
    }

    cursor, err := h.db.Find(context.Background(), filter, options.Find().SetLimit(10))
    if err != nil {
        log.Printf("Error searching words: %v", err)
        respondWithError(w, http.StatusInternalServerError, "Error searching words")
        return
    }
    defer cursor.Close(context.Background())

    var results []interface{}
    if err = cursor.All(context.Background(), &results); err != nil {
        log.Printf("Error processing results: %v", err)
        respondWithError(w, http.StatusInternalServerError, "Error processing results")
        return
    }

    log.Printf("Found %d results for query: %s", len(results), query)
    respondWithJSON(w, http.StatusOK, results)
}

func (h *Handler) GetRandomWord(w http.ResponseWriter, r *http.Request) {
    log.Printf("Handling GetRandomWord request")
    
    count, err := h.db.CountDocuments(context.Background(), bson.M{})
    if err != nil {
        log.Printf("Error counting documents: %v", err)
        respondWithError(w, http.StatusInternalServerError, "Database error")
        return
    }

    randomSkip := rand.Int63n(count)
    var word interface{}
    err = h.db.FindOne(context.Background(), 
        bson.M{},
        options.FindOne().SetSkip(randomSkip),
    ).Decode(&word)

    if err != nil {
        log.Printf("Error retrieving random word: %v", err)
        respondWithError(w, http.StatusInternalServerError, "Error retrieving random word")
        return
    }

    log.Printf("Successfully retrieved random word")
    respondWithJSON(w, http.StatusOK, word)
}

func (h *Handler) GetWordOfDay(w http.ResponseWriter, r *http.Request) {
    log.Printf("Handling GetWordOfDay request")
    key := "word_of_day"
    
    cachedWord, err := h.redis.Get(context.Background(), key).Result()
    if err == nil {
        log.Printf("Returning cached word of the day")
        w.Write([]byte(cachedWord))
        return
    }

    count, err := h.db.CountDocuments(context.Background(), bson.M{})
    if err != nil {
        log.Printf("Error counting documents: %v", err)
        respondWithError(w, http.StatusInternalServerError, "Database error")
        return
    }

    randomSkip := rand.Int63n(count)
    var word interface{}
    err = h.db.FindOne(context.Background(), 
        bson.M{},
        options.FindOne().SetSkip(randomSkip),
    ).Decode(&word)

    if err != nil {
        log.Printf("Error retrieving word of day: %v", err)
        respondWithError(w, http.StatusInternalServerError, "Error retrieving word of the day")
        return
    }

    wordJSON, _ := json.Marshal(word)
    h.redis.Set(context.Background(), key, wordJSON, 24*time.Hour)

    log.Printf("Successfully retrieved and cached new word of the day")
    respondWithJSON(w, http.StatusOK, word)
}

func (h *Handler) GetWord(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    wordID := vars["id"]
    log.Printf("Handling GetWord request for ID: %s", wordID)

    objID, err := primitive.ObjectIDFromHex(wordID)
    if err != nil {
        log.Printf("Invalid word ID format: %s", wordID)
        respondWithError(w, http.StatusBadRequest, "Invalid word ID")
        return
    }

    var word interface{}
    err = h.db.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&word)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            log.Printf("Word not found for ID: %s", wordID)
            respondWithError(w, http.StatusNotFound, "Word not found")
            return
        }
        log.Printf("Database error: %v", err)
        respondWithError(w, http.StatusInternalServerError, "Database error")
        return
    }

    log.Printf("Successfully retrieved word with ID: %s", wordID)
    respondWithJSON(w, http.StatusOK, word)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
    respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
    response, err := json.Marshal(payload)
    if err != nil {
        log.Printf("Error marshaling JSON: %v", err)
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    w.Write(response)
}