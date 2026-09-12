package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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
	cleanupEnabled:=os.Getenv("CLEANUP_ENABLED")

	db, err := database.NewPostgres(databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	queries := postgres.New(db)

	scheduler:=cleanup.NewScheduler(queries)

	go func() {
		fmt.Println("starting smtp server")
		if err := smtp.NewSMTPServer(queries).Start(); err != nil {
			log.Printf("SMTP server stopped: %v", err)
		}
	}()

	port := os.Getenv("PORT")

	handler := router.NewServerMux(queries)
	server := internal.NewApiServer(port, handler)

	fmt.Println("server started at port", port)

	signChan:=make(chan os.Signal ,1)
	signal.Notify(signChan,syscall.SIGINT, syscall.SIGTERM)

	go func(){
		if cleanupEnabled=="1"{
			log.Println("starting cleanup scheduler")
			scheduler.Start()
		}
	}()

	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}

	sig:=<-signChan
	log.Println("shutting down server",sig)
	scheduler.Stop()
}	
