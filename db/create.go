package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type Youtube struct {
	ID        string    `db:"id" json:"id"`
	VideoID   string    `db:"video_id" json:"videoID"`
	Link      string    `db:"link" json:"link"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

func (p *PostgresDB) GetOrCreateYoutube(ctx context.Context, videoID, link string) (*Youtube, error) {
	var yt Youtube

	err := p.Conn.QueryRow(ctx,
		`SELECT id, videoID, link, created_at
		 FROM youtube
		 WHERE videoID=$1`, videoID).Scan(
		&yt.ID,
		&yt.VideoID,
		&yt.Link,
		&yt.CreatedAt,
	)

	if err == nil {
		return &yt, nil
	}

	if err != pgx.ErrNoRows {
		return nil, err
	}

	err = p.Conn.QueryRow(ctx,
		`INSERT INTO youtube (videoID, link)
		 VALUES ($1, $2)
		 RETURNING id, videoID, link, created_at`,
		videoID, link).Scan(
		&yt.ID,
		&yt.VideoID,
		&yt.Link,
		&yt.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &yt, nil
}
