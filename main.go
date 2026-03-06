package main

import (
	"log"
	"net/http"
	"os"

	"github.com/theAnuragMishra/myserver/server"
)

func main() {
	router := server.NewRouter()
	portString := os.Getenv("PORT")

	srv := &http.Server{
		Handler: router,
		Addr:    ":" + portString,
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
