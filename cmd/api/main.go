package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    "crypto/tls"

    "dictionary-api/internal/config"
    "dictionary-api/internal/handlers"
    "dictionary-api/internal/middleware"

    "github.com/gorilla/mux"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
    cfg := config.LoadConfig()

    mongoClient, err := connectMongoDB(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer mongoClient.Disconnect(context.Background())

    redisClient := config.NewRedisClient()
    defer redisClient.Close()

    wordCollection := mongoClient.Database(cfg.MongoDatabase).Collection("words")
    handler := handlers.NewHandler(wordCollection, redisClient)

    router := mux.NewRouter()

    apiRouter := router.PathPrefix("/api/v1").Subrouter()
    apiRouter.Use(middleware.JSONContentType)
    apiRouter.Use(middleware.RateLimit(redisClient))


    apiRouter.HandleFunc("/words/random", handler.GetRandomWord).Methods("GET")
    apiRouter.HandleFunc("/words/word-of-day", handler.GetWordOfDay).Methods("GET")
    apiRouter.HandleFunc("/words/search", handler.SearchWords).Methods("GET")
    apiRouter.HandleFunc("/debug", handler.DebugWord).Methods("GET")

    apiRouter.HandleFunc("/words/{id}", handler.GetWord).Methods("GET")

    srv := &http.Server{
        Handler:      router,
        Addr:         ":" + cfg.Port,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }


    go func() {
        log.Printf("Starting server on port %s", cfg.Port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()

    log.Printf("Server is running on http://localhost:%s", cfg.Port)
    log.Printf("Available endpoints:")
    log.Printf("- GET /api/v1/words/search?q=<query>")
    log.Printf("- GET /api/v1/words/random")
    log.Printf("- GET /api/v1/words/word-of-day")
    log.Printf("- GET /api/v1/words/{id}")
    


    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    <-c

    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }

    log.Println("Server gracefully stopped")
}

func connectMongoDB(cfg *config.Config) (*mongo.Client, error) {
    log.Printf("Attempting to connect to MongoDB at URI: %s", cfg.MongoURI)
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    clientOptions := options.Client().
        ApplyURI(cfg.MongoURI).
        SetConnectTimeout(30*time.Second).
        SetServerSelectionTimeout(30*time.Second).
        SetMaxPoolSize(50).
        SetTLSConfig(&tls.Config{
            InsecureSkipVerify: false,
            MinVersion:         tls.VersionTLS12,
        }).
        SetRetryWrites(true).
        SetRetryReads(true).
        SetDirect(false)

    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        log.Printf("Failed to create MongoDB client: %v", err)
        return nil, err
    }

    pingCtx, pingCancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer pingCancel()
    
    err = client.Ping(pingCtx, nil)
    if err != nil {
        log.Printf("Failed to ping MongoDB: %v", err)
        return nil, err
    }

    log.Println("Successfully connected to MongoDB")
    return client, nil
}