package db

import (
	"context"
	"encoding/json"
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

func (p *PostgresDB) InsertJobStatus(jobID string, status, playList_link string) error {
	_, err := p.Conn.Exec(context.Background(),
		`
	INSERT INTO youtube_jobs(job_id, link, status)	
	VALUES($1, $2, $3);
	`, jobID, playList_link, status)
	return err
}

func (p *PostgresDB) UpdateJobStatus(jobID string, status, errMsg string) error {
	_, err := p.Conn.Exec(context.Background(),
		`UPDATE youtube_jobs
         SET status = $1,
             error = $2,
             updated_at = NOW()
         WHERE job_id = $3`,
		status, errMsg, jobID,
	)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}
	return nil
}

func (p *PostgresDB) UpdateJobResult(jobID string, result string, status string) error {
	_, err := p.Conn.Exec(context.Background(),
		`UPDATE youtube_jobs
         SET result = $1,
             status = $2,
             updated_at = NOW()
         WHERE job_id = $3`,
		result, status, jobID,
	)
	if err != nil {
		return fmt.Errorf("failed to update job result: %w", err)
	}
	return nil
}

type YoutubeJob struct {
	ID           string          `json:"job_id"`
	PlaylistLink string          `json:"playlist_link"`
	Status       string          `json:"status"`
	Result       json.RawMessage `json:"result,omitempty"`
	Error        *string         `json:"error,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

func (p *PostgresDB) GetYoutubeJobByID(ctx context.Context, jobID string) (*YoutubeJob, error) {
	var job YoutubeJob
	err := p.Conn.QueryRow(ctx, `
		SELECT job_id, link, status, result, error, created_at, updated_at
		FROM youtube_jobs
		WHERE job_id = $1
	`, jobID).Scan(
		&job.ID,
		&job.PlaylistLink,
		&job.Status,
		&job.Result,
		&job.Error,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}
	return &job, nil
}
