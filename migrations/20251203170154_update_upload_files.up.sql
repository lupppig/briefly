-- the file names don't match the table because of a dumb mistake i made lmaoooooooooooooo
CREATE TABLE contents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contents TEXT,
    ai_summary TEXT,
    file_id UUID REFERENCES uploaded_files(id),
    youtube_id UUID REFERENCES youtube(id)
);