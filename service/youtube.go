package service

import (
	"context"
	"fmt"

	"github.com/lupppig/briefly/db"
)

type Service struct {
	Db *db.PostgresDB
}

func (s *Service) YoututbeService(ctx context.Context, link, videoId string) error {

	yt, err := s.Db.GetOrCreateYoutube(ctx, videoId, link)

	if err != nil {
		return err
	}

	fmt.Printf("%+v", yt)
	return nil
}
