package service

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	"github.com/hajimehoshi/go-mp3"
	"github.com/lupppig/briefly/db/mini"
)

func (s *Service) TranscribeAudio(audioPath string, localPath string) error {
	buf, err := s.Mc.GetObjectBuffer(mini.DocumentBucket, audioPath)
	if err != nil {
		return err
	}

	modelsPath := "models/ggml-base.en.bin"
	model, err := whisper.New(modelsPath)
	if err != nil {
		return err
	}
	defer model.Close()

	ctx, err := model.NewContext()
	if err != nil {
		return err
	}

	pcm, err := decodeMP3ToPCM(buf.Bytes())
	if err != nil {
		return err
	}

	if err := ctx.Process(pcm, nil, nil, nil); err != nil {
		return err
	}

	var transcription strings.Builder
	for {
		segment, err := ctx.NextSegment()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		transcription.WriteString(segment.Text)
	}

	fmt.Println(transcription.String())
	// if err := os.WriteFile(localPath, []byte(transcription.String()), 0644); err != nil {
	// 	return err
	// }

	return nil
}

func decodeMP3ToPCM(data []byte) ([]float32, error) {
	r := bytes.NewReader(data)

	dec, err := mp3.NewDecoder(r)
	if err != nil {
		return nil, err
	}

	decoded, err := io.ReadAll(dec)
	if err != nil {
		return nil, err
	}

	pcm := make([]int16, len(decoded)/2)
	for i := 0; i < len(pcm); i++ {
		pcm[i] = int16(binary.LittleEndian.Uint16(decoded[i*2:]))
	}

	floatSamples := make([]float32, len(pcm))
	for i, v := range pcm {
		floatSamples[i] = float32(v) / 32768.0
	}

	return floatSamples, nil
}
