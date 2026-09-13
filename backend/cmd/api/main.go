package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/cmd/worker"
	internal "github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/api"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/api/router"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/cleanup"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/database"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/repository/postgres"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/smtp"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	databaseURL := os.Getenv("DB_URL")
	cleanupEnabled := os.Getenv("CLEANUP_ENABLED")

	db, err := database.NewPostgres(databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	queries := postgres.New(db)

	rabbit, err := worker.StartWorker()
	if err != nil {
		panic(err)
	}
	defer rabbit.Close()

	consumerCtx, stopConsumer := context.WithCancel(context.Background())
	defer stopConsumer()
	
	go func() {
		if err := rabbit.Consume(consumerCtx); err != nil && consumerCtx.Err() == nil {
			log.Printf("[WEBHOOK CONSUMER] stopped unexpectedly: %v", err)
		}
	}()

	scheduler := cleanup.NewScheduler(queries, rabbit)

	go func() {
		fmt.Println("starting smtp server")
		if err := smtp.NewSMTPServer(queries, rabbit).Start(); err != nil {
			log.Printf("SMTP server stopped: %v", err)
		}
	}()

	port := os.Getenv("PORT")

	handler := router.NewServerMux(queries)
	server := internal.NewApiServer(port, handler)

	fmt.Println("server started at port", port)

	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if cleanupEnabled == "1" {
			log.Println("[SCHEDULER] starting cleanup scheduler")
			scheduler.Start()
		}
	}()

	serverErr := make(chan error, 1)

	go func() {
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err = <-serverErr:
		if err != http.ErrServerClosed {
			log.Fatal(err)
		}

	case sig := <-signChan:
		log.Println("shutting down server", sig)
		stopConsumer()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("server shutdown error: %v", err)
		}
	}

	scheduler.Stop()
}
