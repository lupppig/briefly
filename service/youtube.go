package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/lupppig/briefly/db/mini"
	"github.com/lupppig/briefly/utils"
	"github.com/minio/minio-go/v7"
)

func (s *Service) ProcessYoutubeJob(jobID, link string) {
	update := func(status string, summary interface{}, errMsg string) {
		s.JobManager.UpdateJob(jobID, status, summary, errMsg)
	}

	defer func() {
		if r := recover(); r != nil {
			update("error", "", fmt.Sprintf("panic: %v", r))
		}
	}()

	update("validating_url", "", "")
	videoID, err := utils.ValidateYouTubeURL(link)
	if err != nil {
		update("error", "", err.Error())
		return
	}

	ctx := context.Background()
	update("checking_cache", "", "")

	yt, err := s.Db.GetOrCreateYoutube(ctx, videoID, "", link)
	if err != nil {
		update("error", "", "failed db fetch")
		return
	}

	saved, _ := s.Db.GetContentByYoutubeID(ctx, yt.ID)
	if saved != nil {
		update("cached_summary_found", saved, "")
		return
	}

	audioPath := yt.AudioPath
	if audioPath != "" {
		exists, _ := s.Mc.ObjectExists(mini.DocumentBucket, audioPath)
		if exists {
			update("cached_audio_found", "", "")
			goto TRANSCRIBE
		}
	}

	update("downloading_audio", "", "")
	audioPath, err = s.ExtractAudioToMinio(ctx, link, mini.DocumentBucket)
	if err != nil {
		update("error", "", "failed to extract audio")
		return
	}
	s.Db.UpdateYoutubeAudioPath(ctx, videoID, audioPath)

TRANSCRIBE:
	update("transcribing", "", "")
	content, err := s.TranscribeAudio(audioPath)
	if err != nil {
		update("error", "", "transcription failed")
		return
	}

	update("summarizing", "", "")
	summary, err := s.AiGenResponse(ctx, content, "youtube")
	if err != nil {
		update("error", "", "summarize failed")
		return
	}

	update("saving", "", "")
	sumCon, err := s.Db.CreateContent(ctx, content, summary, nil, &yt.ID)
	if err != nil {
		update("error", "", "failed to save content")
		return
	}

	update("done", sumCon, "")
}

func (s *Service) ExtractAudioToMinio(ctx context.Context, link string, bucket string) (string, error) {
	timestamp := time.Now().UnixNano()
	videoPath := filepath.Join(os.TempDir(), fmt.Sprintf("yt_%d.mp4", timestamp))
	audioPath := filepath.Join(os.TempDir(), fmt.Sprintf("yt_%d.wav", timestamp))

	downloadCmd := exec.Command(
		"yt-dlp",
		"-f", "bestaudio",
		"-o", videoPath,
		link,
	)

	if out, err := downloadCmd.CombinedOutput(); err != nil {
		os.Remove(videoPath)
		return "", fmt.Errorf("yt-dlp failed: %v\noutput: %s", err, out)
	}

	extractCmd := exec.Command(
		"ffmpeg",
		"-i", videoPath,
		"-vn",
		"-acodec", "pcm_s16le",
		"-ar", "16000",
		"-ac", "1",
		audioPath,
	)

	if out, err := extractCmd.CombinedOutput(); err != nil {
		os.Remove(videoPath)
		os.Remove(audioPath)
		return "", fmt.Errorf("ffmpeg extract audio failed: %v\noutput: %s", err, out)
	}

	os.Remove(videoPath)

	file, err := os.Open(audioPath)
	if err != nil {
		os.Remove(audioPath)
		return "", fmt.Errorf("failed to open extracted audio: %w", err)
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	objectPath := fmt.Sprintf("youtube_audio/%d.wav", timestamp)

	_, err = s.Mc.MinClient.PutObject(
		ctx,
		bucket,
		objectPath,
		file,
		fileInfo.Size(),
		minio.PutObjectOptions{
			ContentType: "audio/wav",
		},
	)
	if err != nil {
		os.Remove(audioPath)
		return "", fmt.Errorf("failed to upload to minio: %w", err)
	}

	os.Remove(audioPath)

	return objectPath, nil
}
