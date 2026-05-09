# video-saver

A Go service that resolves Douyin (TikTok China) share links and proxies the video stream, paired with a WeChat Mini Program frontend.

## How it works

1. User pastes a `v.douyin.com` share link into the Mini Program
2. Server fetches the page with a mobile User-Agent, extracts the video URI from the HTML
3. Returns a `/play/{uri}` proxy URL that streams the video from Douyin's CDN
4. User previews the video and downloads it to their camera roll

## Structure

```
main.go               # Go HTTP server
Dockerfile            # Two-stage build (golang:1.24-alpine → alpine)
Makefile              # build / push / deploy targets
k8s/                  # Kubernetes manifests (Deployment, Service, Ingress)
miniprogram/          # WeChat Mini Program source
.env.example          # Environment variable template
```

## Configuration

Copy `.env.example` to `.env` and fill in your values:

```
HOST=https://your-domain.com
CLUSTER=your-k8s-context
IMAGE=your-registry/video-saver
```

## Run locally

```bash
go run .
# server listens on :8080
```

## Deploy

```bash
make deploy        # build → push to registry → sync k8s secret → kubectl apply
```

Requires Docker, `kubectl` with the target context configured, and `.env` populated.

## API

`POST /api/resolve`

```json
{ "url": "https://v.douyin.com/AzJuA9Rl9eE/" }
```

```json
{ "ok": true, "videoUrl": "https://your-domain.com/play/v0300fg1..." }
```

`GET /play/{uri}` — proxies the video stream from Douyin CDN.

## WeChat Mini Program

Source is in `miniprogram/`. Open with WeChat DevTools, set your AppID in `project.config.json`.

Whitelist `https://your-domain.com` under:
- request合法域名
- downloadFile合法域名
