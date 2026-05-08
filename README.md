# Video Saver

This repository contains a small Go API that validates whether a video URL is allowed to be downloaded.

## Structure

- `main.go`: application entrypoint and server wiring
- `internal/config/`: environment config loading
- `internal/video/`: video URL validation logic
- `internal/httpapi/`: HTTP handlers and transport DTOs

## What this service does

- Reads a URL from user input or clipboard
- Validates whether the URL matches your backend rules
- Returns the direct download URL and file name when allowed

## What this service does not do

- Parse third-party share links
- Remove watermarks
- Proxy or download the media itself
- Bypass platform restrictions

## Run the server

```bash
go run .
```

The server listens on `http://127.0.0.1:3000`.

## Configuration

The service uses environment variables:

- `PORT`: listen port, defaults to `3000`
- `ALLOWED_HOSTS`: comma-separated hostname whitelist for downloadable videos

Example:

```bash
ALLOWED_HOSTS=media.example.com,cdn.example.com go run .
```

## API contract

`POST /api/resolve`

Request:

```json
{
  "url": "https://media.example.com/videos/demo.mp4"
}
```

Response:

```json
{
  "ok": true,
  "downloadUrl": "https://media.example.com/videos/demo.mp4",
  "fileName": "demo.mp4"
}
```
