-- ALTER TABLE uploaded_files
-- ADD CONSTRAINT duration_only_for_audio CHECK (
--     duration_seconds IS NULL OR file_type = 'audio'
-- );

-- ALTER TABLE uploaded_files
-- ADD CONSTRAINT pages_only_for_docs CHECK (
--     page_count IS NULL OR file_type = 'document'
-- );