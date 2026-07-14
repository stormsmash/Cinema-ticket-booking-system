package config

import (
	"fmt"
	"os"
	"strconv"
)

const (
	defaultPort          = "8080"
	defaultGinMode       = "debug"
	defaultMongoURI      = "mongodb://localhost:27017/cinema?replicaSet=rs0&directConnection=true"
	defaultMongoDatabase = "cinema"
	defaultRedisAddress  = "localhost:6379"
	defaultRedisDB       = "0"
)

type Config struct {
	Port          string
	GinMode       string
	MongoURI      string
	MongoDatabase string
	RedisAddress  string
	RedisPassword string
	RedisDB       int
}

func Load() (Config, error) {
	redisDB, err := strconv.Atoi(valueOrDefault("REDIS_DB", defaultRedisDB))
	if err != nil || redisDB < 0 {
		return Config{}, fmt.Errorf("REDIS_DB must be a non-negative integer")
	}

	return Config{
		Port:          valueOrDefault("PORT", defaultPort),
		GinMode:       valueOrDefault("GIN_MODE", defaultGinMode),
		MongoURI:      valueOrDefault("MONGO_URI", defaultMongoURI),
		MongoDatabase: valueOrDefault("MONGO_DATABASE", defaultMongoDatabase),
		RedisAddress:  valueOrDefault("REDIS_ADDRESS", defaultRedisAddress),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       redisDB,
	}, nil
}

func valueOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
