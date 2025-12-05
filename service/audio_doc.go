package service

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	"github.com/lupppig/briefly/db/mini"
	db "github.com/lupppig/briefly/db/postgres"
	"github.com/lupppig/briefly/utils"
)

type Service struct {
	Db           *db.PostgresDB
	Mc           *mini.MinioClient
	WhisperModel whisper.Model
	JobManager   *JobManager
}

var modelsPath string = "models/ggml-base.en.bin"

func NewService(db *db.PostgresDB, m *mini.MinioClient) (*Service, error) {
	model, err := whisper.New(modelsPath)
	if err != nil {
		return nil, err
	}
	return &Service{Db: db, Mc: m, WhisperModel: model, JobManager: NewJobManager()}, nil
}

func (s *Service) AudioDocService(fi multipart.File, fh *multipart.FileHeader) (*db.SummaryContent, error) {
	var objKey string

	if err := utils.ValidateUploadedFile(fh); err != nil {
		return nil, err
	}

	hashedFile, err := utils.HashFile(fi)
	if err != nil {
		return nil, err
	}

	fi.Seek(0, io.SeekStart)

	if utils.IsDoc(fi, fh.Filename) {
		objKey = filepath.Join("uploads", "doc", hashedFile)
	} else {
		objKey = filepath.Join("uploads", "audio", hashedFile)
	}

	objExist, err := s.Mc.ObjectExists(mini.DocumentBucket, objKey)
	if err != nil {
		log.Printf("failed to check object in MinIO: %v", err)
		return nil, err
	}

	if !objExist {
		fi.Seek(0, io.SeekStart)
		_, err := s.Mc.AddToBucket(context.Background(), fi, fh, mini.DocumentBucket, objKey)
		if err != nil {
			log.Printf("failed to upload object to MinIO: %v", err)
			return nil, err
		}
	}

	var durationInSec *float64
	var pageCount *int
	fi.Seek(0, io.SeekStart)
	if utils.IsDoc(fi, fh.Filename) {
		count, err := utils.GetPDFPageCount(fi)
		if err != nil {
			log.Printf("could not get PDF page count: %v", err)
			return nil, err
		}
		pageCount = &count
	} else {
		dur, err := utils.GettMP3Duration(fi)
		if err != nil {
			log.Printf("could not get audio duration: %v", err)
			return nil, err
		}
		durationInSec = &dur
	}

	doc := db.DocumentAudio{
		Name:            fh.Filename,
		FileType:        fh.Header.Get("Content-Type"),
		StoragePath:     objKey,
		Size:            fh.Size,
		FileHash:        hashedFile,
		DurationSeconds: durationInSec,
		PageCount:       pageCount,
	}

	respDoc, err := s.Db.GetOrCreateDocument(context.Background(), doc)
	if err != nil {
		log.Printf("failed to get or create document in DB: %v", err)
		return nil, err
	}

	existingSummary, err := s.Db.GetContentByDocID(context.Background(), respDoc.ID)
	if err == nil && existingSummary != nil {
		return existingSummary, nil
	}

	fi.Seek(0, io.SeekStart)
	var content string
	if utils.IsDoc(fi, fh.Filename) {
		ext := strings.ToLower(filepath.Ext(fh.Filename))
		switch ext {
		case ".pdf":
			content, err = s.ExtractPdfDoc(objKey)
			if err != nil {
				return nil, fmt.Errorf("failed to extract PDF content: %w", err)
			}
		case ".txt":
			content, err = s.GetTXTContent(objKey)
			if err != nil {
				return nil, fmt.Errorf("failed to extract TXT content: %w", err)
			}
		}
	}

	summaryText, err := s.AiGenResponse(context.Background(), content, "pdf document")
	if err != nil {
		return nil, err
	}

	sums, err := s.Db.CreateContent(context.Background(), content, summaryText, &respDoc.ID, nil)
	if err != nil {
		return nil, err
	}

	return sums, nil
}

func (s *Service) ExtractPdfDoc(objKey string) (string, error) {
	buf, err := s.Mc.GetObjectBuffer(mini.DocumentBucket, objKey)
	if err != nil {
		return "", fmt.Errorf("failed to get PDF from MinIO: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "pdf-*.pdf")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(buf.Bytes()); err != nil {
		return "", fmt.Errorf("failed to write PDF to temp file: %w", err)
	}

	cmd := exec.Command("pdftotext", tmpFile.Name(), "-")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to extract text: %w", err)
	}

	return string(out), nil
}

func (s *Service) GetTXTContent(objKey string) (string, error) {
	buf, err := s.Mc.GetObjectBuffer(mini.DocumentBucket, objKey)
	if err != nil {
		return "", err
	}

	b := buf.Bytes()
	scanner := bufio.NewScanner(bytes.NewReader(b))

	const maxCapacity = 20 * 1024 * 1024 // 20mb
	scanner.Buffer(make([]byte, 1024*1024), maxCapacity)

	var doc strings.Builder

	doc.Grow(len(b))

	for scanner.Scan() {
		line := scanner.Text()
		doc.WriteString(line)
		doc.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return doc.String(), nil
}
