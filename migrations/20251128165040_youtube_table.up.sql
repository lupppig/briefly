CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE youtube (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    videoID TEXT NOT NULL UNIQUE,
    link TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
