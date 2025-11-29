package utils

import (
	"mime/multipart"

	"github.com/faiface/beep/mp3"
	"github.com/unidoc/unipdf/v3/model"
)

func GettMP3Duration(fi multipart.File) (float64, error) {
	streamer, format, err := mp3.Decode(fi)
	if err != nil {
		return 0, err
	}
	defer streamer.Close()

	duration := float64(streamer.Len()) / float64(format.SampleRate)
	return duration, nil
}

func GetPDFPageCount(fi multipart.File) (int, error) {

	pdfReader, err := model.NewPdfReader(fi)
	if err != nil {
		return 0, err
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return 0, err
	}

	return numPages, nil
}
