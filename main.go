package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/theAnuragMishra/myserver/server"
)

func main() {
	godotenv.Load()

	pool, err := connectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	router := server.NewRouter(
		server.IPTrackerModule,
		server.TweetModule(pool),
	)
	portString := os.Getenv("PORT")

	if portString == "" {
		portString = "8080"
	}

	srv := &http.Server{
		Handler: router,
		Addr:    ":" + portString,
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func connectDB() (*pgxpool.Pool, error) {
	ctx := context.Background()

	connStr := os.Getenv("PG_URL")
	if connStr == "" {
		return nil, fmt.Errorf("postgres url required")
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
