package middleware

import (
    "context"
    "log"
    "net/http"
    "strconv"
    "time"
    "strings"
    "runtime/debug"

    "github.com/go-redis/redis/v8"
)

func JSONContentType(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        next.ServeHTTP(w, r)
    })
}

func RateLimit(client *redis.Client) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ip := r.RemoteAddr
            if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
                ip = strings.Split(forwardedFor, ",")[0]
            }

            key := "ratelimit:" + ip

            ctx := context.Background()
            count, err := client.Incr(ctx, key).Result()
            if err != nil {
                http.Error(w, "Rate limit error", http.StatusInternalServerError)
                return
            }

            if count == 1 {
                client.Expire(ctx, key, time.Minute)
            }

            w.Header().Set("X-RateLimit-Limit", "100")
            w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(100-count, 10))

            if count > 100 {
                w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Minute).Unix(), 10))
                http.Error(w, "Rate limit exceeded. Try again in 1 minute.", http.StatusTooManyRequests)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

func Recovery(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("panic: %v\n%s", err, debug.Stack())
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}

func CORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}