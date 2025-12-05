# Briefly

> ⚠️ **Disclaimer:** This project is **not production ready**. It is a personal project for learning and experimentation.

Briefly is a backend service built in **Go** for summarizing content from YouTube videos and documents (PDF, TXT). It leverages AI models for generating summaries, supports real-time polling for status updates, and integrates with cloud storage for file management.

---

## Features

1. **YouTube Summarization**

   * Supports summarizing **single YouTube videos**.
   * Uses **polling** to track video processing status and provide real-time updates.
   * Integrates with **Gemini API** for AI-based summarization of transcribed content.

2. **Document Summarization**

   * Supports **PDF and TXT documents**.
   * Extracts text content from documents using:

     * `pdftotext` for PDFs.
     * Direct reading for TXT files.
   * Generates AI summaries via Gemini API.

3. **File Management**

   * All uploaded files (audio/video/documents) are stored in **MinIO**.
   * Prevents duplicate uploads by computing **file hashes**.
   * Keeps metadata in **PostgreSQL** (file name, type, size, hash, duration/pages).

4. **Audio Processing**

   * For audio/video files, computes **duration** for metadata.
   * Audio files can also be summarized by transcribing using **Whisper.cpp**.

5. **Database & Persistence**

   * Uses **PostgreSQL** for metadata storage.
   * Checks for existing summaries before generating new ones to avoid redundancy.

6. **Polling System**

   * Tracks the **status of YouTube summarization** jobs using a polling loop.
   * Emits events for frontend to consume updates on processing progress.

7. **Security & Validation**

   * Validates uploaded files for type and size.
   * Checks for existing content to prevent unnecessary processing.

8. **AI Integration**

   * Summaries are generated using **Gemini API**.
   * Can handle both transcribed YouTube audio and document text.

9. **File Storage**

   * Files are uploaded to **MinIO**, a self-hosted S3-compatible storage.
   * Supports organizing files in `/uploads/doc` and `/uploads/audio`.

10. **Whisper Integration**

    * Supports local audio transcription using **Whisper.cpp**.
    * Requires building the library with appropriate CGO flags.

---

## Architecture & Technical Overview

* **Backend:** Go
* **Database:** PostgreSQL
* **Storage:** MinIO
* **AI Summarization:** Gemini API
* **Audio Transcription:** Whisper.cpp
* **File Processing:**

  * PDFs → `pdftotext`
  * TXT → plain text read
* **Real-time Updates:** Polling loop for YouTube job status
* **File Deduplication:** Hashing of uploaded files to prevent duplicates
* **Containerization:** Supports Docker-based local development

---

## Setup Instructions (Local Development)

1. **Clone the repository**

   ```bash
   git clone <repo-url>
   cd briefly
   ```

2. **Start local services**

   ```bash
   ./local_services.sh
   ```

   This will start:

   * PostgreSQL
   * MinIO
   * Optionally migrate DB schema (`make migrate-up`)

3. **Install PDF tools**

   ```bash
   sudo apt install poppler-utils  # provides pdftotext
   ```

4. **Whisper Setup**

   * Clone `whisper.cpp` into your project directory:

     ```bash
     git clone https://github.com/ggerganov/whisper.cpp.git
     ```
   * Build Whisper:

     ```bash
     cd whisper.cpp
     make
     ```
   * Place `models/ggml-base.en.bin` into `models/` in the project root.
   * Set environment variables for CGO:

     ```bash
     export CGO_ENABLED=1
     export LD_LIBRARY_PATH="$PWD/whisper.cpp/build/src:$PWD/whisper.cpp/build/ggml/src:$LD_LIBRARY_PATH"
     export CGO_CPPFLAGS="-I$PWD/whisper.cpp/include -I$PWD/whisper.cpp/ggml/include"
     export CGO_LDFLAGS="-L$PWD/whisper.cpp/build/src -lwhisper -L$PWD/whisper.cpp/build/ggml/src -lggml -lggml-base -lggml-cpu -lstdc++ -lm -lpthread"
     ```

5. **Run the API**

   ```bash
   go run ./cmd/main.go
   ```

---

## Usage

1. **Upload a document or audio/video**

   * PDF or TXT files → summarized text returned.
   * YouTube video → processed and summarized after transcription.

2. **Polling for YouTube**

   * Submit a video link.
   * Poll endpoint to check the job status until completion.

3. **Fetch Existing Summary**

   * Summaries are stored in the DB and fetched if they already exist.

---

## Notes

* This project is **experimental** and for **learning purposes only**.
* Not optimized for **production performance**.
* PDF extraction is text-only; images or embedded media are ignored.
* YouTube playlists are not supported, only single videos.
* Polling approach is simple and may not scale for high concurrency.
* Whisper integration requires manual build and environment setup.

---

## Future Improvements

* Add **asynchronous job queue** instead of polling.
* Support **more document types** (Word, Excel, etc.).
* Add **rate limiting and authentication**.
* Enhance **PDF extraction** to handle more complex layouts.
* Enable **multi-language support** for Whisper and Gemini summarization.
