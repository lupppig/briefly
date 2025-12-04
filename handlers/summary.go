package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/lupppig/briefly/db/mini"
	db "github.com/lupppig/briefly/db/postgres"
	"github.com/lupppig/briefly/service"
	"github.com/lupppig/briefly/utils"
)

type BriefHandler struct {
	Db      *db.PostgresDB
	Mclient *mini.MinioClient
	Serv    *service.Service
}

func (b *BriefHandler) PostYoutube(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Link string `json:"link"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.FerrorResponse(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	jobID := utils.NewJobID()
	b.Serv.JobManager.CreateJob(jobID)

	go b.Serv.ProcessYoutubeJob(jobID, req.Link)

	utils.JSONResponse(w, http.StatusOK, "ok", map[string]string{"job_id": jobID})
}

func (b *BriefHandler) GetYoutubeJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["job_id"]

	job := b.Serv.JobManager.GetJob(jobID)
	if job == nil {
		utils.FerrorResponse(w, http.StatusNotFound, "job not found", "")
		return
	}

	utils.JSONResponse(w, http.StatusOK, "ok", job)
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
