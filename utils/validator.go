package utils

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

var youtubeIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{11}$`)

func ValidateYouTubeURL(link string) (string, error) {
	if link == "" {
		return "", errors.New("youtube link is required")
	}

	u, err := url.Parse(link)
	if err != nil {
		return "", errors.New("invalid url")
	}

	host := strings.ToLower(u.Host)

	// youtu.be/<id>
	if host == "youtu.be" {
		id := strings.TrimPrefix(u.Path, "/")
		if youtubeIDRegex.MatchString(id) {
			return id, nil
		}
		return "", errors.New("invalid youtube video id")
	}

	if strings.Contains(host, "youtube.com") {
		if v := u.Query().Get("v"); v != "" && youtubeIDRegex.MatchString(v) {
			return v, nil
		}

		if strings.HasPrefix(u.Path, "/shorts/") {
			id := strings.TrimPrefix(u.Path, "/shorts/")
			if youtubeIDRegex.MatchString(id) {
				return id, nil
			}
		}
		return "", errors.New("invalid youtube video id")
	}

	return "", errors.New("not a youtube link")
}

var allowedMimeTypes = map[string]bool{
	"application/pdf": true,
	"text/plain":      true,

	"audio/mpeg": true,
	"audio/wav":  true,
	"audio/mp4":  true,
}

var allowedExtensions = map[string]bool{
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".txt":  true,
	".mp3":  true,
	".wav":  true,
	".m4a":  true,
}

func ValidateUploadedFile(header *multipart.FileHeader) error {
	mime := header.Header.Get("Content-Type")
	ext := strings.ToLower(filepath.Ext(header.Filename))

	if !allowedMimeTypes[mime] {
		return fmt.Errorf("unsupported file type: %s", mime)
	}

	if !allowedExtensions[ext] {
		return fmt.Errorf("unsupported file extension: %s", ext)
	}

	return nil
}

func IsDoc(file multipart.File, filename string) bool {
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil {
		return false
	}

	if seeker, ok := file.(interface {
		Seek(int64, int) (int64, error)
	}); ok {
		seeker.Seek(0, io.SeekStart)
	}

	contentType := http.DetectContentType(buf[:n])
	if contentType == "application/pdf" || contentType == "text/plain" {
		return true
	}

	ext := filepath.Ext(filename)
	if ext == ".pdf" || ext == ".txt" {
		return true
	}

	return false
}
