package handlers

import (
    "context"
    "encoding/json"
    "log"
    "math"
    "math/rand"
    "net/http"
    "strconv"
    "strings"
    "time"
    "dictionary-api/internal/models"

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

func sanitizeWord(word *models.Word) {
    for i := range word.Definitions {
        word.Definitions[i].PartOfSpeech = strings.Trim(word.Definitions[i].PartOfSpeech, "\"")
        word.Definitions[i].Definition = strings.Trim(word.Definitions[i].Definition, "\"")
    }
}

func (h *Handler) SearchWords(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query()
    searchTerm := query.Get("q")
    if searchTerm == "" {
        respondWithError(w, http.StatusBadRequest, "Search query is required")
        return
    }

    page, _ := strconv.Atoi(query.Get("page"))
    if page < 1 {
        page = 1
    }
    limit, _ := strconv.Atoi(query.Get("limit"))
    if limit < 1 {
        limit = 10
    }

    filter := bson.M{
        "$or": []bson.M{
            {"word": bson.M{"$regex": searchTerm, "$options": "i"}},
            {"definitions.definition": bson.M{"$regex": searchTerm, "$options": "i"}},
        },
    }

    findOptions := options.Find().
        SetLimit(int64(limit)).
        SetSkip(int64((page - 1) * limit)).
        SetSort(bson.D{{"word", 1}})

    cursor, err := h.db.Find(context.Background(), filter, findOptions)
    if err != nil {
        log.Printf("Search error: %v", err)
        respondWithError(w, http.StatusInternalServerError, "Error searching words")
        return
    }
    defer cursor.Close(context.Background())

    var words []models.Word
    if err = cursor.All(context.Background(), &words); err != nil {
        log.Printf("Cursor error: %v", err)
        respondWithError(w, http.StatusInternalServerError, "Error processing results")
        return
    }

    for i := range words {
        sanitizeWord(&words[i])
    }

    totalCount, err := h.db.CountDocuments(context.Background(), filter)
    if err != nil {
        log.Printf("Count error: %v", err)
        respondWithError(w, http.StatusInternalServerError, "Error counting results")
        return
    }

    meta := &models.MetaData{
        TotalCount:  totalCount,
        CurrentPage: page,
        PageSize:    limit,
        TotalPages:  int(math.Ceil(float64(totalCount) / float64(limit))),
    }

    respondWithAPIResponse(w, http.StatusOK, true, words, "", meta)
}

func (h *Handler) GetRandomWord(w http.ResponseWriter, r *http.Request) {
    count, err := h.db.CountDocuments(context.Background(), bson.M{})
    if err != nil {
        log.Printf("Count error: %v", err)
        respondWithError(w, http.StatusInternalServerError, "Database error")
        return
    }

    randomSkip := rand.Int63n(count)
    var word models.Word
    err = h.db.FindOne(context.Background(), 
        bson.M{},
        options.FindOne().SetSkip(randomSkip),
    ).Decode(&word)

    if err != nil {
        log.Printf("Random word error: %v", err)
        respondWithError(w, http.StatusInternalServerError, "Error retrieving random word")
        return
    }

    sanitizeWord(&word)
    respondWithAPIResponse(w, http.StatusOK, true, word, "", nil)
}

func (h *Handler) GetWordOfDay(w http.ResponseWriter, r *http.Request) {
    key := "word_of_day"
    
    var word models.Word
    cachedWord, err := h.redis.Get(context.Background(), key).Result()
    if err == nil {
        if err := json.Unmarshal([]byte(cachedWord), &word); err == nil {
            sanitizeWord(&word)
            respondWithAPIResponse(w, http.StatusOK, true, word, "", nil)
            return
        }
    }

    count, err := h.db.CountDocuments(context.Background(), bson.M{})
    if err != nil {
        log.Printf("Count error: %v", err)
        respondWithError(w, http.StatusInternalServerError, "Database error")
        return
    }

    seed := time.Now().Format("2006-01-02")
    rand.Seed(int64(hash(seed)))
    randomSkip := rand.Int63n(count)
    
    err = h.db.FindOne(context.Background(), 
        bson.M{},
        options.FindOne().SetSkip(randomSkip),
    ).Decode(&word)

    if err != nil {
        log.Printf("Word of day error: %v", err)
        respondWithError(w, http.StatusInternalServerError, "Error retrieving word of the day")
        return
    }

    sanitizeWord(&word)
    wordJSON, _ := json.Marshal(word)
    h.redis.Set(context.Background(), key, wordJSON, 24*time.Hour)

    respondWithAPIResponse(w, http.StatusOK, true, word, "", nil)
}

func (h *Handler) GetWord(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    wordID := vars["id"]

    objID, err := primitive.ObjectIDFromHex(wordID)
    if err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid word ID")
        return
    }

    var word models.Word
    err = h.db.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&word)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            respondWithError(w, http.StatusNotFound, "Word not found")
            return
        }
        log.Printf("Get word error: %v", err)
        respondWithError(w, http.StatusInternalServerError, "Database error")
        return
    }

    sanitizeWord(&word)
    respondWithAPIResponse(w, http.StatusOK, true, word, "", nil)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
    respondWithAPIResponse(w, code, false, nil, message, nil)
}

func respondWithAPIResponse(w http.ResponseWriter, code int, success bool, data interface{}, errorMsg string, meta *models.MetaData) {
    response := models.APIResponse{
        Success: success,
        Data:    data,
        Error:   errorMsg,
        Meta:    meta,
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(response)
}

func hash(s string) uint32 {
    var h uint32
    for i := 0; i < len(s); i++ {
        h = h*31 + uint32(s[i])
    }
    return h
}