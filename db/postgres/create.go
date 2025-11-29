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
		`SELECT id, video_id, link, created_at
		 FROM youtube
		 WHERE video_id=$1`, videoID).Scan(
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
		`INSERT INTO youtube (video_id, link)
		 VALUES ($1, $2)
		 RETURNING id, video_id, link, created_at`,
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

type DocumentAudio struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	FileType        string    `json:"file_type"`
	StoragePath     string    `json:"storage_path"`
	MimeType        string    `json:"mime_type"`
	FileHash        string    `json:"file_hash"`
	Size            int64     `json:"size"`
	DurationSeconds *float64  `json:"duration_seconds,omitempty"`
	PageCount       *int      `json:"page_count,omitempty"`
	CreatedAt       time.Time `json:"created_at,omitempty"`
}

func (p *PostgresDB) GetOrCreateDocument(ctx context.Context, doc DocumentAudio) (*DocumentAudio, error) {
	query := `
	INSERT INTO uploaded_files (
		file_type, original_name, storage_path, mime_type, size, file_hash, duration_seconds, page_count
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	ON CONFLICT (file_hash) DO UPDATE
	SET file_type = EXCLUDED.file_type
	RETURNING id, file_type, original_name, storage_path, mime_type, size, file_hash, duration_seconds, page_count, created_at;
	`

	row := p.Conn.QueryRow(ctx, query,
		doc.FileType,
		doc.Name,
		doc.StoragePath,
		doc.MimeType,
		doc.Size,
		doc.FileHash,
		doc.DurationSeconds,
		doc.PageCount,
	)

	var result DocumentAudio
	err := row.Scan(
		&result.ID,
		&result.FileType,
		&result.Name,
		&result.StoragePath,
		&result.MimeType,
		&result.Size,
		&result.FileHash,
		&result.DurationSeconds,
		&result.PageCount,
		&result.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
