package handlers

import (
	"context"
	"fmt"
	"log"

	"github.com/lupppig/briefly/db/mini"
	db "github.com/lupppig/briefly/db/postgres"
	"github.com/lupppig/briefly/service"
)

func processYoutubeJob(jobID string, link string, db *db.PostgresDB, m *mini.MinioClient) {
	err := db.InsertJobStatus(jobID, "processing", link)

	if err != nil {
		log.Println("failed to insert into youtube job status:", err.Error())
		return
	}

	serv := service.Service{Db: db, Mc: m}
	results, err := serv.YoutubeService(context.Background(), link)
	if err != nil {
		db.UpdateJobStatus(jobID, "failed", err.Error())
		return
	}

	storagePath := fmt.Sprintf("uploads/youtube/transcripts")
	for _, yt := range results {
		serv.TranscribeAudio(yt.AudioPath, storagePath)
	}
	// db.UpdateJobResult(jobID, summary)y
}
