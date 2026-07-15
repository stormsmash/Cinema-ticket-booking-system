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
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	adminservice "github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/admin"
	auditservice "github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/audit"
	authservice "github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/auth"
	bookingservice "github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/booking"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/config"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/domain"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/health"
	mongostore "github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/platform/mongodb"
	redisstore "github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/platform/redis"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/realtime"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/screening"
	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/seatlock"
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
	defer cancelStartup()
	mongoClient, err := mongostore.Connect(startupContext, cfg.MongoURI)
	if err != nil {
		return fmt.Errorf("connect MongoDB: %w", err)
	}
	defer disconnectMongo(mongoClient)

	redisClient, err := redisstore.Connect(startupContext, redisstore.Config{
		Address:  cfg.RedisAddress,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	if err != nil {
		return fmt.Errorf("connect Redis: %w", err)
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("close Redis: %v", err)
		}
	}()

	database := mongoClient.Database(cfg.MongoDatabase)
	if err := mongostore.Bootstrap(startupContext, database, cfg.AdminEmails); err != nil {
		return fmt.Errorf("bootstrap MongoDB: %w", err)
	}
	cancelStartup()

	log.Printf("connected to MongoDB database %q and Redis", cfg.MongoDatabase)

	screeningRepository := screening.NewMongoRepository(
		database.Collection(mongostore.CollectionScreenings),
	)
	screeningService := screening.NewService(screeningRepository)
	userRepository := authservice.NewMongoUserRepository(
		database.Collection(mongostore.CollectionUsers),
		cfg.AdminEmails,
	)
	sessionStore := authservice.NewRedisSessionStore(redisClient)
	googleProvider := authservice.NewGoogleProvider(
		cfg.GoogleClientID,
		cfg.GoogleClientSecret,
		cfg.GoogleRedirectURL,
	)
	authService := authservice.NewService(
		googleProvider,
		userRepository,
		sessionStore,
		cfg.SessionTTL,
		cfg.GoogleOAuthEnabled(),
	)
	seatLockStore := seatlock.NewRedisStore(redisClient)
	auditRepository := auditservice.NewMongoRepository(
		database.Collection(mongostore.CollectionAuditLogs),
	)
	seatLockService := seatlock.NewService(
		screeningService,
		seatLockStore,
		auditRepository,
		cfg.SeatLockTTL,
	)
	seatEventPublisher := realtime.NewRedisPublisher(redisClient)
	bookingRepository := bookingservice.NewMongoRepository(
		mongoClient,
		database.Collection(mongostore.CollectionScreenings),
		database.Collection(mongostore.CollectionBookings),
		database.Collection(mongostore.CollectionAuditLogs),
	)
	bookingService := bookingservice.NewService(
		bookingRepository,
		screeningService,
		seatLockStore,
		seatEventPublisher,
	)
	adminRepository := adminservice.NewMongoRepository(
		database.Collection(mongostore.CollectionBookings),
		database.Collection(mongostore.CollectionUsers),
		database.Collection(mongostore.CollectionScreenings),
		database.Collection(mongostore.CollectionAuditLogs),
	)
	eventHub := realtime.NewHub(200)
	eventSource := realtime.NewRedisSeatEventSource(redisClient, cfg.RedisDB)
	realtimeContext, stopRealtime := context.WithCancel(context.Background())
	defer stopRealtime()
	defer eventHub.Close()
	go func() {
		if err := eventSource.Run(realtimeContext, func(event realtime.SeatEvent) {
			eventHub.Publish(event)
			if event.Type != realtime.SeatExpired {
				return
			}
			screeningID, err := bson.ObjectIDFromHex(event.ScreeningID)
			if err != nil {
				return
			}
			auditContext, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := auditRepository.Create(auditContext, domain.AuditLog{
				ID:          bson.NewObjectID(),
				Event:       domain.AuditEventBookingTimeout,
				ScreeningID: screeningID,
				SeatID:      event.SeatID,
				CreatedAt:   event.OccurredAt,
			}); err != nil {
				log.Printf("record expired seat audit: %v", err)
			}
		}); err != nil {
			log.Printf("seat event source stopped: %v", err)
		}
	}()

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
		Addr: ":" + cfg.Port,
		Handler: httptransport.NewRouter(httptransport.Dependencies{
			Readiness:   readiness,
			Screenings:  screeningService,
			Auth:        authService,
			SeatLocks:   seatLockService,
			SeatEvents:  eventHub,
			FrontendURL: cfg.FrontendURL,
			Bookings:    bookingService,
			Admin:       adminRepository,
			AuthConfig: httptransport.AuthHandlerConfig{
				FrontendURL:  cfg.FrontendURL,
				SessionTTL:   cfg.SessionTTL,
				CookieSecure: cfg.CookieSecure,
			},
		}),
		ReadHeaderTimeout: 5 * time.Second,
	}
	server.RegisterOnShutdown(func() {
		stopRealtime()
		eventHub.Close()
	})

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
