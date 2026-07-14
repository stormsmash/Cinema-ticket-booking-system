package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/config"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/health"
	mongostore "github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/platform/mongodb"
	redisstore "github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/platform/redis"
	httptransport "github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/transport/http"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("run API: %v", err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	startupContext, cancelStartup := context.WithTimeout(context.Background(), 10*time.Second)
	mongoClient, err := mongostore.Connect(startupContext, cfg.MongoURI)
	if err != nil {
		cancelStartup()
		return fmt.Errorf("connect MongoDB: %w", err)
	}
	defer disconnectMongo(mongoClient)

	redisClient, err := redisstore.Connect(startupContext, redisstore.Config{
		Address:  cfg.RedisAddress,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	cancelStartup()
	if err != nil {
		return fmt.Errorf("connect Redis: %w", err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("close Redis: %v", err)
		}
	}()

	log.Printf("connected to MongoDB database %q and Redis", cfg.MongoDatabase)

	readiness := health.NewService(map[string]health.CheckFunc{
		"mongodb": func(ctx context.Context) error {
			return mongoClient.Ping(ctx, readpref.Primary())
		},
		"redis": func(ctx context.Context) error {
			return redisClient.Ping(ctx).Err()
		},
	})

	gin.SetMode(cfg.GinMode)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httptransport.NewRouter(readiness),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return serve(server, cfg.Port)
}

func serve(server *http.Server, port string) error {
	shutdownSignal, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("API listening on port %s", port)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("listen: %w", err)
	case <-shutdownSignal.Done():
		log.Print("shutting down API")
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownContext); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}

	log.Print("API stopped")
	return nil
}

func disconnectMongo(client *mongo.Client) {
	shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Disconnect(shutdownContext); err != nil {
		log.Printf("disconnect MongoDB: %v", err)
	}
}
