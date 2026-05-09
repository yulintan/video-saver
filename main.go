package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var (
	reDouyin   = regexp.MustCompile(`https://v\.douyin\.com/[A-Za-z0-9]+/?`)
	reVideoURI = regexp.MustCompile(`"uri":"(v[^"]+)"`)
	host       = getenv("HOST", "http://localhost:8080")
)

func loadEnv(file string) {
	data, err := os.ReadFile(file)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, val := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	loadEnv(".env")

	http.HandleFunc("/", home)
	http.HandleFunc("/api/resolve", apiResolve)
	http.HandleFunc("/play/", handlePlay)

	fmt.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Douyin Downloader</title>
  <style>
    body { font-family: sans-serif; max-width: 500px; margin: 60px auto; padding: 0 20px; }
    textarea { width: 100%; padding: 10px; font-size: 14px; border: 1px solid #ddd; border-radius: 8px; resize: vertical; box-sizing: border-box; }
    button { margin-top: 10px; width: 100%; padding: 12px; background: #fe2c55; color: white; border: none; border-radius: 8px; font-size: 16px; cursor: pointer; }
    button:disabled { opacity: 0.6; }
    video { width: 100%; margin-top: 16px; border-radius: 8px; display: none; }
    #status { margin-top: 12px; padding: 10px; border-radius: 8px; display: none; }
    .ok { background: #e8f5e9; color: #2e7d32; } .err { background: #fdecea; color: #c62828; }
  </style>
</head>
<body>
  <h2>抖音下载</h2>
  <textarea id="url" rows="4" placeholder="粘贴抖音分享链接..."></textarea>
  <button id="btn" onclick="resolve()">预览</button>
  <video id="player" controls preload="metadata"></video>
  <div id="status"></div>
  <script>
    async function resolve() {
      const url = document.getElementById('url').value.trim()
      if (!url) return
      const btn = document.getElementById('btn')
      btn.disabled = true
      btn.textContent = '解析中...'
      setStatus('', '')
      try {
        const res = await fetch('/api/resolve', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ url })
        })
        const data = await res.json()
        if (data.ok) {
          const player = document.getElementById('player')
          player.src = data.videoUrl
          player.style.display = 'block'
          setStatus('ok', '✓ 解析成功 — <a href="' + data.videoUrl + '" download>点击下载</a>')
        } else {
          setStatus('err', '✗ ' + data.error)
        }
      } catch (e) {
        setStatus('err', '✗ 网络错误')
      }
      btn.disabled = false
      btn.textContent = '预览'
    }
    function setStatus(cls, msg) {
      const s = document.getElementById('status')
      s.className = cls; s.innerHTML = msg; s.style.display = msg ? 'block' : 'none'
    }
  </script>
</body>
</html>`))
}

// apiResolve extracts the video URI and returns a playable URL.
func apiResolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"ok": false, "error": "method_not_allowed"})
		return
	}

	var body struct {
		URL string `json:"url"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	match := reDouyin.FindString(body.URL)
	if match == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "no Douyin URL found"})
		return
	}

	log.Printf("resolving: %s", match)

	uri, err := extractVideoURI(match)
	if err != nil {
		log.Printf("resolve error: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
		return
	}

	videoURL := host + "/play/" + uri
	log.Printf("resolved: %s", videoURL)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "videoUrl": videoURL})
}

// handlePlay proxies the video stream from Douyin CDN.
func handlePlay(w http.ResponseWriter, r *http.Request) {
	uri := strings.TrimPrefix(r.URL.Path, "/play/")
	if uri == "" {
		http.Error(w, "missing uri", http.StatusBadRequest)
		return
	}

	upstream := "https://aweme.snssdk.com/aweme/v1/play/?video_id=" + uri + "&ratio=1080p&line=0"

	req, err := http.NewRequest("GET", upstream, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15")
	req.Header.Set("Referer", "https://www.douyin.com/")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if cl := resp.Header.Get("Content-Length"); cl != "" {
		w.Header().Set("Content-Length", cl)
	}
	io.Copy(w, resp.Body)
}

func extractVideoURI(shareURL string) (string, error) {
	req, err := http.NewRequest("GET", shareURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	match := reVideoURI.FindSubmatch(body)
	if len(match) < 2 {
		return "", fmt.Errorf("video URI not found in page")
	}
	return string(match[1]), nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
