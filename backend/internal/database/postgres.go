package database

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewPostgres(url string) (*sql.DB, error) {
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, err
	}

	fmt.Println("connected to db")

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	fmt.Println("db connected")

	return db, nil
}