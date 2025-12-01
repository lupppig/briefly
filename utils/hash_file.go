package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"mime/multipart"

	"github.com/google/uuid"
)

func HashFile(file multipart.File) (string, error) {
	hasher := sha256.New()

	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	hash := hex.EncodeToString(hasher.Sum(nil))

	if seeker, ok := file.(interface {
		Seek(int64, int) (int64, error)
	}); ok {
		_, err := seeker.Seek(0, io.SeekStart)
		if err != nil {
			return "", err
		}
	}

	return hash, nil
}

func NewJobID() string {
	return uuid.NewString()
}
