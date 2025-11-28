package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/lupppig/briefly/db"
	"github.com/lupppig/briefly/service"
	"github.com/lupppig/briefly/utils"
)

type BriefHandler struct {
	Db *db.PostgresDB
}

func (b BriefHandler) PostYoutube(w http.ResponseWriter, r *http.Request) {
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
