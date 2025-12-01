package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/lupppig/briefly/db/mini"
	db "github.com/lupppig/briefly/db/postgres"
	"github.com/minio/minio-go/v7"
)

type Service struct {
	Db *db.PostgresDB
	Mc *mini.MinioClient
}

func (s *Service) GetYoutubeStatus(id string) (*db.YoutubeJob, error) {
	return s.Db.GetYoutubeJobByID(context.Background(), id)
}

func (s *Service) YoutubeService(ctx context.Context, playlistLink string) ([]*db.Youtube, error) {
	objPaths, videoIDs, err := convertPlaylistConcurrently(ctx, playlistLink, s.Mc)
	if err != nil {
		return nil, err
	}

	var results []*db.Youtube
	for i, objPath := range objPaths {
		videoID := videoIDs[i]

		yt, err := s.Db.GetOrCreateYoutube(ctx, videoID, "", playlistLink)
		if err != nil && err != pgx.ErrNoRows {
			return nil, err
		}

		if yt.AudioPath == "" {
			yt, err = s.Db.UpdateYoutubeAudioPath(ctx, videoID, objPath)
			if err != nil {
				return nil, err
			}
		}

		results = append(results, yt)
	}

	return results, nil
}

func getPlaylistVideoLinks(playlistURL string) ([]string, error) {
	cmd := exec.Command("yt-dlp", "-j", "--flat-playlist", playlistURL)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list playlist videos: %w", err)
	}

	var videoLinks []string
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		var info map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &info); err != nil {
			continue
		}
		if id, ok := info["id"].(string); ok {
			videoLinks = append(videoLinks, "https://www.youtube.com/watch?v="+id)
		}
	}
	return videoLinks, nil
}

func convertPlaylistConcurrently(ctx context.Context, playlistURL string, m *mini.MinioClient) ([]string, []string, error) {
	videoLinks, err := getPlaylistVideoLinks(playlistURL)
	if err != nil {
		return nil, nil, err
	}

	var uploaded []string
	var videoIDs []string
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 4)

	for _, link := range videoLinks {
		wg.Add(1)
		go func(l string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			videoID := extractVideoID(l)
			if videoID == "" {
				log.Println("Invalid video URL, skipping:", l)
				return
			}

			// Check if already exists in MinIO
			objectPath := "youtube/" + videoID + ".mp3"
			exists, err := minioObjectExists(ctx, m, mini.DocumentBucket, objectPath)
			if err != nil {
				log.Println("Error checking MinIO:", err)
				return
			}
			if exists {
				mu.Lock()
				uploaded = append(uploaded, objectPath)
				videoIDs = append(videoIDs, videoID)
				mu.Unlock()
				return
			}

			// Use unique temp dir per video
			tmpDir := filepath.Join("tmp", videoID)
			os.MkdirAll(tmpDir, 0755)
			defer os.RemoveAll(tmpDir)

			objPath, err := convertYoutubeVideotoAudioWithDir(ctx, l, m, tmpDir)
			if err != nil {
				log.Println("Download error:", l, err)
				return
			}

			mu.Lock()
			uploaded = append(uploaded, objPath)
			videoIDs = append(videoIDs, videoID)
			mu.Unlock()
		}(link)
	}

	wg.Wait()
	return uploaded, videoIDs, nil
}

func convertYoutubeVideotoAudioWithDir(ctx context.Context, link string, m *mini.MinioClient, tmpDir string) (string, error) {
	outputPattern := filepath.Join(tmpDir, "%(id)s.%(ext)s")
	cmd := exec.Command(
		"yt-dlp",
		"-x",
		"--audio-format", "mp3",
		"-o", outputPattern,
		"--extractor-args", "youtube:player_client=default",
		link,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("yt-dlp error: %w", err)
	}

	files, err := os.ReadDir(tmpDir)
	if err != nil || len(files) == 0 {
		return "", fmt.Errorf("failed to locate downloaded audio")
	}

	fileName := files[0].Name()
	filePath := filepath.Join(tmpDir, fileName)
	objectPath := "youtube/" + fileName

	if err := uploadToMinioIfNotExists(ctx, m, mini.DocumentBucket, objectPath, filePath); err != nil {
		return "", err
	}

	return objectPath, nil
}

func uploadToMinioIfNotExists(ctx context.Context, m *mini.MinioClient, bucket, objectPath, filePath string) error {
	exists, err := minioObjectExists(ctx, m, bucket, objectPath)
	if err != nil {
		return err
	}
	if exists {
		log.Println("Skipping upload, already exists:", objectPath)
		return nil
	}

	_, err = m.MinClient.FPutObject(
		ctx,
		bucket,
		objectPath,
		filePath,
		minio.PutObjectOptions{ContentType: "audio/mpeg"},
	)
	if err != nil {
		return fmt.Errorf("minio upload error: %w", err)
	}

	log.Println("Uploaded:", objectPath)
	return nil
}

func minioObjectExists(ctx context.Context, m *mini.MinioClient, bucket, objectPath string) (bool, error) {
	_, err := m.MinClient.StatObject(ctx, bucket, objectPath, minio.StatObjectOptions{})
	if err == nil {
		return true, nil
	}
	if minio.ToErrorResponse(err).Code == "NoSuchKey" || strings.Contains(err.Error(), "not found") {
		return false, nil
	}
	return false, err
}

func extractVideoID(link string) string {
	u, err := url.Parse(link)
	if err != nil {
		return ""
	}
	return u.Query().Get("v")
}
