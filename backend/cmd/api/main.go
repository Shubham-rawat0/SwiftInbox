package main

import (
	"fmt"
	"log"
	"os"

	internal "github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/api"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/api/router"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/database"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/repository/postgres"
	"github.com/joho/godotenv"
)

func main(){
	err:=godotenv.Load()
	if err!=nil{
		panic(err)
	}
	databaseURL:=os.Getenv("DB_URL")

	db, err := database.NewPostgres(databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	queries:=postgres.New(db)

	port:=os.Getenv("PORT")

	handler:=router.NewServerMux(queries)
	server:=internal.NewApiServer(port,handler)
	
	fmt.Println("server started at port",port)
	err=server.ListenAndServe()
	if err!=nil{
		panic(err)
	}
}
