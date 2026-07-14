package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/config"
	httptransport "github.com/stormsmash/Cinema-ticket-booking-system/backend/internal/transport/http"
)

func main() {
	cfg := config.Load()
	gin.SetMode(cfg.GinMode)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httptransport.NewRouter(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	shutdownSignal, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("API listening on port %s", cfg.Port)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("start server: %v", err)
		}
		return
	case <-shutdownSignal.Done():
		log.Print("shutting down API")
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownContext); err != nil {
		log.Fatalf("shutdown server: %v", err)
	}

	log.Print("API stopped")
}
