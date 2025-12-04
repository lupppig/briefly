CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE youtube (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id TEXT NOT NULL UNIQUE,
    link TEXT NOT NULL,
    audio_path TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_video_id
ON youtube (video_id);
