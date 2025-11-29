package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

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
		log.Println(err.Error())
		utils.FerrorResponse(w, http.StatusBadRequest, "bad response", err.Error())
		return
	}

	videoID, err := utils.ValidateYouTubeURL(req.Link)
	if err != nil {
		utils.FerrorResponse(w, http.StatusBadRequest, "bad response", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	serv := service.Service{Db: b.Db}

	err = serv.YoututbeService(ctx, req.Link, videoID)

	if err != nil {
		log.Println(err.Error())
		utils.InternalServerResponse(w)
		return
	}
	utils.JSONResponse(w, http.StatusOK, "youtube link seen", videoID)
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
