CREATE TABLE youtube_jobs (
    job_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    link TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'processing', -- processing, completed, failed
    result JSONB,                         -- store summary or audio path etc.
    error TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
