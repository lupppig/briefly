package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/lupppig/briefly/db"
	"github.com/lupppig/briefly/handlers"
)

func main() {
	godotenv.Load()

	port := os.Getenv("PORT")
	dbUrl := os.Getenv("DB_URL")
	r := mux.NewRouter()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := db.ConnectPostgres(ctx, dbUrl)
	if err != nil {
		log.Println(err.Error())
		return
	}

	h := handlers.BriefHandler{Db: db}

	r.HandleFunc("/brief-youtube", h.PostYoutube)

	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf(":%s", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Server port started on: %v", port)
	log.Fatal(srv.ListenAndServe())
}
