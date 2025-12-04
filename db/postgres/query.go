package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type Youtube struct {
	ID        string    `db:"id" json:"id"`
	VideoID   string    `db:"video_id" json:"videoID"`
	Link      string    `db:"link" json:"link"`
	AudioPath string    `json:"audio_path" db:"audio_path"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

func (p *PostgresDB) GetOrCreateYoutube(ctx context.Context, videoID, audioPath, link string) (*Youtube, error) {
	var yt Youtube
	err := p.Conn.QueryRow(ctx,
		`SELECT id, video_id, link, audio_path, created_at
		 FROM youtube
		 WHERE video_id=$1`, videoID).Scan(
		&yt.ID, &yt.VideoID, &yt.Link, &yt.AudioPath, &yt.CreatedAt)
	if err == nil {
		return &yt, nil
	}
	if err != pgx.ErrNoRows {
		return nil, err
	}

	err = p.Conn.QueryRow(ctx,
		`INSERT INTO youtube (video_id, link, audio_path)
		 VALUES ($1, $2, $3)
		 RETURNING id, video_id, link, audio_path, created_at`,
		videoID, link, audioPath).Scan(
		&yt.ID, &yt.VideoID, &yt.Link, &yt.AudioPath, &yt.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &yt, nil
}

func (p *PostgresDB) UpdateYoutubeAudioPath(ctx context.Context, videoID, audioPath string) (*Youtube, error) {
	var yt Youtube
	err := p.Conn.QueryRow(ctx,
		`UPDATE youtube
		 SET audio_path = $1
		 WHERE video_id = $2
		 RETURNING id, video_id, link, audio_path, created_at`,
		audioPath, videoID).Scan(&yt.ID, &yt.VideoID, &yt.Link, &yt.AudioPath, &yt.CreatedAt)
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

type SummaryContent struct {
	Id        string  `json:"id"`
	Content   string  `json:"content"`
	AiSummary string  `json:"ai_summary"`
	FileID    *string `json:"file_id,omitempty"`
	YoutubeId *string `json:"y_id,omitempty"`
}

func (p *PostgresDB) CreateContent(ctx context.Context, content string, aiSummary string, fileID, yID *string) (*SummaryContent, error) {
	summ := &SummaryContent{
		Content:   content,
		AiSummary: aiSummary,
		FileID:    fileID,
		YoutubeId: yID,
	}

	err := p.Conn.QueryRow(ctx, `
		INSERT INTO contents(contents, ai_summary, file_id, youtube_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, contents, ai_summary, file_id, youtube_id
	`, content, aiSummary, fileID, yID).Scan(
		&summ.Id,
		&summ.Content,
		&summ.AiSummary,
		&summ.FileID,
		&summ.YoutubeId,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create content: %w", err)
	}

	return summ, nil
}

func (p *PostgresDB) GetContentByYoutubeID(ctx context.Context, ytID string) (*SummaryContent, error) {
	var c SummaryContent

	err := p.Conn.QueryRow(ctx,
		`SELECT id, contents, ai_summary, file_id, youtube_id 
		 FROM contents 
		 WHERE youtube_id = $1 
		 LIMIT 1`,
		ytID,
	).Scan(&c.Id, &c.Content, &c.AiSummary, &c.FileID, &c.YoutubeId)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &c, nil
}
