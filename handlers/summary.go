package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/lupppig/briefly/db/mini"
	db "github.com/lupppig/briefly/db/postgres"
	"github.com/lupppig/briefly/service"
	"github.com/lupppig/briefly/utils"
)

type BriefHandler struct {
	Db      *db.PostgresDB
	Mclient *mini.MinioClient
}

func (b *BriefHandler) PostYoutube(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Link string `json:"link"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.FerrorResponse(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if _, err := utils.ValidateYouTubeURL(req.Link); err != nil {
		utils.FerrorResponse(w, http.StatusBadRequest, "invalid YouTube link", err.Error())
		return
	}

	jobID := utils.NewJobID()

	go processYoutubeJob(jobID, req.Link, b.Db, b.Mclient)
	utils.JSONResponse(w, http.StatusAccepted, "youtube processing queued", map[string]string{"job_id": jobID})

}

func (b *BriefHandler) GetYoutubeJobStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID, err := uuid.Parse(vars["job_id"])
	if err != nil {
		utils.FerrorResponse(w, http.StatusBadRequest, "invalid job id", err.Error())
		return
	}

	serv := service.Service{Db: b.Db, Mc: b.Mclient}
	job, err := serv.GetYoutubeStatus(jobID.String())
	if err != nil {
		utils.FerrorResponse(w, http.StatusNotFound, "job not found", err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, "job status", job)
}

func (b *BriefHandler) PostAudioDoc(w http.ResponseWriter, r *http.Request) {

	const MaxUploadSize int64 = 20 << 20
	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)

	if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
		utils.FerrorResponse(w, http.StatusBadRequest, "file too large", err.Error())
		return
	}
	fi, fh, err := r.FormFile("upload-file")
	if err != nil {
		log.Println("could not get file", err.Error())
		utils.InternalServerResponse(w)
		return
	}

	serv := service.Service{Db: b.Db, Mc: b.Mclient}
	doc, err := serv.AudioDocService(fi, fh)

	if err != nil {
		log.Printf("could not upload file: %v", err)
		if strings.Contains(err.Error(), "unsupported") {
			utils.FerrorResponse(w, http.StatusBadRequest, err.Error(), "")
			return
		}
		utils.InternalServerResponse(w)
		return
	}
	utils.JSONResponse(w, http.StatusOK, "upload successful", doc)

}
