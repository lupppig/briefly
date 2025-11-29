package service

import (
	"context"
	"log"
	"mime/multipart"
	"path/filepath"

	"github.com/lupppig/briefly/db/mini"
	db "github.com/lupppig/briefly/db/postgres"
	"github.com/lupppig/briefly/utils"
)

func (s *Service) AudioDocService(fi multipart.File, fh *multipart.FileHeader) (*db.DocumentAudio, error) {
	var objKey string

	err := utils.ValidateUploadedFile(fh)
	if err != nil {
		return nil, err
	}

	hashedFile, err := utils.HashFile(fi)
	if err != nil {
		return nil, err
	}

	if utils.IsDoc(fi, fh.Filename) {
		objKey = filepath.Join("uploads", "doc", hashedFile)
	} else {
		objKey = filepath.Join("uploads", "audio", hashedFile)
	}

	objExist, err := s.Mc.ObjectExists(mini.DocumentBucket, objKey)

	if err != nil {
		log.Printf("failed to locate object file: %v", err.Error())
		return nil, err
	}

	var respDoc *db.DocumentAudio
	if !objExist {
		var durationInSec *float64
		var pageCount *int

		if utils.IsDoc(fi, fh.Filename) {
			count, err := utils.GetPDFPageCount(fi)
			if err != nil {
				log.Printf("could not get page count: %v", err)
				return nil, err
			}

			pageCount = &count
		} else {
			dur, err := utils.GettMP3Duration(fi)
			if err != nil {
				log.Printf("could not get audio upload duration: %v", err)
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

		respDoc, err = s.Db.GetOrCreateDocument(context.Background(), doc)
		if err != nil {
			log.Println(err.Error())
			return nil, err
		}

		_, err = s.Mc.AddToBucket(context.Background(), fi, fh, mini.DocumentBucket, objKey)
		if err != nil {
			log.Printf("failed to add object to bucket: %v", err)
			return nil, err
		}

	}

	// perform extraction logic right here...

	return respDoc, nil
}
