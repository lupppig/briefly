package utils

import (
	"errors"
	"net/url"
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

	// youtube.com/watch?v=<id>
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
