// internal/config/config.go

package config

import (
    "context"
    "log"
    "os"
    "strconv"

    "github.com/go-redis/redis/v8"
    "github.com/joho/godotenv"
)

type Config struct {
    Port          string
    MongoURI      string
    MongoDatabase string
    RedisAddr     string
    RedisUsername string
    RedisPassword string
    RedisDB       int
}

func LoadConfig() *Config {
    _ = godotenv.Load()

    config := &Config{
        Port:          getEnv("PORT", "8080"),
        MongoURI:      getEnv("MONGODB_URI", ""),
        MongoDatabase: getEnv("MONGODB_DATABASE", "dictionary"),
        RedisAddr:     getEnv("REDIS_ADDR", ""),
        RedisUsername: getEnv("REDIS_USERNAME", "default"),
        RedisPassword: getEnv("REDIS_PASSWORD", ""),
        RedisDB:       getEnvAsInt("REDIS_DB", 0),
    }

    if config.MongoURI == "" {
        log.Fatal("MONGODB_URI is required")
    }
    if config.RedisAddr == "" {
        log.Fatal("REDIS_ADDR is required")
    }

    return config
}

func NewRedisClient() *redis.Client {
    client := redis.NewClient(&redis.Options{
        Addr:     getEnv("REDIS_ADDR", ""),
        Username: getEnv("REDIS_USERNAME", "default"),
        Password: getEnv("REDIS_PASSWORD", ""),
        DB:       getEnvAsInt("REDIS_DB", 0),
    })

    ctx := context.Background()
    _, err := client.Ping(ctx).Result()
    if err != nil {
        log.Fatalf("Failed to connect to Redis: %v", err)
    }

    log.Println("Successfully connected to Redis")
    return client
}

func getEnv(key, defaultValue string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
    if value, exists := os.LookupEnv(key); exists {
        if intVal, err := strconv.Atoi(value); err == nil {
            return intVal
        }
    }
    return defaultValue
}