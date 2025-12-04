package service

import (
	"bytes"
	"fmt"
	"strings"

	whisper "github.com/ggerganov/whisper.cpp/bindings/go"
	"github.com/go-audio/wav"

	"github.com/lupppig/briefly/db/mini"
)

func (s *Service) TranscribeAudio(audioPath string) (string, error) {
	buf, err := s.Mc.GetObjectBuffer(mini.DocumentBucket, audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to read audio from MinIO: %w", err)
	}
	samples, err := wavToFloat32(buf.Bytes())
	if err != nil {
		return "", fmt.Errorf("failed to convert WAV to float32 samples: %w", err)
	}

	ctx, err := s.WhisperModel.NewContext()
	if err != nil {
		return "", fmt.Errorf("failed to create whisper context: %w", err)
	}

	ctx.SetThreads(4)
	ctx.SetLanguage("en")
	ctx.SetTranslate(false)
	ctx.SetMaxSegmentLength(0)
	ctx.SetTokenTimestamps(false)

	if err := ctx.Process(samples, nil, nil, nil); err != nil {
		return "", fmt.Errorf("whisper process failed: %w", err)
	}
	var transcript strings.Builder
	transcript.Grow(len(samples) / 100)

	for {
		seg, err := ctx.NextSegment()
		if err != nil {
			break
		}
		transcript.WriteString(seg.Text)
		transcript.WriteString(" ")
	}

	return strings.TrimSpace(transcript.String()), nil
}

func wavToFloat32(wavBytes []byte) ([]float32, error) {
	reader := wav.NewDecoder(bytes.NewReader(wavBytes))
	buf, err := reader.FullPCMBuffer()

	if err != nil {
		return nil, fmt.Errorf("failed to read PCM buffer: %w", err)
	} else if reader.SampleRate != whisper.SampleRate {
		return nil, fmt.Errorf("unsupported error rate %d", reader.SampleRate)
	} else if reader.NumChans != 1 {
		return nil, fmt.Errorf("unsupported number of channels %d", reader.NumChans)
	}

	samples := buf.AsFloat32Buffer().Data

	return samples, nil
}
