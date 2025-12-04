

CREATE TABLE uploaded_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    file_type VARCHAR(20) NOT NULL,      
    original_name TEXT NOT NULL,
    storage_path TEXT NOT NULL,
    mime_type VARCHAR(100),
    size BIGINT NOT NULL,

    file_hash TEXT NOT NULL UNIQUE,

    duration_seconds DOUBLE PRECISION,                
    page_count INT,                      

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_uploaded_files_file_hash
ON uploaded_files (file_hash);

CREATE INDEX idx_uploaded_files_file_type
ON uploaded_files (file_type);

CREATE INDEX idx_uploaded_files_original_name
ON uploaded_files (original_name);


CREATE INDEX idx_uploaded_files_created_at
ON uploaded_files (created_at DESC);