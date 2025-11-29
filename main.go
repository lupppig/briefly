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
	"github.com/lupppig/briefly/db/mini"
	db "github.com/lupppig/briefly/db/postgres"
	"github.com/lupppig/briefly/handlers"
)

func main() {
	godotenv.Load()

	port := os.Getenv("PORT")
	dbUrl := os.Getenv("DB_URL")
	endPoint := os.Getenv("END_POINT")
	accessPoint := os.Getenv("ACCESS_POINT")
	secretAccesskey := os.Getenv("SECRET_ACCESSKEY")
	useSSL := os.Getenv("USE_SSL") != "0"

	r := mux.NewRouter()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// db connection
	db, err := db.ConnectPostgres(ctx, dbUrl)
	if err != nil {
		log.Println(err.Error())
		return
	}

	// minio connection
	mc, err := mini.MinioConnect(endPoint, accessPoint, secretAccesskey, useSSL)
	if err != nil {
		log.Printf("failed to connect to minio %v", err)
		return
	}
	h := handlers.BriefHandler{Db: db, Mclient: mc}

	r.HandleFunc("/brief-youtube", h.PostYoutube)
	r.HandleFunc("/brief-doc", h.PostAudioDoc)

	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf(":%s", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Server port started on: %v", port)
	log.Fatal(srv.ListenAndServe())
}
