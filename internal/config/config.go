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
    err := godotenv.Load()
    if err != nil {
        log.Println("Warning: .env file not found")
    }

    return &Config{
        Port:          getEnv("PORT", "8080"),
        MongoURI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
        MongoDatabase: getEnv("MONGODB_DATABASE", "dictionary"),
        RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
        RedisUsername: getEnv("REDIS_USERNAME", "default"),
        RedisPassword: getEnv("REDIS_PASSWORD", ""),
        RedisDB:       getEnvAsInt("REDIS_DB", 0),
    }
}

func getEnv(key, defaultValue string) string {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    return value
}

func getEnvAsInt(key string, defaultValue int) int {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    
    intValue, err := strconv.Atoi(value)
    if err != nil {
        return defaultValue
    }
    return intValue
}

func NewRedisClient() *redis.Client {
    client := redis.NewClient(&redis.Options{
        Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
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